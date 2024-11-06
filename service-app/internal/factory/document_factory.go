package factory

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

const (
	DocumentTypeLP1F = "LP1F"
)

// type DocumentFactory struct{}

func NewDocumentProcessor(data *types.BaseDocument, docType, format string) (*DocumentProcessor, error) {
	doc, err := util.DecodeEmbeddedXML(data.EmbeddedXML)
	if err != nil {
		return nil, fmt.Errorf("failed to decode embedded XML: %w", err)
	}

	parser, err := getParser(docType)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve parser: %w", err)
	}

	parsedDoc, err := parser.Parse(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	validator, err := getValidator(docType, parsedDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve validator: %w", err)
	}

	sanitizer, err := getSanitizer(docType)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sanitizer: %w", err)
	}

	return &DocumentProcessor{
		doc:       parsedDoc,
		validator: validator,
		sanitizer: sanitizer,
	}, nil
}
