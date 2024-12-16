package api

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

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
		logger:    logger.NewLogger(appConfig),
		validator: ingestion.NewValidator(),
		Queue:     ingestion.NewJobQueue(appConfig),
		AwsClient: awsClient,
	}
}

func (c *IndexController) HandleRequests() {
	http.Handle("/ingest", telemetry.Middleware(c.logger.SlogLogger)(http.HandlerFunc(c.IngestHandler)))

	c.logger.Info("Starting server on :"+c.config.HTTP.Port, nil)
	http.ListenAndServe(":"+c.config.HTTP.Port, nil)
}

func (c *IndexController) IngestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.respondWithError(w, http.StatusMethodNotAllowed, "Invalid HTTP method", nil)
		return
	}

	c.logger.Info("Received ingestion request", nil)

	bodyStr, err := c.readRequestBody(r)
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Failed to read request body", err)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/xml" && contentType != "text/xml" {
		c.respondWithError(w, http.StatusBadRequest, "Invalid content type", fmt.Errorf("expected application/xml or text/xml, got %s", contentType))
		return
	}

	parsedBaseXml, err := c.validateAndSanitizeXML(bodyStr)
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid XML data", err)
		return
	}

	// Validate the parsed set
	if err := c.validator.ValidateSet(parsedBaseXml); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid XML data", err)
		return
	}

	// Sirius API integration
	// Create a case stub in Sirius if we have a case to create
	httpClient := httpclient.NewHttpClient(*c.config, *c.logger)
	middleware, err := httpclient.NewMiddleware(httpClient, c.AwsClient)
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, "Failed to create middleware", err)
		return
	}

	// Create a new client and prepare to attach documents
	client := NewClient(middleware)
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
	for i := range parsedBaseXml.Body.Documents {
		doc := &parsedBaseXml.Body.Documents[i]
		c.Queue.AddToQueue(doc, "xml", func(processedDoc interface{}, originalDoc *types.BaseDocument) {
			// Create a new context
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.config.HTTP.Timeout)*time.Second)
			defer cancel()

			// Attach documents to case
			// Set the documents original and processed entities before attaching
			service.originalDoc = originalDoc
			attchResp, docErr := service.AttachDocuments(ctx, scannedCaseResponse)
			if docErr != nil {
				c.logger.Error("Failed to attach document", map[string]interface{}{
					"Set UID":       scannedCaseResponse.UID,
					"Document type": originalDoc.Type,
					"Error":         docErr.Error(),
				})
				return
			}

			// Persist form data
			persistMsg, persistErr := c.processAndPersist(ctx, processedDoc, originalDoc)
			if persistErr != nil {
				c.logger.Error(persistMsg, map[string]interface{}{
					"Set UID":       scannedCaseResponse.UID,
					"Document type": originalDoc.Type,
					"Error":         persistErr.Error(),
				})
				return
			}

			c.logger.Info("Job processing completed for document. %s", map[string]interface{}{
				"Set UID":       scannedCaseResponse.UID,
				"PDF UUID":      attchResp.UUID,
				"Document type": originalDoc.Type,
			}, persistMsg)

		})
		c.logger.Info("Document queued for processing", map[string]interface{}{
			"Set UID":       scannedCaseResponse.UID,
			"Document type": doc.Type,
		})
	}

	// Send the UUID response
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

	return parsedBaseXml, nil
}

func (c *IndexController) processAndPersist(ctx context.Context, processedDoc interface{}, originalDoc *types.BaseDocument) (message string, err error) {
	// Convert processedDoc to XML
	xmlBytes, err := xml.MarshalIndent(processedDoc, "", "  ")
	if err != nil {
		return "Failed to marshal processedDoc to XML", err
	}

	// Persist the XML
	xmlReader := bytes.NewReader(xmlBytes)
	awsErr := c.AwsClient.PersistFormData(ctx, xmlReader, originalDoc.Type)
	if awsErr != nil {
		return "Failed to persist form data", awsErr
	}

	return "Form data persisted successfully", nil
}
