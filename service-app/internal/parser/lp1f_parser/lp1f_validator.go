package lp1f_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
)

type Validator struct {
	doc             *lp1f_types.LP1FDocument
	commonValidator *parser.Validator
}

func NewValidator() *Validator {
	return &Validator{
		doc: &lp1f_types.LP1FDocument{},
	}
}

func (v *Validator) Setup(doc interface{}) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.doc = doc.(*lp1f_types.LP1FDocument)
	v.commonValidator = parser.NewValidator(v.doc)

	return nil
}

func (v *Validator) Validate() error {
	// Common witness validations
	v.commonValidator.WitnessSignatureFullNameAddressValidator("Page10", "Section9")

	// Section validations
	v.commonValidator.ValidateSection("Page10", "Section9", "Donor")
	v.commonValidator.ValidateSection("Page11", "Section10", "")

	// Iterate over each instance of Page12 (since its an array)
	// and validate them individually
	for i := range v.doc.Page12 {
		v.commonValidator.WitnessSignatureFullNameAddressValidator(fmt.Sprintf("Page12[%d]", i), "Section11")
		v.commonValidator.ValidateSection(fmt.Sprintf("Page12[%d]", i), "Section11", "Attorney")
	}

	// Applicant validation iterations
	for i := range v.doc.Page20 {
		v.commonValidator.ApplicantSignatureValidator(fmt.Sprintf("Page20[%d]", i))
	}

	// Return errors if any
	if messages := v.commonValidator.GetValidatorErrorMessages(); len(messages) > 0 {
		return fmt.Errorf("failed to validate LP1F document: %v", messages)
	}
	return nil
}
