package factory

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

// DocumentProcessor processes documents through parsing, validation, and sanitization.
type DocumentProcessor struct {
	logger    logger.Logger
	doc       interface{}
	validator parser.CommonValidator
	sanitizer parser.CommonSanitizer
}

// NewDocumentProcessor initializes a new DocumentProcessor.
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
func (p *DocumentProcessor) Process() (interface{}, error) {
	// Validate the document
	p.validator.Setup(p.doc)
	if err := p.validator.Validate(); err != nil {
		p.logger.Error("Validation failed: " + err.Error())
	}

	// Sanitize the document
	p.sanitizer.Setup(p.doc)
	sanitizedDoc, err := p.sanitizer.Sanitize()
	if err != nil {
		return nil, fmt.Errorf("sanitization failed: %w", err)
	}

	p.doc = sanitizedDoc
	return p.doc, nil
}

// GetDocument retrieves the processed document.
func (p *DocumentProcessor) GetDocument() interface{} {
	return p.doc
}
