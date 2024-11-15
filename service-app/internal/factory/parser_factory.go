package factory

import (
	"fmt"

	lp1f_parser "github.com/ministryofjustice/opg-scanning/internal/parser/lp1f_parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

func getParser(docType string) (types.Parser, error) {
	switch docType {
	case DocumentTypeLP1F:
		return lp1f_parser.NewParser(), nil
	// Add cases for other document types here
	default:
		return nil, fmt.Errorf("parser not found for document type: %s", docType)
	}
}
