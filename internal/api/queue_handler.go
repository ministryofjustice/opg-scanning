package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/sirius"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

// Handles the queueing of documents for processing.
func (c *IndexController) processQueue(ctx context.Context, scannedCaseResponse *sirius.ScannedCaseResponse, parsedBaseXml *types.BaseSet) error {
	c.logger.InfoContext(ctx, "Queueing documents for processing", slog.Any("Header", parsedBaseXml.Header))

	// Iterate over each document in the parsed set.
	for i := range parsedBaseXml.Body.Documents {
		doc := &parsedBaseXml.Body.Documents[i]

		if err := c.documentTracker.SetProcessing(ctx, doc.ID, scannedCaseResponse.UID); err != nil {
			return fmt.Errorf("failed to set document to processing '%s': %w", doc.ID, err)
		}

		// Here we use AddToQueueSequentially to ensure that the documents are processed in order.
		err := c.Queue.AddToQueueSequentially(ctx, c.config, doc, "xml", func(ctx context.Context, processedDoc any, originalDoc *types.BaseDocument) error {
			// Create a new service instance for attaching documents.
			service := newService(c.siriusClient, parsedBaseXml)
			service.originalDoc = originalDoc

			attchResp, decodedXML, docErr := service.AttachDocuments(ctx, scannedCaseResponse)
			if docErr != nil {
				return fmt.Errorf("failed to attach document: %w", docErr)
			}

			// Persist the processed document.
			fileName, persistErr := c.processAndPersist(ctx, decodedXML, originalDoc)
			if persistErr != nil {
				return fmt.Errorf("failed to persist document: %w", persistErr)
			}

			// If not a Sirius extraction document, skip external job processing.
			if !slices.Contains(constants.SiriusExtractionDocuments, originalDoc.Type) {
				c.logger.InfoContext(ctx, "Skipping external job processing, checks completed for document",
					slog.String("set_uid", scannedCaseResponse.UID),
					slog.String("pdf_uuid", attchResp.UUID),
					slog.String("filename", fileName),
					slog.String("document_type", originalDoc.Type),
				)
				return nil
			}

			c.logger.InfoContext(ctx, "Stored Form data", slog.String("filename", fileName))

			// Queue the document for external processing.
			messageID, err := c.AwsClient.QueueSetForProcessing(ctx, scannedCaseResponse, fileName)
			if err != nil {
				c.logger.ErrorContext(ctx, "Failed to queue document for processing",
					slog.String("set_uid", scannedCaseResponse.UID),
					slog.String("document_type", originalDoc.Type),
					slog.String("error", err.Error()),
				)
				return err
			}

			c.logger.InfoContext(ctx, "Job processing completed for document",
				slog.String("set_uid", scannedCaseResponse.UID),
				slog.String("pdf_uuid", attchResp.UUID),
				slog.String("job_queue_id", messageID),
				slog.String("filename", fileName),
				slog.String("document_type", originalDoc.Type),
			)

			return nil
		})

		if err != nil {
			if err := c.documentTracker.SetFailed(ctx, doc.ID); err != nil {
				c.logger.ErrorContext(ctx, err.Error(),
					slog.String("set_uid", scannedCaseResponse.UID),
					slog.String("document_type", doc.Type),
				)
			}

			if !errors.As(err, &sirius.Error{}) {
				c.logger.ErrorContext(ctx, err.Error(),
					slog.String("set_uid", scannedCaseResponse.UID),
					slog.String("document_type", doc.Type),
				)
			}

			return err
		}

		if err := c.documentTracker.SetCompleted(ctx, doc.ID); err != nil {
			c.logger.ErrorContext(ctx, err.Error(),
				slog.String("set_uid", scannedCaseResponse.UID),
				slog.String("document_type", doc.Type),
			)
		}

		c.logger.InfoContext(ctx, "Document added for processing",
			slog.String("set_uid", scannedCaseResponse.UID),
			slog.String("document_type", doc.Type),
		)
	}

	c.logger.InfoContext(ctx, "No errors found!")
	return nil
}
