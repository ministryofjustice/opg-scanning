package lp1f_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
)

type Validator struct {
	doc           *lp1f_types.LP1FDocument
	baseValidator *parser.BaseValidator
}

func NewValidator() *Validator {
	return &Validator{
		doc: &lp1f_types.LP1FDocument{},
	}
}

func (v *Validator) Setup(doc any) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.doc = doc.(*lp1f_types.LP1FDocument)
	v.baseValidator = parser.NewBaseValidator(v.doc)

	return nil
}

func (v *Validator) Validate() []string {
	// Common witness validations
	v.baseValidator.WitnessSignatureFullNameAddressValidator("Page10", "Section9")

	// Section validations
	v.baseValidator.ValidateSignatureDate("Page10", "Section9", "Donor")
	v.baseValidator.ValidateSignatureDate("Page11", "Section10", "")

	// Iterate over each instance of Page12 (since its an array)
	// and validate them individually
	for i := range v.doc.Page12 {
		v.baseValidator.WitnessSignatureFullNameAddressValidator(fmt.Sprintf("Page12[%d]", i), "Section11")
		v.baseValidator.ValidateSignatureDate(fmt.Sprintf("Page12[%d]", i), "Section11", "Attorney")
	}

	// Applicant validation iterations
	for i := range v.doc.Page20 {
		v.baseValidator.ApplicantSignatureValidator(fmt.Sprintf("Page20[%d]", i))
	}

	return v.baseValidator.GetValidatorErrorMessages()
}
