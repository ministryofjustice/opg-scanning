package generic_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
)

type Validator struct {
	doc           interface{}
	baseValidator *parser.BaseValidator
}

func NewValidator() *Validator {
	return &Validator{
		doc: struct{}{},
	}
}

func (v *Validator) Setup(doc interface{}) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.baseValidator = parser.NewBaseValidator(v.doc)

	return nil
}

func (v *Validator) Validate() error {
	// Return errors if any
	if messages := v.baseValidator.GetValidatorErrorMessages(); len(messages) > 0 {
		return fmt.Errorf("failed to validate generic document: %v", messages)
	}
	return nil
}
