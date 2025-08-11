package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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

		if err := c.Queue.AddToQueueSequentially(ctx, c.config, doc, parsedBaseXml, scannedCaseResponse); err != nil {
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
