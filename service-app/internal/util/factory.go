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

	// Cast parsedDoc to LP1FDocument
	// TODO: this should be more generic and handle other document types.
	lp1fDoc, ok := parsedDoc.(*types.LP1FDocument)
	if !ok {
		return nil, fmt.Errorf("failed to cast parsed document to LP1FDocument")
	}
	sanitizer := lp1f.NewSanitizer()
	sanitizedData, err := sanitizer.Sanitize(lp1fDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize data: %w", err)
	}

	if errors := parser.GetErrors(); len(errors) > 0 {
		fmt.Println("Parsing completed with the following errors:")
		for _, e := range errors {
			fmt.Println(e)
		}
	}

	fmt.Printf("Parsed Document: %+v\n", sanitizedData)
	return sanitizedData, nil
}

// NewParser returns a DocumentParser based on the given document type and format.
// The returned parser is responsible for parsing the given data.
// Supported document types are "LP1F" for Lasting Power of Attorney forms.
// Supported formats are "xml" and "json".
func NewParser(data []byte, docType string, format string) (parser.DocumentParser, error) {
	switch docType {
	case "LP1F":
		var lp1fDoc *types.LP1FDocument
		var err error
		if format == "xml" {
			lp1fDoc, err = lp1f.ParseLP1FXml(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse LP1F XML: %w", err)
			}
		}
		return &lp1f.LP1FParser{Doc: lp1fDoc}, nil
	default:
		return nil, fmt.Errorf("unsupported document type: %s", docType)
	}
}
