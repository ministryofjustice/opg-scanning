package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/ingestion"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type IndexController struct {
	config         *config.Config
	logger         *logger.Logger
	validator      *ingestion.Validator
	httpMiddleware *httpclient.Middleware
	authMiddleware *auth.Middleware
	Queue          *ingestion.JobQueue
	AwsClient      *aws.AwsClient
}

func NewIndexController(awsClient *aws.AwsClient, appConfig *config.Config) *IndexController {
	logger := logger.NewLogger(appConfig)
	httpClient := httpclient.NewHttpClient(*appConfig, *logger)
	httpMiddleware, err := httpclient.NewMiddleware(httpClient, awsClient)
	if err != nil {
		logger.Error("Failed to create middleware: %v", nil, err)
		return nil
	}

	return &IndexController{
		config:         appConfig,
		logger:         logger,
		validator:      ingestion.NewValidator(),
		httpMiddleware: httpMiddleware,
		authMiddleware: auth.NewMiddleware(awsClient, httpMiddleware, appConfig, logger),
		Queue:          ingestion.NewJobQueue(appConfig),
		AwsClient:      awsClient,
	}
}

func (c *IndexController) HandleRequests() {
	// Create the /auth route to handle user authentication and issue JWT token
	http.Handle("/auth", http.HandlerFunc(c.AuthHandler))

	// Protect the /ingest route with JWT validation (using the authMiddleware)
	http.Handle("/ingest", telemetry.Middleware(c.logger.SlogLogger)(
		c.authMiddleware.CheckAuth(http.HandlerFunc(c.IngestHandler)),
	))

	c.logger.Info("Starting server on :" + c.config.HTTP.Port)
	http.ListenAndServe(":"+c.config.HTTP.Port, nil)
}

func (c *IndexController) AuthHandler(w http.ResponseWriter, r *http.Request) {
	// This handler will pass on to the Authenticate middleware that will validate the user
	// credentials and issue a token.
	c.authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This function will only be called if authentication is successful
		w.Write([]byte("Authentication successful"))
	})).ServeHTTP(w, r)
}

func (c *IndexController) IngestHandler(w http.ResponseWriter, r *http.Request) {
	// Extract claims from context
	_, ok := r.Context().Value("claims").(jwt.MapClaims)
	if !ok {
		http.Error(w, "Unauthorized: Unable to extract claims", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		c.respondWithError(w, http.StatusMethodNotAllowed, "Invalid HTTP method", nil)
		return
	}

	c.logger.Info("Received ingestion request", nil)

	// Step 1: Read request body
	bodyStr, err := c.readRequestBody(r)
	if err != nil {
		c.logger.Error("Failed to read request body: " + err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Step 2: Determine content type and validate XML
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/xml" && contentType != "text/xml" {
		c.logger.Error("Unsupported Content-Type: " + contentType)
		http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
		return
	}

	parsedBaseXml, err := c.validateAndSanitizeXML(bodyStr)
	if err != nil {
		c.logger.Error("XML validation failed: " + err.Error())
		http.Error(w, "Invalid XML data", http.StatusBadRequest)
		return
	}

	// Step 3: Validate the parsed set
	if err := c.validator.ValidateSet(parsedBaseXml); err != nil {
		c.logger.Error("Document validation failed: " + err.Error())
		http.Error(w, "Invalid document data", http.StatusBadRequest)
		return
	}

<<<<<<< Updated upstream
	// Step 4: Create a case stub in Sirius if we have a case to create
	scannedCaseResponse, err := CreateStubCase(c.config.App.SiriusBaseURL, *parsedBaseXml)
=======
	// Sirius API integration
	// Create a case stub in Sirius if we have a case to create
>>>>>>> Stashed changes
	if err != nil {
		c.logger.Error("Failed to create case stub in Sirius: " + err.Error())
		http.Error(w, "Failed to create case stub in Sirius", http.StatusInternalServerError)
		return
	}

<<<<<<< Updated upstream
	// Step 5: Queue each document for further processing
	c.logger.Info("Queueing documents for processing")
=======
	// Create a new client and prepare to attach documents
	client := NewClient(c.httpMiddleware)
	service := NewService(client, parsedBaseXml)
	scannedCaseResponse, err := service.CreateCaseStub(r.Context())
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
	c.logger.Info("Queueing documents for processing", nil)
>>>>>>> Stashed changes
	for i := range parsedBaseXml.Body.Documents {
		doc := &parsedBaseXml.Body.Documents[i]
		c.Queue.AddToQueue(doc, "xml", func(processedDocument interface{}) {
			c.logger.Info(fmt.Sprintf("%v", processedDocument))
			c.logger.Info("Job processing completed for document")
		})
		c.logger.Info("Job added to queue for document")
	}

	// Step 6: Send the UUID response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(scannedCaseResponse)
	c.logger.Info("Ingestion request processed successfully, UID: " + scannedCaseResponse.UID)

}

func (c *IndexController) CloseQueue() {
	c.Queue.Close()
}

// Helper Method: Read Request Body
func (c *IndexController) readRequestBody(r *http.Request) (string, error) {
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
	c.logger.Info("Validating against XSD")
	xsdValidator, err := ingestion.NewXSDValidator(c.config.App.ProjectFullPath+"/xsd/"+schemaLocation, bodyStr)
	if err != nil {
		return nil, err
	}
	if err := xsdValidator.ValidateXsd(); err != nil {
		return nil, fmt.Errorf("XSD validation failed: %w", err)
	}

	// Validate and sanitize the XML
	c.logger.Info("Validating and sanitizing XML")
	xmlValidator := ingestion.NewXmlValidator(*c.config)
	parsedBaseXml, err := xmlValidator.XmlValidateSanitize(bodyStr)
	if err != nil {
		return nil, err
	}

	return parsedBaseXml, nil
}
