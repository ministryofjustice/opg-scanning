package util

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lp1f"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

func ProcessDocument(data []byte, docType, format string) (interface{}, error) {
	parser, err := NewParser(data, docType, format)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	parsedDoc, parseErr := parser.ParseDocument(data)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse document: %w", parseErr)
	}

	if errors := parser.GetErrors(); len(errors) > 0 {
		fmt.Println("Parsing completed with the following errors:")
		for _, e := range errors {
			fmt.Println(e)
		}
	}

	fmt.Printf("Parsed Document: %+v\n", parsedDoc)
	return parsedDoc, nil
}

// NewParser returns a DocumentParser based on the given document type and format.
// The returned parser is responsible for parsing the given data.
// Supported document types are "LP1F" for Lasting Power of Attorney forms.
// Supported formats are "xml" and "json".
func NewParser(data []byte, docType string, format string) (parser.DocumentParser, error) {
	switch docType {
	case "LP1F":
		lp1fDoc, err := parseLP1FDocument(data, format) // Helper function based on format
		if err != nil {
			return nil, fmt.Errorf("failed to parse LP1F %s: %w", format, err)
		}
		return &lp1f.LP1FParser{Doc: lp1fDoc}, nil
	default:
		return nil, fmt.Errorf("unsupported document type: %s", docType)
	}
}

// parseLP1FDocument parses an LP1F document from the given data, based on the given format.
// It currently supports "xml" and will return an error for other formats.
func parseLP1FDocument(data []byte, format string) (*types.LP1FDocument, error) {
	switch format {
	case "xml":
		return lp1f.ParseLP1FXml(data)
	case "json":
		// return lp1f.ParseLP1FJson(data) // Add JSON parsing function
	}

	return nil, fmt.Errorf("unsupported format: %s", format)
}
