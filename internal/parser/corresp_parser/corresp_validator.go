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

func (v *Validator) Setup(doc any) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.doc = doc.(*corresp_types.Correspondence)
	v.baseValidator = parser.NewBaseValidator(v.doc)

	return nil
}

func (v *Validator) Validate() []string {
	return v.baseValidator.GetValidatorErrorMessages()
}
