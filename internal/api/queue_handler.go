package api

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/sirius"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

// Handles the queueing of documents for processing.
func (c *IndexController) processQueue(ctx context.Context, scannedCaseResponse *sirius.ScannedCaseResponse, parsedBaseXml *types.BaseSet) error {
	c.logger.InfoWithContext(ctx, "Queueing documents for processing", map[string]any{
		"Header": parsedBaseXml.Header,
	})

	// Iterate over each document in the parsed set.
	for i := range parsedBaseXml.Body.Documents {
		doc := &parsedBaseXml.Body.Documents[i]
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
				c.logger.InfoWithContext(ctx, "Skipping external job processing, checks completed for document", map[string]any{
					"set_uid":       scannedCaseResponse.UID,
					"pdf_uuid":      attchResp.UUID,
					"filename":      fileName,
					"document_type": originalDoc.Type,
				})
				return nil
			}

			c.logger.InfoWithContext(ctx, "Stored Form data", map[string]any{
				"filename": fileName,
			})

			// Queue the document for external processing.
			messageID, err := c.AwsClient.QueueSetForProcessing(ctx, scannedCaseResponse, fileName)
			if err != nil {
				c.logger.ErrorWithContext(ctx, "Failed to queue document for processing", map[string]any{
					"set_uid":       scannedCaseResponse.UID,
					"document_type": originalDoc.Type,
					"error":         err.Error(),
				})
				return err
			}

			c.logger.InfoWithContext(ctx, "Job processing completed for document", map[string]any{
				"set_uid":       scannedCaseResponse.UID,
				"pdf_uuid":      attchResp.UUID,
				"job_queue_id":  messageID,
				"filename":      fileName,
				"document_type": originalDoc.Type,
			})

			return nil
		})

		if err != nil {
			if (!errors.As(err, &sirius.Error{})) {
				c.logger.ErrorWithContext(ctx, err.Error(), map[string]any{
					"set_uid":       scannedCaseResponse.UID,
					"document_type": doc.Type,
				})
			}

			return err
		}

		c.logger.InfoWithContext(ctx, "Document added for processing", map[string]any{
			"set_uid":       scannedCaseResponse.UID,
			"document_type": doc.Type,
		})
	}

	c.logger.InfoWithContext(ctx, "No errors found!", nil)
	return nil
}
