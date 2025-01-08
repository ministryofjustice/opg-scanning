package corresp_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/corresp_types"
)

type Validator struct {
	doc             *corresp_types.Correspondence
	commonValidator *parser.Validator
}

func NewValidator() *Validator {
	return &Validator{
		doc: &corresp_types.Correspondence{},
	}
}

func (v *Validator) Setup(doc interface{}) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.doc = doc.(*corresp_types.Correspondence)
	v.commonValidator = parser.NewValidator(v.doc)

	return nil
}

func (v *Validator) Validate() error {
	// Common witness validations
	v.commonValidator.WitnessSignatureFullNameAddressValidator("Page10", "Section9")

	// Return errors if any
	if messages := v.commonValidator.GetValidatorErrorMessages(); len(messages) > 0 {
		return fmt.Errorf("failed to validate LP1F document: %v", messages)
	}
	return nil
}
