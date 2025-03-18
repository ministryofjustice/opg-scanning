package factory

import (
	"context"
	"errors"
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type DocumentProcessor struct {
	logger    logger.Logger
	doc       interface{}
	validator parser.CommonValidator
	sanitizer parser.CommonSanitizer
}

// Initializes a new DocumentProcessor.
func NewDocumentProcessor(data *types.BaseDocument, docType, format string, registry RegistryInterface, logger *logger.Logger) (*DocumentProcessor, error) {
	// Decode the embedded XML
	embeddedXML, err := util.DecodeEmbeddedXML(data.EmbeddedXML)
	if err != nil {
		return nil, fmt.Errorf("failed to decode embedded XML: %w", err)
	}

	return nil, errors.New("this is an error")

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

	// Fetch the sanitizer
	sanitizer, err := registry.GetSanitizer(docType)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sanitizer: %w", err)
	}

	return &DocumentProcessor{
		logger:    *logger,
		doc:       parsedDoc,
		validator: validator,
		sanitizer: sanitizer,
	}, nil
}

// Process validates and sanitizes the document.
func (p *DocumentProcessor) Process(ctx context.Context) (interface{}, error) {
	// If the document type doesn't declare a validator or sanitizer, skip.
	if p.validator == nil || p.sanitizer == nil {
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

	// Sanitize the document
	if err := p.sanitizer.Setup(p.doc); err != nil {
		return nil, fmt.Errorf("sanitization setup failed: %w", err)
	}

	sanitizedDoc, err := p.sanitizer.Sanitize()
	if err != nil {
		return nil, fmt.Errorf("sanitization failed: %w", err)
	}

	p.doc = sanitizedDoc
	return p.doc, nil
}
