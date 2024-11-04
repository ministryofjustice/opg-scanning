package util

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser/lp1f"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

func ProcessDocument(data *types.Document, docType string, format string) (interface{}, error) {
	// TODO: Validate docType against xsd schema

	// Parse the document based on the document type
	parsedDoc, err := NewXMLParser(data, docType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	// Cast parsedDoc to LP1FDocument for sanitization
	// TODO: this should be more generic and handle other document types.
	lp1fDoc, ok := parsedDoc.(*types.LP1FDocument)
	if !ok {
		return nil, fmt.Errorf("failed to cast parsed document to LP1FDocument")
	}

	// Sanitize the parsed LP1F document
	sanitizer := lp1f.NewSanitizer()
	sanitizedData, err := sanitizer.Sanitize(lp1fDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize data: %w", err)
	}

	//fmt.Printf("Parsed Document: %+v\n", sanitizedData)
	return sanitizedData, nil
}

func NewXMLParser(data *types.Document, docType string) (interface{}, error) {
	// Decode embedded XML
	xml, err := DecodeEmbeddedXML(data.EmbeddedXML)
	if err != nil {
		return nil, fmt.Errorf("failed to decode embedded XML: %w", err)
	}

	// Parse based on document type
	switch docType {
	case "LP1F":
		lp1fDoc, err := lp1f.ParseLP1FXml(xml)
		if err != nil {
			return nil, fmt.Errorf("failed to parse LP1F XML: %w", err)
		}
		return lp1fDoc, nil // Return the parsed document directly
	default:
		return nil, fmt.Errorf("unsupported document type: %s", docType)
	}
}
