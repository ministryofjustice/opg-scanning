package lpa115_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lpa115_types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type Validator struct {
	doc           *lpa115_types.LPA115Document
	baseValidator *parser.BaseValidator
}

func NewValidator() *Validator {
	return &Validator{
		doc: &lpa115_types.LPA115Document{},
	}
}

func (v *Validator) Setup(doc any) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	var ok bool
	v.doc, ok = doc.(*lpa115_types.LPA115Document)
	if !ok {
		return fmt.Errorf("invalid document type, expected *lpa115_types.LPA115Document")
	}

	v.baseValidator = parser.NewBaseValidator(v.doc)
	return nil
}

func (v *Validator) Validate() []string {
	if len(v.doc.Page3) > 0 {
		for _, part := range v.doc.Page3 {
			if _, err := util.ParseDate(part.PartA.OptionA.Date, ""); err != nil {
				v.baseValidator.AddValidatorErrorMessage(
					fmt.Sprintf("Failed to parse Page 3 Option A date: %v", err),
				)
			}

			if _, err := util.ParseDate(part.PartA.OptionB.Date, ""); err != nil {
				v.baseValidator.AddValidatorErrorMessage(
					fmt.Sprintf("Failed to parse Page 3 Option B date: %v", err),
				)
			}
		}
	}

	if len(v.doc.Page6) > 0 {
		for _, part := range v.doc.Page6 {
			if _, err := util.ParseDate(part.PartB.Date, ""); err != nil {
				v.baseValidator.AddValidatorErrorMessage(
					fmt.Sprintf("Failed to parse Page 6 Part B date: %v", err),
				)
			}
		}
	}

	return v.baseValidator.GetValidatorErrorMessages()
}
