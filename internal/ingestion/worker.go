package ingestion

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ministryofjustice/opg-scanning/internal/config"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/factory"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/sirius"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type FailedToCreateCaseStubError struct {
	Err error
}

func (e FailedToCreateCaseStubError) Error() string { return e.Err.Error() }

var ErrScannedCaseResponseUIDMissing = errors.New("scannedCaseResponse UID missing")

type documentTracker interface {
	SetProcessing(ctx context.Context, id, caseNo string) error
	SetCompleted(ctx context.Context, id string) error
	SetFailed(ctx context.Context, id string) error
}

type AwsClient interface {
	PersistFormData(ctx context.Context, body io.Reader, docType string) (string, error)
	PersistSetData(ctx context.Context, body []byte) (string, error)
	QueueSetForProcessing(ctx context.Context, scannedCaseResponse *sirius.ScannedCaseResponse, fileName string) (string, error)
}

type SiriusService interface {
	AttachDocuments(ctx context.Context, set *types.BaseSet, originalDoc *types.BaseDocument, caseResponse *sirius.ScannedCaseResponse) (*sirius.ScannedDocumentResponse, []byte, error)
	CreateCaseStub(ctx context.Context, set *types.BaseSet) (*sirius.ScannedCaseResponse, error)
}

type Worker struct {
	logger          *slog.Logger
	config          *config.Config
	siriusService   SiriusService
	awsClient       AwsClient
	documentTracker documentTracker
}

func NewWorker(logger *slog.Logger, config *config.Config, siriusService SiriusService, awsClient AwsClient, dynamoClient *dynamodb.Client) *Worker {
	return &Worker{
		logger:          logger,
		config:          config,
		siriusService:   siriusService,
		awsClient:       awsClient,
		documentTracker: NewDocumentTracker(dynamoClient, config.Aws.DocumentsTable),
	}
}

func (q *Worker) Process(ctx context.Context, set *types.BaseSet) (*sirius.ScannedCaseResponse, error) {
	scannedCaseResponse, err := q.createCaseStub(ctx, set)
	if err != nil {
		return scannedCaseResponse, err
	}

	q.logger.InfoContext(ctx, "Queueing documents for processing", slog.Any("Header", set.Header))

	// Iterate over each document in the parsed set.
	for i := range set.Body.Documents {
		doc := &set.Body.Documents[i]

		ctx := logger.ContextWithAttrs(ctx,
			slog.String("set_uid", scannedCaseResponse.UID),
			slog.String("document_id", doc.ID),
			slog.String("document_type", doc.Type),
		)

		if err := q.documentTracker.SetProcessing(ctx, doc.ID, scannedCaseResponse.UID); err != nil {
			return scannedCaseResponse, fmt.Errorf("failed to set document to processing '%s': %w", doc.ID, err)
		}

		if err := q.processDocument(ctx, set, doc, scannedCaseResponse); err != nil {
			if err := q.documentTracker.SetFailed(ctx, doc.ID); err != nil {
				q.logger.ErrorContext(ctx, err.Error())
			}

			if !errors.As(err, &sirius.Error{}) {
				q.logger.ErrorContext(ctx, err.Error())
			}

			return scannedCaseResponse, err
		}

		if err := q.documentTracker.SetCompleted(ctx, doc.ID); err != nil {
			q.logger.ErrorContext(ctx, err.Error())
		}

		q.logger.InfoContext(ctx, "Document added for processing")
	}

	q.logger.InfoContext(ctx, "No errors found!")
	return scannedCaseResponse, nil
}

func (q *Worker) createCaseStub(ctx context.Context, set *types.BaseSet) (*sirius.ScannedCaseResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, q.config.HTTP.Timeout)
	defer cancel()

	scannedCaseResponse, err := q.siriusService.CreateCaseStub(ctx, set)
	if err != nil {
		return nil, FailedToCreateCaseStubError{Err: err}
	}

	if scannedCaseResponse == nil || scannedCaseResponse.UID == "" {
		return nil, ErrScannedCaseResponseUIDMissing
	}

	return scannedCaseResponse, nil
}

func (q *Worker) processDocument(ctx context.Context, set *types.BaseSet, document *types.BaseDocument, scannedCaseResponse *sirius.ScannedCaseResponse) error {
	ctx, cancel := context.WithTimeout(ctx, q.config.HTTP.Timeout)
	defer cancel()

	ctx = context.WithValue(ctx, constants.TokenContextKey, ctx.Value(constants.TokenContextKey))

	registry, err := factory.NewRegistry()
	if err != nil {
		return fmt.Errorf("failed to create registry: %v", err)
	}

	processor, err := factory.NewDocumentProcessor(document, document.Type, registry, q.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize processor: %v", err)
	}

	if _, err := processor.Process(ctx); err != nil {
		return fmt.Errorf("failed to process job: %v", err)
	}

	attchResp, decodedXML, docErr := q.siriusService.AttachDocuments(ctx, set, document, scannedCaseResponse)
	if docErr != nil {
		return fmt.Errorf("failed to attach document: %w", docErr)
	}

	// Persist the processed document.
	fileName, persistErr := q.persist(ctx, decodedXML, document)
	if persistErr != nil {
		return fmt.Errorf("failed to persist document: %w", persistErr)
	}

	// If not a Sirius extraction document, skip external job processing.
	if !slices.Contains(constants.SiriusExtractionDocuments, document.Type) {
		q.logger.InfoContext(ctx, "Skipping external job processing, checks completed for document",
			slog.String("pdf_uuid", attchResp.UUID),
			slog.String("filename", fileName),
		)
		return nil
	}

	q.logger.InfoContext(ctx, "Stored Form data", slog.String("filename", fileName))

	// Queue the document for external processing.
	messageID, err := q.awsClient.QueueSetForProcessing(ctx, scannedCaseResponse, fileName)
	if err != nil {
		q.logger.ErrorContext(ctx, "Failed to queue document for processing",
			slog.String("error", err.Error()),
		)
		return err
	}

	q.logger.InfoContext(ctx, "Job processing completed for document",
		slog.String("pdf_uuid", attchResp.UUID),
		slog.String("job_queue_id", messageID),
		slog.String("filename", fileName),
	)

	return nil
}

func (q *Worker) persist(ctx context.Context, decodedXML []byte, originalDoc *types.BaseDocument) (string, error) {
	xmlReader := bytes.NewReader(decodedXML)

	fileName, err := q.awsClient.PersistFormData(ctx, xmlReader, originalDoc.Type)
	if err != nil {
		return "", err
	}

	return fileName, nil
}
