package lp1f_parser

import (
	"errors"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	lp1f_types "github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
)

type Validator struct {
	doc *lp1f_types.LP1FDocument
}

func NewValidator(doc *lp1f_types.LP1FDocument) types.Validator {
	return &Validator{doc: doc}
}

func (v *Validator) Validate() error {
	if err := v.donorSignatureDateValidator(); err != nil {
		return err
	}
	if err := v.certificateProviderSignatureDateValidator(); err != nil {
		return err
	}
	if err := v.attorneySignatureDateValidator(); err != nil {
		return err
	}
	return v.applicantSignatureValidator()
}

func (v *Validator) donorSignatureDateValidator() error {
	if !v.doc.Page10.Section9.Donor.Signature {
		return errors.New("section 9 Donor Signature not set")
	}

	// Parse the signature date
	signatureDate, err := time.Parse("2006-01-02", v.doc.Page10.Section9.Donor.Date)
	if err != nil {
		return errors.New("invalid date format, expected YYYY-MM-DD")
	}

	// Check if the date is not in the future
	if signatureDate.After(time.Now()) {
		return errors.New("signature date cannot be in the future")
	}

	return nil
}

func (v *Validator) certificateProviderSignatureDateValidator() error {
	if !v.doc.Page11.Section10.Signature {
		return errors.New("section 10 Certificate Provider Signature not set")
	}

	// Parse the signed date
	signedDate, err := time.Parse("2006-01-02", v.doc.Page11.Section10.Date)
	if err != nil {
		return errors.New("invalid date format, expected YYYY-MM-DD")
	}

	// Check if the date is not in the future
	if signedDate.After(time.Now()) {
		return errors.New("signed date cannot be in the future")
	}

	return nil
}

func (v *Validator) attorneySignatureDateValidator() error {
	if !v.doc.Page12.Section11.Attorney.Signature {
		return errors.New("section 11 Attorney Signature not set")
	}

	// Parse the signature date
	signatureDate, err := time.Parse("2006-01-02", v.doc.Page12.Section11.Attorney.Date)
	if err != nil {
		return errors.New("invalid date format, expected YYYY-MM-DD")
	}

	// Check if the date is not in the future
	if signatureDate.After(time.Now()) {
		return errors.New("signature date cannot be in the future")
	}

	return nil
}

func (v *Validator) applicantSignatureValidator() error {
	for _, applicant := range v.doc.Page20.Section15.Applicant {
		if !applicant.Signature {
			return errors.New("section 15 Applicant Signature not set")
		}

		// Parse the signature date
		signatureDate, err := time.Parse("2006-01-02", applicant.Date)
		if err != nil {
			return errors.New("invalid date format, expected YYYY-MM-DD")
		}

		// Check if the date is not in the future
		if signatureDate.After(time.Now()) {
			return errors.New("signature date cannot be in the future")
		}

	}
	return nil
}
