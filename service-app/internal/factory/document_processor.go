package factory

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type DocumentProcessor struct {
	logger    logger.Logger
	doc       any
	validator parser.CommonValidator
}

// Initializes a new DocumentProcessor.
func NewDocumentProcessor(data *types.BaseDocument, docType, format string, registry RegistryInterface, logger *logger.Logger) (*DocumentProcessor, error) {
	// Decode the embedded XML
	embeddedXML, err := util.DecodeEmbeddedXML(data.EmbeddedXML)
	if err != nil {
		return nil, fmt.Errorf("failed to decode embedded XML: %w", err)
	}

	// Fetch the parser
	parser, err := registry.GetParser(docType)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve parser: %w", err)
	}

	// Parse the document
	parsedDoc, err := parser(embeddedXML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	// Fetch and create the validator
	validator, err := registry.GetValidator(docType)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve validator factory: %w", err)
	}

	return &DocumentProcessor{
		logger:    *logger,
		doc:       parsedDoc,
		validator: validator,
	}, nil
}

// Process validates and sanitizes the document.
func (p *DocumentProcessor) Process(ctx context.Context) (any, error) {
	// If the document type doesn't declare a validator or sanitizer, skip.
	if p.validator == nil {
		return p.doc, nil
	}

	// Validate the document
	if err := p.validator.Setup(p.doc); err != nil {
		return nil, fmt.Errorf("validation setup failed: %w", err)
	}

	// Return an error if any validations failed.
	if messages := p.validator.Validate(); len(messages) > 0 {
		p.logger.Info("Validation failed: %v", nil, messages)
	}

	return p.doc, nil
}
