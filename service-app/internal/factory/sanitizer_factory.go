package factory

import (
	"fmt"

	lp1f_parser "github.com/ministryofjustice/opg-scanning/internal/parser/lp1f_parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

func getSanitizer(docType string) (types.Sanitizer, error) {
	switch docType {
	case DocumentTypeLP1F:
		return lp1f_parser.NewSanitizer(), nil
	// Add cases for other document types here
	default:
		return nil, fmt.Errorf("sanitizer not found for document type: %s", docType)
	}
}
