package ingestion

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"

	"github.com/ministryofjustice/opg-scanning/internal/config"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/factory"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/sirius"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type AwsClient interface {
	PersistFormData(ctx context.Context, body io.Reader, docType string) (string, error)
	PersistSetData(ctx context.Context, body []byte) (string, error)
	QueueSetForProcessing(ctx context.Context, scannedCaseResponse *sirius.ScannedCaseResponse, fileName string) (string, error)
}

type SiriusService interface {
	AttachDocuments(ctx context.Context, set *types.BaseSet, originalDoc *types.BaseDocument, caseResponse *sirius.ScannedCaseResponse) (*sirius.ScannedDocumentResponse, []byte, error)
	CreateCaseStub(ctx context.Context, set *types.BaseSet) (*sirius.ScannedCaseResponse, error)
}

type JobQueue struct {
	logger        *slog.Logger
	siriusService SiriusService
	awsClient     AwsClient
}

func NewJobQueue(logger *slog.Logger, siriusService SiriusService, awsClient AwsClient) *JobQueue {
	return &JobQueue{
		logger:        logger,
		siriusService: siriusService,
		awsClient:     awsClient,
	}
}

func (q *JobQueue) AddToQueueSequentially(ctx context.Context, cfg *config.Config, data *types.BaseDocument, parsedBaseXml *types.BaseSet, scannedCaseResponse *sirius.ScannedCaseResponse) error {
	jobCtx := logger.NewContextFromOld(ctx)

	registry, err := factory.NewRegistry()
	if err != nil {
		return fmt.Errorf("failed to create registry: %v", err)
	}

	processor, err := factory.NewDocumentProcessor(data, data.Type, registry, q.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize processor: %v", err)
	}

	// Use a per-job timeout context.
	processCtx, cancel := context.WithTimeout(jobCtx, cfg.HTTP.Timeout)
	processCtx = context.WithValue(processCtx, constants.TokenContextKey, ctx.Value(constants.TokenContextKey))
	defer cancel()

	if _, err := processor.Process(processCtx); err != nil {
		return fmt.Errorf("failed to process job: %v", err)
	}

	return q.onComplete(processCtx, data, parsedBaseXml, scannedCaseResponse)
}

func (q *JobQueue) onComplete(ctx context.Context, originalDoc *types.BaseDocument, parsedBaseXml *types.BaseSet, scannedCaseResponse *sirius.ScannedCaseResponse) error {
	attchResp, decodedXML, docErr := q.siriusService.AttachDocuments(ctx, parsedBaseXml, originalDoc, scannedCaseResponse)
	if docErr != nil {
		return fmt.Errorf("failed to attach document: %w", docErr)
	}

	// Persist the processed document.
	fileName, persistErr := q.persist(ctx, decodedXML, originalDoc)
	if persistErr != nil {
		return fmt.Errorf("failed to persist document: %w", persistErr)
	}

	// If not a Sirius extraction document, skip external job processing.
	if !slices.Contains(constants.SiriusExtractionDocuments, originalDoc.Type) {
		q.logger.InfoContext(ctx, "Skipping external job processing, checks completed for document",
			slog.String("set_uid", scannedCaseResponse.UID),
			slog.String("pdf_uuid", attchResp.UUID),
			slog.String("filename", fileName),
			slog.String("document_type", originalDoc.Type),
		)
		return nil
	}

	q.logger.InfoContext(ctx, "Stored Form data", slog.String("filename", fileName))

	// Queue the document for external processing.
	messageID, err := q.awsClient.QueueSetForProcessing(ctx, scannedCaseResponse, fileName)
	if err != nil {
		q.logger.ErrorContext(ctx, "Failed to queue document for processing",
			slog.String("set_uid", scannedCaseResponse.UID),
			slog.String("document_type", originalDoc.Type),
			slog.String("error", err.Error()),
		)
		return err
	}

	q.logger.InfoContext(ctx, "Job processing completed for document",
		slog.String("set_uid", scannedCaseResponse.UID),
		slog.String("pdf_uuid", attchResp.UUID),
		slog.String("job_queue_id", messageID),
		slog.String("filename", fileName),
		slog.String("document_type", originalDoc.Type),
	)

	return nil
}

func (q *JobQueue) persist(ctx context.Context, decodedXML []byte, originalDoc *types.BaseDocument) (string, error) {
	xmlReader := bytes.NewReader(decodedXML)

	fileName, err := q.awsClient.PersistFormData(ctx, xmlReader, originalDoc.Type)
	if err != nil {
		return "", err
	}

	return fileName, nil
}
