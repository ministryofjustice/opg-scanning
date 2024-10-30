package api

import (
	"net/http"

	"github.com/ministryofjustice/opg-scanning/internal/ingestion"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

type IndexController struct {
	validator *ingestion.Validator
	queue     *ingestion.JobQueue
	logger    *logger.Logger
}

func NewIndexController() *IndexController {
	return &IndexController{
		validator: ingestion.NewValidator(),
		queue:     ingestion.NewJobQueue(),
		logger:    logger.NewLogger(),
	}
}

func (c *IndexController) HandleRequests() {
	http.HandleFunc("/ingest", c.IngestHandler)
	http.ListenAndServe(":8080", nil)
}

func (c *IndexController) IngestHandler(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("Received ingestion request")

	// Extract XML from request
	data := []byte(r.FormValue("data"))
	docType := r.FormValue("docType")
	format := r.FormValue("format")

	// Perform validation
	if err := c.validator.Validate(string(data)); err != nil {
		c.logger.Error("Validation failed: " + err.Error())
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	// Queue the parsed document for further processing
	c.queue.AddToQueue(data, docType, format)
	c.logger.Info("Job added to queue")
	w.WriteHeader(http.StatusAccepted)
}

func (c *IndexController) CloseQueue() {
	c.queue.Close()
}
