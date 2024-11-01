package api

import (
	"io"
	"net/http"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/ingestion"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type IndexController struct {
	config    *config.Config
	validator *ingestion.Validator
	Queue     *ingestion.JobQueue
	logger    *logger.Logger
}

func NewIndexController() *IndexController {
	return &IndexController{
		config:    config.GetConfig(),
		validator: ingestion.NewValidator(),
		Queue:     ingestion.NewJobQueue(),
		logger:    logger.NewLogger(),
	}
}

func (c *IndexController) HandleRequests() {
	http.HandleFunc("/ingest", c.IngestHandler)
	c.logger.Info("Starting server on :8080")
	http.ListenAndServe(":8080", nil)
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

	docType, parsedBaseXml, err := c.validateAndSanitizeXML(bodyStr)
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

	// Step 4: Queue each document for further processing
	c.logger.Info("Queueing documents for processing")
	for i := range parsedBaseXml.Body.Documents {
		doc := &parsedBaseXml.Body.Documents[i]
		c.Queue.AddToQueue(doc, docType, "xml", func() {
			c.logger.Info("Job processing completed for document")
		})
		c.logger.Info("Job added to queue for document")
	}

	w.WriteHeader(http.StatusAccepted)
	c.logger.Info("Ingestion request processed successfully")
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
func (c *IndexController) validateAndSanitizeXML(bodyStr string) (string, *types.Set, error) {
	// Extract the document type from the XML
	docType, err := ingestion.ExtractDocType(bodyStr)
	if err != nil {
		return "", nil, err
	}

	// Validate against XSD
	xsdValidator, err := ingestion.NewXSDValidator(c.config.XSDDirectory+"/"+docType+".xsd", bodyStr)
	if err != nil {
		return "", nil, err
	}
	if err := xsdValidator.ValidateXsd(); err != nil {
		return "", nil, err
	}

	// Validate and sanitize the XML
	xmlValidator := ingestion.NewXmlValidator(*c.config)
	parsedBaseXml, err := xmlValidator.XmlValidateSanitize(bodyStr)
	if err != nil {
		return "", nil, err
	}

	return docType, parsedBaseXml, nil
}
