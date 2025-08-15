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
	"github.com/lestrrat-go/libxml2/xsd"
	"github.com/ministryofjustice/opg-scanning/internal/config"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/factory"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/sirius"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

var ErrScannedCaseResponseUIDMissing = errors.New("scannedCaseResponse UID missing")

type Problem struct {
	Title            string
	ValidationErrors []string
}

func (p Problem) Error() string {
	return p.Title
}

type FailedToCreateCaseStubError struct {
	Err error
}

func (e FailedToCreateCaseStubError) Error() string { return e.Err.Error() }
func (e FailedToCreateCaseStubError) Unwrap() error { return e.Err }

type ValidateSetError struct {
	Err error
}

func (e ValidateSetError) Error() string { return e.Err.Error() }
func (e ValidateSetError) Unwrap() error { return e.Err }

type ValidateAndSanitizeError struct {
	Err error
}

func (e ValidateAndSanitizeError) Error() string { return e.Err.Error() }
func (e ValidateAndSanitizeError) Unwrap() error { return e.Err }

type PersistSetError struct {
	Err error
}

func (e PersistSetError) Error() string { return e.Err.Error() }
func (e PersistSetError) Unwrap() error { return e.Err }

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
	validator       *Validator
}

func NewWorker(logger *slog.Logger, config *config.Config, awsClient AwsClient, dynamoClient *dynamodb.Client) *Worker {
	return &Worker{
		logger:          logger,
		config:          config,
		siriusService:   sirius.NewService(config),
		awsClient:       awsClient,
		documentTracker: NewDocumentTracker(dynamoClient, config.Aws.DocumentsTable),
		validator:       NewValidator(),
	}
}

func (w *Worker) Process(ctx context.Context, body []byte) (*sirius.ScannedCaseResponse, error) {
	filename, err := w.awsClient.PersistSetData(ctx, body)
	if err != nil {
		return nil, PersistSetError{Err: err}
	}

	w.logger.InfoContext(ctx, "Stored Set data", slog.String("set_filename", filename))

	set, err := w.validateAndSanitizeXML(ctx, body)
	if err != nil {
		return nil, ValidateAndSanitizeError{Err: err}
	}

	if err := w.validator.ValidateSet(set); err != nil {
		return nil, ValidateSetError{Err: err}
	}

	scannedCaseResponse, err := w.createCaseStub(ctx, set)
	if err != nil {
		return scannedCaseResponse, err
	}

	w.logger.InfoContext(ctx, "Queueing documents for processing", slog.Any("Header", set.Header))

	// Iterate over each document in the parsed set.
	for i := range set.Body.Documents {
		doc := &set.Body.Documents[i]

		ctx := logger.ContextWithAttrs(ctx,
			slog.String("set_uid", scannedCaseResponse.UID),
			slog.String("document_id", doc.ID),
			slog.String("document_type", doc.Type),
		)

		if err := w.documentTracker.SetProcessing(ctx, doc.ID, scannedCaseResponse.UID); err != nil {
			return scannedCaseResponse, fmt.Errorf("failed to set document to processing '%s': %w", doc.ID, err)
		}

		if err := w.processDocument(ctx, set, doc, scannedCaseResponse); err != nil {
			if err := w.documentTracker.SetFailed(ctx, doc.ID); err != nil {
				w.logger.ErrorContext(ctx, err.Error())
			}

			if !errors.As(err, &sirius.Error{}) {
				w.logger.ErrorContext(ctx, err.Error())
			}

			return scannedCaseResponse, err
		}

		if err := w.documentTracker.SetCompleted(ctx, doc.ID); err != nil {
			w.logger.ErrorContext(ctx, err.Error())
		}

		w.logger.InfoContext(ctx, "Document added for processing")
	}

	w.logger.InfoContext(ctx, "No errors found!")
	return scannedCaseResponse, nil
}

func (w *Worker) createCaseStub(ctx context.Context, set *types.BaseSet) (*sirius.ScannedCaseResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, w.config.HTTP.Timeout)
	defer cancel()

	scannedCaseResponse, err := w.siriusService.CreateCaseStub(ctx, set)
	if err != nil {
		return nil, FailedToCreateCaseStubError{Err: err}
	}

	if scannedCaseResponse == nil || scannedCaseResponse.UID == "" {
		return nil, ErrScannedCaseResponseUIDMissing
	}

	return scannedCaseResponse, nil
}

func (w *Worker) processDocument(ctx context.Context, set *types.BaseSet, document *types.BaseDocument, scannedCaseResponse *sirius.ScannedCaseResponse) error {
	ctx, cancel := context.WithTimeout(ctx, w.config.HTTP.Timeout)
	defer cancel()

	ctx = context.WithValue(ctx, constants.TokenContextKey, ctx.Value(constants.TokenContextKey))

	registry, err := factory.NewRegistry()
	if err != nil {
		return fmt.Errorf("failed to create registry: %v", err)
	}

	processor, err := factory.NewDocumentProcessor(document, document.Type, registry, w.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize processor: %v", err)
	}

	if _, err := processor.Process(ctx); err != nil {
		return fmt.Errorf("failed to process job: %v", err)
	}

	attchResp, decodedXML, docErr := w.siriusService.AttachDocuments(ctx, set, document, scannedCaseResponse)
	if docErr != nil {
		return fmt.Errorf("failed to attach document: %w", docErr)
	}

	// Persist the processed document.
	fileName, persistErr := w.persist(ctx, decodedXML, document)
	if persistErr != nil {
		return fmt.Errorf("failed to persist document: %w", persistErr)
	}

	// If not a Sirius extraction document, skip external job processing.
	if !slices.Contains(constants.SiriusExtractionDocuments, document.Type) {
		w.logger.InfoContext(ctx, "Skipping external job processing, checks completed for document",
			slog.String("pdf_uuid", attchResp.UUID),
			slog.String("filename", fileName),
		)
		return nil
	}

	w.logger.InfoContext(ctx, "Stored Form data", slog.String("filename", fileName))

	// Queue the document for external processing.
	messageID, err := w.awsClient.QueueSetForProcessing(ctx, scannedCaseResponse, fileName)
	if err != nil {
		w.logger.ErrorContext(ctx, "Failed to queue document for processing",
			slog.String("error", err.Error()),
		)
		return err
	}

	w.logger.InfoContext(ctx, "Job processing completed for document",
		slog.String("pdf_uuid", attchResp.UUID),
		slog.String("job_queue_id", messageID),
		slog.String("filename", fileName),
	)

	return nil
}

func (w *Worker) persist(ctx context.Context, decodedXML []byte, originalDoc *types.BaseDocument) (string, error) {
	xmlReader := bytes.NewReader(decodedXML)

	fileName, err := w.awsClient.PersistFormData(ctx, xmlReader, originalDoc.Type)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func (w *Worker) validateAndSanitizeXML(ctx context.Context, body []byte) (*types.BaseSet, error) {
	schemaLocation, err := ExtractSchemaLocation(body)
	if err != nil {
		return nil, err
	}

	// Validate against XSD
	w.logger.InfoContext(ctx, "Validating against XSD")
	xsdValidator, err := NewXSDValidator(w.config, schemaLocation, body)
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
	w.logger.InfoContext(ctx, "Validating XML")
	xmlValidator := NewXmlValidator(*w.config)
	parsedBaseXml, err := xmlValidator.XmlValidate(body)
	if err != nil {
		return nil, err
	}

	// Validate embedded documents
	for _, document := range parsedBaseXml.Body.Documents {
		if err := w.validateDocument(document); err != nil {
			return nil, err
		}
	}

	return parsedBaseXml, nil
}

func (w *Worker) validateDocument(document types.BaseDocument) error {
	if !slices.Contains(constants.SupportedDocumentTypes, document.Type) {
		return Problem{
			Title: fmt.Sprintf("Document type %s is not supported", document.Type),
		}
	}

	decodedXML, err := util.DecodeEmbeddedXML(document.EmbeddedXML)
	if err != nil {
		return fmt.Errorf("failed to decode XML data from %s: %w", document.Type, err)
	}

	schemaLocation, err := ExtractSchemaLocation(decodedXML)
	if err != nil {
		return fmt.Errorf("failed to extract schema from %s: %w", document.Type, err)
	}

	xsdValidator, err := NewXSDValidator(w.config, schemaLocation, decodedXML)
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
