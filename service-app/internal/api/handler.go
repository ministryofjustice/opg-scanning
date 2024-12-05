package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/ingestion"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type IndexController struct {
	config    *config.Config
	logger    *logger.Logger
	validator *ingestion.Validator
	Queue     *ingestion.JobQueue
	AwsClient *aws.AwsClient
}

func NewIndexController(awsClient *aws.AwsClient, appConfig *config.Config) *IndexController {
	return &IndexController{
		config:    appConfig,
		logger:    logger.NewLogger(),
		validator: ingestion.NewValidator(),
		Queue:     ingestion.NewJobQueue(),
		AwsClient: awsClient,
	}
}

func (c *IndexController) HandleRequests() {
	http.Handle("/ingest", telemetry.Middleware(c.logger.SlogLogger)(http.HandlerFunc(c.IngestHandler)))

	c.logger.Info("Starting server on :" + c.config.HTTP.Port)
	http.ListenAndServe(":"+c.config.HTTP.Port, nil)
}

func (c *IndexController) IngestHandler(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Received ingestion request")

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

	// Step 4: Sirius API integration
	// Step 4.1: Create a case stub in Sirius if we have a case to create
	httpClient := httpclient.NewHttpClient(*c.config, *c.logger)
	middleware, err := httpclient.NewMiddleware(httpClient, c.AwsClient)
	if err != nil {
		c.logger.Error("Failed to create middleware: " + err.Error())
		http.Error(w, "Failed to create middleware", http.StatusInternalServerError)
		return
	}

	// Step 5: Create a new client and prepare to attach documents
	client := NewClient(middleware)
	service := NewService(client, parsedBaseXml)
	scannedCaseResponse, err := service.CreateCaseStub(r.Context())
	if err != nil {
		c.logger.Error("Failed to create case stub in Sirius: " + err.Error())
		http.Error(w, "Failed to create case stub in Sirius", http.StatusInternalServerError)
		return
	}
	// Step 5.1: Queue each document for further processing
	c.logger.Info("Queueing documents for processing")
	for i := range parsedBaseXml.Body.Documents {
		doc := &parsedBaseXml.Body.Documents[i]
		c.Queue.AddToQueue(doc, "xml", func(processedDoc interface{}, originalDoc *types.BaseDocument) {
			c.logger.Info(fmt.Sprintf("Job processing completed for document type: %v", originalDoc.Type))

			// Step 5.2	: Attach documents to case
			// Set the documents original and processed before processing
			service.processedDoc = processedDoc
			service.originalDoc = originalDoc
			_, err = service.AttachDocuments(r.Context(), scannedCaseResponse)
			if err != nil {
				c.logger.Error("Failed to attach documents to case stub: " + err.Error())
				http.Error(w, "Failed to attach documents to case stub", http.StatusInternalServerError)
				return
			}

			// Log the scanned document resposne.
			c.logger.Info(fmt.Sprintf("%v", scannedCaseResponse))
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
