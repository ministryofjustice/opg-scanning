package corresp_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/corresp_types"
)

type Validator struct {
	doc           *corresp_types.Correspondence
	baseValidator *parser.BaseValidator
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
	v.baseValidator = parser.NewBaseValidator(v.doc)

	return nil
}

func (v *Validator) Validate() error {
	// Return errors if any
	if messages := v.baseValidator.GetValidatorErrorMessages(); len(messages) > 0 {
		return fmt.Errorf("failed to validate Correspondence document: %v", messages)
	}
	return nil
}

func (v *Validator) GetValidatorErrorMessages() []string {
	return v.baseValidator.GetValidatorErrorMessages()
}
