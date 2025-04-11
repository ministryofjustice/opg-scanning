package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/lestrrat-go/libxml2/xsd"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/ingestion"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
	logger := logger.GetLogger(appConfig)

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

		if _, err := w.Write([]byte("OK")); err != nil {
			c.logger.Error(err.Error(), nil)
		}
	}))

	// Create the route to handle user authentication and issue JWT token
	http.Handle("/auth/sessions", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.AuthHandler(w, r)
	}))

	// Protect the route with JWT validation (using the authMiddleware)
	http.Handle("/api/ddc", otelhttp.NewHandler(logger.LoggingMiddleware(c.logger.SlogLogger)(
		c.authMiddleware.CheckAuthMiddleware(http.HandlerFunc(c.IngestHandler)),
	), "scanning"))

	c.logger.Info("Starting server on :"+c.config.HTTP.Port, nil)

	server := &http.Server{
		Addr:              ":" + c.config.HTTP.Port,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		c.logger.Error(err.Error(), nil)
	}
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
		c.authResponse(ctx, w, ErrorResponse{Error: errMsg})
		return
	}

	// Retrieve user from context
	userFromCtx, ok := auth.UserFromContext(ctx)
	if !ok {
		errMsg := "Failed to retrieve user from context"
		c.logger.Error(errMsg, nil)
		c.authResponse(ctx, w, ErrorResponse{Error: errMsg})
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

	c.authResponse(ctx, w, resp)
}

func (c *IndexController) authResponse(ctx context.Context, w http.ResponseWriter, resp any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.respondWithError(ctx, w, http.StatusInternalServerError, "Failed to encode response", err)
	}
}

func (c *IndexController) IngestHandler(w http.ResponseWriter, r *http.Request) {
	reqCtx := r.Context()

	if r.Method != http.MethodPost {
		c.respondWithError(reqCtx, w, http.StatusMethodNotAllowed, "Invalid HTTP method", nil)
		return
	}

	c.logger.InfoWithContext(reqCtx, "Received ingestion request", nil)

	bodyStr, err := c.readRequestBody(r)
	if err != nil {
		c.respondWithError(reqCtx, w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/xml") && !strings.HasPrefix(contentType, "text/xml") {
		c.respondWithError(reqCtx, w, http.StatusBadRequest, "Invalid content type", fmt.Errorf("expected application/xml or text/xml, got %s", contentType))
		return
	}

	// Save Set to S3
	filename, err := c.AwsClient.PersistSetData(reqCtx, []byte(bodyStr))
	if err != nil {
		c.respondWithError(reqCtx, w, http.StatusInternalServerError, "Could not persist set to S3", err)
		return
	}

	c.logger.InfoWithContext(reqCtx, "Stored Set data", map[string]any{
		"set_filename": filename,
	})

	parsedBaseXml, err := c.validateAndSanitizeXML(reqCtx, bodyStr)
	if err != nil {
		c.respondWithError(reqCtx, w, http.StatusBadRequest, "Validate and sanitize XML failed", err)
		return
	}

	// Validate the parsed set
	if err := c.validator.ValidateSet(parsedBaseXml); err != nil {
		c.respondWithError(reqCtx, w, http.StatusBadRequest, "Validate set failed", err)
		return
	}

	client := NewClient(c.httpMiddleware)
	service := NewService(client, parsedBaseXml)
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(c.config.HTTP.Timeout)*time.Second)
	defer cancel()
	scannedCaseResponse, err := service.CreateCaseStub(ctx)
	if err != nil {
		c.respondWithError(reqCtx, w, http.StatusInternalServerError, "Failed to create case stub in Sirius", err)
		return
	}

	if scannedCaseResponse == nil || scannedCaseResponse.UID == "" {
		c.respondWithError(reqCtx, w, http.StatusInternalServerError,
			"Invalid response from Sirius when creating case stub, scannedCaseResponse is nil or missing UID",
			errors.New("scannedCaseResponse UID missing"))
		return
	}

	// Processing qeueue
	if err := c.ProcessQueue(reqCtx, scannedCaseResponse, parsedBaseXml); err != nil {
		c.respondWithError(reqCtx, w, http.StatusInternalServerError, err.Error(), err)
		return
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
		c.respondWithError(reqCtx, w, http.StatusInternalServerError, "Failed to encode response", err)
	} else {
		c.logger.InfoWithContext(reqCtx, "Ingestion request processed successfully", map[string]any{
			"uid": scannedCaseResponse.UID,
		})
	}
}

func (c *IndexController) respondWithError(ctx context.Context, w http.ResponseWriter, statusCode int, message string, err error) {
	if statusCode >= 500 {
		c.logger.ErrorWithContext(ctx, message, map[string]any{
			"error": err,
		})
	} else {
		c.logger.InfoWithContext(ctx, message, map[string]any{
			"error": err,
		})
	}

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
		c.respondWithError(ctx, w, http.StatusInternalServerError, "Failed to encode response", err)
	}
}

// Helper Method: Validate and Sanitize XML
func (c *IndexController) readRequestBody(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", errors.New("request body is empty")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	defer r.Body.Close() //nolint:errcheck // no need to check error when closing body
	return string(body), nil
}

func (c *IndexController) validateAndSanitizeXML(ctx context.Context, bodyStr string) (*types.BaseSet, error) {
	schemaLocation, err := ingestion.ExtractSchemaLocation(bodyStr)
	if err != nil {
		return nil, err
	}

	// Validate against XSD
	c.logger.InfoWithContext(ctx, "Validating against XSD", nil)
	xsdValidator, err := ingestion.NewXSDValidator(c.config.App.ProjectFullPath+"/xsd/"+schemaLocation, bodyStr)
	if err != nil {
		return nil, err
	}
	if err := xsdValidator.ValidateXsd(); err != nil {
		if schemaValidationError, ok := err.(xsd.SchemaValidationError); ok {
			var validationErrors []string
			for _, error := range schemaValidationError.Errors() {
				validationErrors = append(validationErrors, error.Error())
			}
			return nil, Problem{
				Title:            "Validate and sanitize XML failed",
				ValidationErrors: validationErrors,
			}
		}
		return nil, fmt.Errorf("set failed XSD validation: %w", err)
	}

	// Validate and sanitize the XML
	c.logger.InfoWithContext(ctx, "Validating XML", nil)
	xmlValidator := ingestion.NewXmlValidator(*c.config)
	parsedBaseXml, err := xmlValidator.XmlValidate(bodyStr)
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
	if !slices.Contains(constants.SupportedDocumentTypes, document.Type) {
		return Problem{
			Title: fmt.Sprintf("Document type %s is not supported", document.Type),
		}
	}

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

func (c *IndexController) processAndPersist(ctx context.Context, decodedXML []byte, originalDoc *types.BaseDocument) (string, error) {
	xmlReader := bytes.NewReader(decodedXML)
	fileName, awsErr := c.AwsClient.PersistFormData(ctx, xmlReader, originalDoc.Type)
	if awsErr != nil {
		return "", awsErr
	}
	return fileName, nil
}
