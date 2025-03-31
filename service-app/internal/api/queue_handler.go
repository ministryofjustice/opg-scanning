package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

// Handles the queueing of documents for processing.
func (c *IndexController) ProcessQueue(ctx context.Context, scannedCaseResponse *types.ScannedCaseResponse, parsedBaseXml *types.BaseSet) error {
	// Clear any queued errors when done.
	defer c.Queue.ClearErrors()

	c.logger.InfoWithContext(ctx, "Queueing documents for processing", map[string]any{
		"Header": parsedBaseXml.Header,
	})

	// Iterate over each document in the parsed set.
	for i := range parsedBaseXml.Body.Documents {
		doc := &parsedBaseXml.Body.Documents[i]
		c.Queue.AddToQueue(ctx, c.config, doc, "xml", func(ctx context.Context, processedDoc any, originalDoc *types.BaseDocument) error {
			// Wrap the job context with a timeout.
			ctx, cancel := context.WithTimeout(ctx, time.Duration(c.config.HTTP.Timeout)*time.Second)
			defer cancel()

			// Create a new service instance for attaching documents.
			service := NewService(NewClient(c.httpMiddleware), parsedBaseXml)
			service.originalDoc = originalDoc

			attchResp, decodedXML, docErr := service.AttachDocuments(ctx, scannedCaseResponse)
			if docErr != nil {
				c.logger.ErrorWithContext(ctx, "Failed to attach document", map[string]any{
					"set_uid":       scannedCaseResponse.UID,
					"document_type": originalDoc.Type,
					"error":         docErr.Error(),
				})
				return docErr
			}

			// Persist the processed document.
			fileName, persistErr := c.processAndPersist(ctx, decodedXML, originalDoc)
			if persistErr != nil {
				c.logger.ErrorWithContext(ctx, "Failed to persist document", map[string]any{
					"set_uid":       scannedCaseResponse.UID,
					"document_type": originalDoc.Type,
					"error":         persistErr.Error(),
				})
				return persistErr
			}

			// If not a Sirius extraction document, skip external job processing.
			if !util.Contains(constants.SiriusExtractionDocuments, originalDoc.Type) {
				c.logger.InfoWithContext(ctx, "Skipping external job processing, checks completed for document", map[string]any{
					"set_uid":       scannedCaseResponse.UID,
					"pdf_uuid":      attchResp.UUID,
					"filename":      fileName,
					"document_type": originalDoc.Type,
				})
				return nil
			}

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

		c.logger.InfoWithContext(ctx, "Document queued for processing", map[string]any{
			"set_uid":       scannedCaseResponse.UID,
			"document_type": doc.Type,
		})
	}

	// Wait for all jobs to finish.
	c.Queue.Wait()

	// Retrieve and handle any errors that occurred during processing.
	jobErrors := c.Queue.GetErrors()
	if len(jobErrors) > 0 {
		var errorMessages []string
		for _, err := range jobErrors {
			errorMessages = append(errorMessages, err.Error())
		}
		return errors.New(fmt.Sprintf("Errors encountered during processing: %s", strings.Join(errorMessages, "; ")))
	}

	c.logger.InfoWithContext(ctx, "No errors found!", nil)
	return nil
}
