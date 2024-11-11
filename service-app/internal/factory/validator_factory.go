package factory

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	lp1f_parser "github.com/ministryofjustice/opg-scanning/internal/parser/lp1f_parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	lp1f_types "github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
)

func getValidator(docType string, doc interface{}) (types.Validator, error) {
	commonValidator := parser.NewCommonValidator(doc)

	switch docType {
	case DocumentTypeLP1F:
		commonValidator.Validate("Page10", "Section9")
		commonValidator.Validate("Page12", "Section11")
		// TODO: introduce other validation types.

		if commonValidator.GetValidatorErrorMessages() != nil {
			return nil, fmt.Errorf("failed to validate LP1F document: %v", commonValidator.GetValidatorErrorMessages())
		}

		if lp1fDoc, ok := doc.(*lp1f_types.LP1FDocument); ok {
			return lp1f_parser.NewValidator(lp1fDoc), nil
		}
		return nil, fmt.Errorf("invalid document type for LP1F validator")
	// Add cases for other document types here
	default:
		return nil, fmt.Errorf("validator not found for document type: %s", docType)
	}
}
