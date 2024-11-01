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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		c.logger.Error("Failed to read request body: " + err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var format string
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/xml" || contentType == "text/xml" {
		format = "xml"
	}
	docType := "LP1F" // TODO: Determine the document type dynamically if possible

	c.logger.Info("Begin validation of XML document")
	var parsedBaseXml *types.Set
	if format == "xml" {
		xmlValidator := ingestion.NewXmlValidator(*c.config)
		parsedBaseXml, err = xmlValidator.XmlValidate(string(body))
		if err != nil {
			c.logger.Error("XML validation failed: " + err.Error())
			http.Error(w, "Invalid XML data", http.StatusBadRequest)
			return
		}
	}

	// Validate the parsed document
	c.logger.Info("Begin validation of generic document")
	if err := c.validator.Validate(parsedBaseXml); err != nil {
		c.logger.Error("Document validation failed: " + err.Error())
		http.Error(w, "Invalid document data", http.StatusBadRequest)
		return
	}

	// Queue each document within the parsed XML set for further processing
	c.logger.Info("Queueing documents for processing")
	for i := range parsedBaseXml.Body.Documents {
		doc := &parsedBaseXml.Body.Documents[i]
		c.Queue.AddToQueue(doc, docType, format, func() {
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
