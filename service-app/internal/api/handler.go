package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/lestrrat-go/libxml2/xsd"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/ingestion"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type IndexController struct {
	config         *config.Config
	logger         *logger.Logger
	validator      *ingestion.Validator
	httpMiddleware *httpclient.Middleware
	authMiddleware *auth.Middleware
	Queue          *ingestion.JobQueue
	AwsClient      aws.AwsClientInterface
}

type response struct {
	Data responseData `json:"data"`
}

type responseData struct {
	Success          bool     `json:"success"`
	Message          string   `json:"message"`
	Uid              string   `json:"uid,omitempty"`
	ValidationErrors []string `json:"validationErrors,omitempty"`
}

var uidReplacementRegex = regexp.MustCompile(`^7[0-9]{3}-[0-9]{4}-[0-9]{4}$`)

func NewIndexController(awsClient aws.AwsClientInterface, appConfig *config.Config) *IndexController {
	logger := logger.NewLogger(appConfig)

	// Create dependencies
	httpClient := httpclient.NewHttpClient(*appConfig, *logger)
	tokenGenerator := auth.NewJWTTokenGenerator(awsClient, appConfig, logger)
	cookieHelper := auth.MembraneCookieHelper{
		CookieName: "membrane",
		Secure:     appConfig.App.Environment != "local",
	}
	authenticator := auth.NewBasicAuthAuthenticator(awsClient, cookieHelper, tokenGenerator)

	// Create authentication middleware
	authMiddleware := auth.NewMiddleware(authenticator, tokenGenerator, cookieHelper, logger)
	// Create HTTP middleware
	httpMiddleware, _ := httpclient.NewMiddleware(httpClient, tokenGenerator)

	return &IndexController{
		config:         appConfig,
		logger:         logger,
		validator:      ingestion.NewValidator(),
		httpMiddleware: httpMiddleware,
		authMiddleware: authMiddleware,
		Queue:          ingestion.NewJobQueue(appConfig),
		AwsClient:      awsClient,
	}
}

func (c *IndexController) HandleRequests() {
	http.Handle("/health-check", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Create the route to handle user authentication and issue JWT token
	http.Handle("/auth/sessions", http.HandlerFunc(c.AuthHandler))

	// Protect the route with JWT validation (using the authMiddleware)
	http.Handle("/api/ddc", telemetry.Middleware(c.logger.SlogLogger)(
		c.authMiddleware.CheckAuthMiddleware(http.HandlerFunc(c.IngestHandler)),
	))

	c.logger.Info("Starting server on :"+c.config.HTTP.Port, nil)
	http.ListenAndServe(":"+c.config.HTTP.Port, nil)
}

func (c *IndexController) AuthHandler(w http.ResponseWriter, r *http.Request) {
	// Define response error struct
	type ErrorResponse struct {
		Error string `json:"error"`
	}

	// Authenticate user credentials and issue JWT token
	ctx, err := c.authMiddleware.Authenticator.Authenticate(w, r)
	if err != nil {
		errMsg := fmt.Sprintf("Authentication failed: %v", err)
		c.logger.Error(errMsg, nil)
		c.authResponse(w, ErrorResponse{Error: errMsg})
		return
	}

	// Retrieve user from context
	userFromCtx, ok := auth.UserFromContext(ctx)
	if !ok {
		errMsg := "Failed to retrieve user from context"
		c.logger.Error(errMsg, nil)
		c.authResponse(w, ErrorResponse{Error: errMsg})
		return
	}

	token := c.authMiddleware.TokenGenerator.GetToken()

	// Build response with email and token
	resp := struct {
		Email string `json:"email"`
		Token string `json:"authentication_token"`
	}{
		Email: userFromCtx.Email,
		Token: token,
	}

	c.authResponse(w, resp)
}

func (c *IndexController) authResponse(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
	}
}

func (c *IndexController) IngestHandler(w http.ResponseWriter, r *http.Request) {
	reqID, _ := r.Context().Value(constants.TraceIDKey).(string)

	if r.Method != http.MethodPost {
		c.respondWithError(w, http.StatusMethodNotAllowed, "Invalid HTTP method", nil)
		return
	}

	c.logger.Info("Received ingestion request", nil)

	bodyStr, err := c.readRequestBody(r)
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/xml" && contentType != "text/xml" {
		c.respondWithError(w, http.StatusBadRequest, "Invalid content type", fmt.Errorf("expected application/xml or text/xml, got %s", contentType))
		return
	}

	parsedBaseXml, err := c.validateAndSanitizeXML(bodyStr)
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Validate and sanitize XML failed", err)
		return
	}

	// Validate the parsed set
	if err := c.validator.ValidateSet(parsedBaseXml); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Validate set failed", err)
		return
	}

	// Create a new client and prepare to attach documents
	client := NewClient(c.httpMiddleware)
	service := NewService(client, parsedBaseXml)
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(c.config.HTTP.Timeout)*time.Second)
	defer cancel()
	scannedCaseResponse, err := service.CreateCaseStub(ctx)
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, "Failed to create case stub in Sirius", err)
		return
	}

	// Ensure scannedCaseResponse
	if scannedCaseResponse == nil || scannedCaseResponse.UID == "" {
		c.respondWithError(w, http.StatusInternalServerError,
			"Invalid response from Sirius when creating case stub, scannedCaseResponse is nil or missing UID",
			fmt.Errorf("scannedCaseResponse UID missing"))
		return
	}
	// Queue each document for further processing
	c.logger.Info("Queueing documents for processing", map[string]interface{}{
		"Header": parsedBaseXml.Header,
	})

	for i := range parsedBaseXml.Body.Documents {
		doc := &parsedBaseXml.Body.Documents[i]
		c.Queue.AddToQueue(ctx, doc, "xml", func(processedDoc interface{}, originalDoc *types.BaseDocument) {
			// Create a new context
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.config.HTTP.Timeout)*time.Second)
			defer cancel()

			// Attach documents to case
			// Set the documents original and processed entities before attaching
			service.originalDoc = originalDoc
			attchResp, decodedXML, docErr := service.AttachDocuments(ctx, scannedCaseResponse)
			if docErr != nil {
				c.logger.Error("Failed to attach document", map[string]interface{}{
					"trace_id":      reqID,
					"set_uid":       scannedCaseResponse.UID,
					"document_type": originalDoc.Type,
					"error":         docErr.Error(),
				})
				return
			}

			// Persist form data in S3 bucket
			fileName, persistErr := c.processAndPersist(ctx, decodedXML, originalDoc)
			if persistErr != nil {
				c.logger.Error("Failed to persist document", map[string]interface{}{
					"trace_id":      reqID,
					"set_uid":       scannedCaseResponse.UID,
					"document_type": originalDoc.Type,
					"error":         persistErr.Error(),
				})
				return
			}

			// Check if the document is a correspondence type; if so do not send to the job queue
			if util.Contains([]string{"Correspondence", "SupCorrespondence"}, originalDoc.Type) {
				c.logger.Info("Skipping external job processing, checks completed for document", map[string]interface{}{
					"trace_id":      reqID,
					"set_uid":       scannedCaseResponse.UID,
					"pdf_uuid":      attchResp.UUID,
					"filename":      fileName,
					"document_type": originalDoc.Type,
				})
				return
			}

			// Persist external aws job queue with UID+fileName
			AwsQueue, err := aws.NewAwsQueue(c.config)
			if err != nil {
				c.logger.Error("Failed to create AWS queue", nil, err)
			}
			messageID, err := AwsQueue.QueueSetForProcessing(ctx, scannedCaseResponse, fileName)
			if err != nil {
				c.logger.Error("Failed to queue document for processing", map[string]interface{}{
					"trace_id":      reqID,
					"set_uid":       scannedCaseResponse.UID,
					"document_type": originalDoc.Type,
					"error":         err.Error(),
				})
				return
			}

			c.logger.Info("Job processing completed for document", map[string]interface{}{
				"trace_id":      reqID,
				"set_uid":       scannedCaseResponse.UID,
				"pdf_uuid":      attchResp.UUID,
				"job_queue_id":  messageID,
				"filename":      fileName,
				"document_type": originalDoc.Type,
			})

		})
		c.logger.Info("Document queued for processing", map[string]interface{}{
			"trace_id":      reqID,
			"set_uid":       scannedCaseResponse.UID,
			"document_type": doc.Type,
		})
	}

	// Send the UID response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	uid := scannedCaseResponse.UID
	if uidReplacementRegex.MatchString(uid) {
		uid = strings.ReplaceAll(uid, "-", "")
	}

	resp := response{
		Data: responseData{
			Success: true,
			Message: fmt.Sprintf("The document set for case %s has been queued for processing", uid),
			Uid:     uid,
		},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
	} else {
		c.logger.Info("Ingestion request processed successfully, UID: %s", nil, scannedCaseResponse.UID)
	}
}

func (c *IndexController) CloseQueue() {
	c.Queue.Close()
}

func (c *IndexController) respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	c.logger.Error("%s: %v", nil, message, err)

	resp := response{
		Data: responseData{
			Success: false,
			Message: message,
		},
	}

	if problem, ok := err.(Problem); ok {
		resp.Data.Message = problem.Title
		resp.Data.ValidationErrors = problem.ValidationErrors
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.respondWithError(w, http.StatusInternalServerError, "Failed to encode response", err)
	}
}

// Helper Method: Read Request Body
func (c *IndexController) readRequestBody(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", fmt.Errorf("request body is empty")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	return string(body), nil
}

// Helper Method: Validate and Sanitize XML
func (c *IndexController) validateAndSanitizeXML(bodyStr string) (*types.BaseSet, error) {
	// Extract the document type from the XML
	schemaLocation, err := ingestion.ExtractSchemaLocation(bodyStr)
	if err != nil {
		return nil, err
	}

	// Validate against XSD
	c.logger.Info("Validating against XSD", nil)
	xsdValidator, err := ingestion.NewXSDValidator(c.config.App.ProjectFullPath+"/xsd/"+schemaLocation, bodyStr)
	if err != nil {
		return nil, err
	}
	if err := xsdValidator.ValidateXsd(); err != nil {
		return nil, fmt.Errorf("XSD validation failed: %w", err)
	}

	// Validate and sanitize the XML
	c.logger.Info("Validating and sanitizing XML", nil)
	xmlValidator := ingestion.NewXmlValidator(*c.config)
	parsedBaseXml, err := xmlValidator.XmlValidateSanitize(bodyStr)
	if err != nil {
		return nil, err
	}

	// Validate embedded documents
	for _, document := range parsedBaseXml.Body.Documents {
		if err := c.validateDocument(document); err != nil {
			return nil, err
		}
	}

	return parsedBaseXml, nil
}

func (c *IndexController) validateDocument(document types.BaseDocument) error {
	decodedXML, err := util.DecodeEmbeddedXML(document.EmbeddedXML)
	if err != nil {
		return fmt.Errorf("failed to decode XML data from %s: %w", document.Type, err)
	}

	schemaLocation, err := ingestion.ExtractSchemaLocation(string(decodedXML))
	if err != nil {
		return fmt.Errorf("failed to extract schema from %s: %w", document.Type, err)
	}

	xsdValidator, err := ingestion.NewXSDValidator(c.config.App.ProjectFullPath+"/xsd/"+schemaLocation, string(decodedXML))
	if err != nil {
		return fmt.Errorf("failed to load schema %s: %w", schemaLocation, err)
	}

	if err := xsdValidator.ValidateXsd(); err != nil {
		if schemaValidationError, ok := err.(xsd.SchemaValidationError); ok {
			var validationErrors []string
			for _, error := range schemaValidationError.Errors() {
				validationErrors = append(validationErrors, error.Error())
			}

			return Problem{
				Title:            fmt.Sprintf("XML for %s failed XSD validation", document.Type),
				ValidationErrors: validationErrors,
			}
		}

		return fmt.Errorf("failed XSD validation: %w", err)
	}

	return nil
}

func (c *IndexController) processAndPersist(ctx context.Context, decodedXML []byte, originalDoc *types.BaseDocument) (fileName string, err error) {
	// Persist the decoded XML data in its original form as represented in the origin request.
	xmlReader := bytes.NewReader(decodedXML)
	fileName, awsErr := c.AwsClient.PersistFormData(ctx, xmlReader, originalDoc.Type)
	if awsErr != nil {
		return "", awsErr
	}

	return fileName, nil
}
