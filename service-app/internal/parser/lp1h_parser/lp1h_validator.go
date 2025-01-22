package lp1h_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser/lp1f_parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1h_types"
)

type Validator struct {
	// We use LP1F validator to validate the common fields
	lp1f_parser.Validator
	doc *lp1h_types.LP1HDocument
}

func NewValidator() *Validator {
	return &Validator{
		doc: &lp1h_types.LP1HDocument{},
	}
}

func (v *Validator) Setup(doc interface{}) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.doc = doc.(*lp1h_types.LP1HDocument)
	v.Validator.Setup(v.doc)

	return nil
}

func (v *Validator) Validate() error {
	// Validate the common LP1F fields
	if err := v.Validator.Validate(); err != nil {
		return err
	}

	if v.doc.XMLName.Local != "LP1H" {
		return fmt.Errorf("should be LP1H")
	}

	return nil
}
