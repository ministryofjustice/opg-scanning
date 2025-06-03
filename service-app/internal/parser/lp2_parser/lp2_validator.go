package lp2_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/date"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp2_types"
)

type Validator struct {
	doc           *lp2_types.LP2Document
	baseValidator *parser.BaseValidator
}

func NewValidator() *Validator {
	return &Validator{
		doc: &lp2_types.LP2Document{},
	}
}

func (v *Validator) Setup(doc any) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.doc = doc.(*lp2_types.LP2Document)
	v.baseValidator = parser.NewBaseValidator(v.doc)

	return nil
}

func (v *Validator) Validate() []string {
	// Validate LP2 Sub-type Selection
	isPF, err := v.baseValidator.GetFieldByPath("Page1", "Section1", "PropertyFinancialAffairs")
	if err != nil {
		v.baseValidator.AddValidatorErrorMessage(fmt.Sprintf("Error reading PropertyFinancialAffairs: %v", err))
	}
	isHW, err := v.baseValidator.GetFieldByPath("Page1", "Section1", "HealthWelfare")
	if err != nil {
		v.baseValidator.AddValidatorErrorMessage(fmt.Sprintf("Error reading HealthWelfare: %v", err))
	}
	// Exactly one must be true.
	if isPF[0].(bool) == isHW[0].(bool) {
		if isHW[0].(bool) {
			v.baseValidator.AddValidatorErrorMessage("Both LPA sub-types are selected")
		} else {
			v.baseValidator.AddValidatorErrorMessage("Neither LPA sub-type is selected")
		}
	}

	// Validate Attorney Signature Dates
	for _, attorney := range v.doc.Page5.Section5.Attorney {
		if attorney.Date == "" {
			continue
		}
		if _, err := date.Parse(attorney.Date); err != nil {
			v.baseValidator.AddValidatorErrorMessage("Failed to parse attorney signature date: " + err.Error())
		}
	}

	// Validate Attorney Date of Birth
	for _, attorney := range v.doc.Page2.Section2.Attorney {
		if attorney.DOB == "" {
			continue
		}
		if _, err := date.Parse(attorney.DOB); err != nil {
			v.baseValidator.AddValidatorErrorMessage("Failed to parse attorney date of birth: " + err.Error())
		}
	}

	return v.baseValidator.GetValidatorErrorMessages()
}
