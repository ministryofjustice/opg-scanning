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

func (q *Worker) Process(ctx context.Context, bodyStr string) (*sirius.ScannedCaseResponse, error) {
	filename, err := q.awsClient.PersistSetData(ctx, []byte(bodyStr))
	if err != nil {
		return nil, PersistSetError{Err: err}
	}

	q.logger.InfoContext(ctx, "Stored Set data", slog.String("set_filename", filename))

	set, err := q.validateAndSanitizeXML(ctx, bodyStr)
	if err != nil {
		return nil, ValidateAndSanitizeError{Err: err}
	}

	if err := q.validator.ValidateSet(set); err != nil {
		return nil, ValidateSetError{Err: err}
	}

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

func (c *Worker) validateAndSanitizeXML(ctx context.Context, bodyStr string) (*types.BaseSet, error) {
	schemaLocation, err := ExtractSchemaLocation(bodyStr)
	if err != nil {
		return nil, err
	}

	// Validate against XSD
	c.logger.InfoContext(ctx, "Validating against XSD")
	xsdValidator, err := NewXSDValidator(c.config, schemaLocation, bodyStr)
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
	c.logger.InfoContext(ctx, "Validating XML")
	xmlValidator := NewXmlValidator(*c.config)
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

func (c *Worker) validateDocument(document types.BaseDocument) error {
	if !slices.Contains(constants.SupportedDocumentTypes, document.Type) {
		return Problem{
			Title: fmt.Sprintf("Document type %s is not supported", document.Type),
		}
	}

	decodedXML, err := util.DecodeEmbeddedXML(document.EmbeddedXML)
	if err != nil {
		return fmt.Errorf("failed to decode XML data from %s: %w", document.Type, err)
	}

	schemaLocation, err := ExtractSchemaLocation(string(decodedXML))
	if err != nil {
		return fmt.Errorf("failed to extract schema from %s: %w", document.Type, err)
	}

	xsdValidator, err := NewXSDValidator(c.config, schemaLocation, string(decodedXML))
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
