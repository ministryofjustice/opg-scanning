package lp1h_parser

import (
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1h_types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type Validator struct {
	doc             *lp1h_types.LP1HDocument
	commonValidator *parser.Validator
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
	v.commonValidator = parser.NewValidator(v.doc)

	return nil
}

func (v *Validator) Validate() error {
	// Common witness validations
	v.commonValidator.WitnessSignatureFullNameAddressValidator("Page10", "Section9")

	// Section validations
	v.commonValidator.ValidateSignatureDate("Page11", "Section10", "")

	// Iterate over each instance of Page12 (since it's an array) and validate them individually
	for i := range v.doc.Page12 {
		v.commonValidator.WitnessSignatureFullNameAddressValidator(fmt.Sprintf("Page12[%d]", i), "Section11")
		v.commonValidator.ValidateSignatureDate(fmt.Sprintf("Page12[%d]", i), "Section11", "Attorney")
	}

	// Applicant validation iterations for Page20
	for i := range v.doc.Page20 {
		v.commonValidator.ApplicantSignatureValidator(fmt.Sprintf("Page20[%d]", i))
	}

	// LP1H specific validation
	err := v.donorSignatureDateValidator()
	if err != nil {
		v.commonValidator.AddValidatorErrorMessage(err.Error())
	}

	// Return errors if any
	if messages := v.commonValidator.GetValidatorErrorMessages(); len(messages) > 0 {
		return fmt.Errorf("failed to validate LP1H document: %v", messages)
	}

	return nil
}

// Helper function to extract date from a given section.
func (v *Validator) extractDate(page, section, path string) (*time.Time, error) {
	fields, err := v.commonValidator.GetFieldByPath(page, section, path, "DOB")
	if err != nil || len(fields) == 0 {
		return nil, err
	}

	dateStr, ok := fields[0].(string)
	if !ok || dateStr == "" {
		return nil, fmt.Errorf("invalid date format or empty value")
	}

	date, err := util.ParseDate(dateStr, "02012006")
	if err != nil {
		return nil, err
	}

	return &date, nil
}

// Helper function to extract the value of the signature.
func (v *Validator) extractSignature(page string, section string, path string) (string, error) {
	fields, err := v.commonValidator.GetFieldByPath(page, section, path, "Signature")
	if err != nil || len(fields) == 0 {
		return "", err
	}

	signature, ok := fields[0].(string)
	if !ok {
		return "", fmt.Errorf("invalid signature format")
	}

	return signature, nil
}

func (v *Validator) donorSignatureDateValidator() error {
	// Option A validation
	signatureDateA, err := v.extractDate("Page6", "Section5", "OptionA")
	if err != nil {
		return fmt.Errorf("failed to extract DOB OptionA: %v", err)
	}
	section9Date, err := v.extractDate("Page10", "Section9", "Donor")
	if err != nil {
		return fmt.Errorf("failed to extract DOB from Donor: %v", err)
	}
	// Ensure that Section 9 date is after Section 5 date
	if section9Date.After(*signatureDateA) {
		return fmt.Errorf("section 9 donor signature date is not after Section 5's donor signature date")
	}

	signatureA, err := v.extractSignature("Page6", "Section5", "OptionA")
	if err != nil {
		return fmt.Errorf("failed to extract signature from OptionA: %v", err)
	}

	// Option A must have both the date and signature
	if signatureDateA != nil && signatureA != "false" {
		return nil
	}

	// Option B validation
	signatureDateB, err := v.extractDate("Page6", "Section5", "OptionB")
	if err != nil {
		return fmt.Errorf("failed to extract date from OptionB: %v", err)
	}

	signatureB, err := v.extractSignature("Page6", "Section5", "OptionB")
	if err != nil {
		return fmt.Errorf("failed to extract signature from OptionB: %v", err)
	}

	// Option B must have both the date and signature
	if signatureDateB != nil && signatureB != "false" {
		return nil
	}

	// If neither option has both date and signature, return an error
	return fmt.Errorf("donor signature and date are not set in either Option A or Option B")
}
