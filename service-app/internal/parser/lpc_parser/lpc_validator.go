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
	return &Validator{
		doc: &lpc_types.LPCDocument{},
	}
}

func (v *Validator) Setup(doc interface{}) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.doc = doc.(*lpc_types.LPCDocument)
	v.baseValidator = parser.NewBaseValidator(v.doc)

	return nil
}

func (v *Validator) Validate() error {
	// Common witness validations
	v.baseValidator.WitnessSignatureFullNameAddressValidator("Page10", "Section9")

	// Section validations
	v.baseValidator.ValidateSignatureDate("Page11", "Section10", "")

	// TODO add more validation for LPC

	// Return errors if any
	if messages := v.baseValidator.GetValidatorErrorMessages(); len(messages) > 0 {
		return fmt.Errorf("failed to validate LP1H document: %v", messages)
	}

	return nil
}
