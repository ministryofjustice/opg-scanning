package lpc_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lpc_types"
)

type Validator struct {
	doc           *lpc_types.LPCDocument
	baseValidator *parser.BaseValidator
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Setup(doc any) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	castDoc, ok := doc.(*lpc_types.LPCDocument)
	if !ok {
		return fmt.Errorf("failed to cast document to LPCDocument")
	}
	v.doc = castDoc
	v.baseValidator = parser.NewBaseValidator(v.doc)
	return nil
}

func (v *Validator) Validate() []string {
	v.validatePage1()
	v.validatePage3()
	v.validatePage4()

	return v.baseValidator.GetValidatorErrorMessages()
}

func (v *Validator) validatePage1() {
	for i, p := range v.doc.Page1 {
		cs := p.ContinuationSheet1

		if len(cs.Attorneys) < 2 {
			v.baseValidator.AddValidatorErrorMessage(
				fmt.Sprintf("Page1[%d] requires exactly 2 Attorney blocks, found %d", i, len(cs.Attorneys)),
			)
		}
	}
}

func (v *Validator) validatePage3() {
	for i, p := range v.doc.Page3 {
		cs := p.ContinuationSheet3

		if len(cs.Witnesses) < 2 {
			v.baseValidator.AddValidatorErrorMessage(
				fmt.Sprintf("Page3[%d] requires exactly 2 Witness blocks, found %d", i, len(cs.Witnesses)),
			)
		}
	}
}

func (v *Validator) validatePage4() {
	for i, p := range v.doc.Page4 {
		cs := p.ContinuationSheet4

		if len(cs.AuthorisedPerson) < 2 {
			v.baseValidator.AddValidatorErrorMessage(
				fmt.Sprintf("Page4[%d] requires exactly 2 AuthorisedPerson blocks, found %d", i, len(cs.AuthorisedPerson)),
			)
		}
	}
}
