package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
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
	http.Handle("/api/ddc", logger.LoggingMiddleware(c.logger.SlogLogger)(
		c.authMiddleware.CheckAuthMiddleware(http.HandlerFunc(c.IngestHandler)),
	))

	c.logger.Info("Starting server on :"+c.config.HTTP.Port, nil)
	http.ListenAndServe(":"+c.config.HTTP.Port, nil)
}

func (c *IndexController) AuthHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate user credentials and issue JWT token
	_, err := c.authMiddleware.Authenticator.Authenticate(w, r)
	if err != nil {
		c.respondWithError(w, http.StatusUnauthorized, "Authentication failed", err)
		return
	}

	w.Write([]byte("Authentication successful"))
}

func (c *IndexController) IngestHandler(w http.ResponseWriter, r *http.Request) {
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

	reqCtx := r.Context()

	for i := range parsedBaseXml.Body.Documents {
		doc := &parsedBaseXml.Body.Documents[i]
		// r.Context() carries the enriched logger injected by the middleware.
    	c.Queue.AddToQueue(reqCtx, doc, "xml", func(processedDoc interface{}, originalDoc *types.BaseDocument) {
			// Extract the enriched logger from the original request context.
			enrichedLogger := logger.LoggerFromContext(reqCtx)
			// Create a new context starting with context.Background() and inject the enriched logger.
			loggerCtx := logger.ContextWithLogger(context.Background(), enrichedLogger)
			// Then apply your timeout on the new context.
			ctx, cancel := context.WithTimeout(loggerCtx, time.Duration(c.config.HTTP.Timeout)*time.Second)
			defer cancel()

			// Attach documents to case
			// Set the documents original and processed entities before attaching
			service.originalDoc = originalDoc
			attchResp, decodedXML, docErr := service.AttachDocuments(ctx, scannedCaseResponse)
			if docErr != nil {
				c.logger.ErrorWithContext(ctx, "Failed to attach document", map[string]interface{}{
					"set_uid":       scannedCaseResponse.UID,
					"document_type": originalDoc.Type,
					"error":         docErr.Error(),
				})
				return
			}

			// Persist form data in S3 bucket
			fileName, persistErr := c.processAndPersist(ctx, decodedXML, originalDoc)
			if persistErr != nil {
				c.logger.ErrorWithContext(ctx, "Failed to persist document", map[string]interface{}{
					"set_uid":       scannedCaseResponse.UID,
					"document_type": originalDoc.Type,
					"error":         persistErr.Error(),
				})
				return
			}

			// Check if the document is a correspondence type; if so do not send to the job queue
			if util.Contains([]string{"Correspondence", "SupCorrespondence"}, originalDoc.Type) {
				c.logger.InfoWithContext(ctx, "Skipping external job processing, checks completed for document", map[string]interface{}{
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
				c.logger.ErrorWithContext(ctx, "Failed to create AWS queue", nil, err)
			}
			messageID, err := AwsQueue.QueueSetForProcessing(ctx, scannedCaseResponse, fileName)
			if err != nil {
				c.logger.ErrorWithContext(ctx, "Failed to queue document for processing", map[string]interface{}{
					"set_uid":       scannedCaseResponse.UID,
					"document_type": originalDoc.Type,
					"error":         err.Error(),
				})
				return
			}

			c.logger.InfoWithContext(ctx, "Job processing completed for document", map[string]interface{}{
				"set_uid":       scannedCaseResponse.UID,
				"pdf_uuid":      attchResp.UUID,
				"job_queue_id":  messageID,
				"filename":      fileName,
				"document_type": originalDoc.Type,
			})

		})
		c.logger.InfoWithContext(ctx, "Document queued for processing", map[string]interface{}{
			"set_uid":       scannedCaseResponse.UID,
			"document_type": doc.Type,
		})
	}

	// Send the UID response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(scannedCaseResponse); err != nil {
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
	http.Error(w, message, statusCode)
}

// Helper Method: Read Request Body
func (c *IndexController) readRequestBody(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", errors.New("request body is empty")
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

	return parsedBaseXml, nil
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
