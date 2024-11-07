package lp1f_parser

import (
	"errors"
	"sort"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	lp1f_types "github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type Validator struct {
	doc                     *lp1f_types.LP1FDocument
	dates                   []time.Time
	applicantSignatureDates []time.Time
}

func NewValidator(doc *lp1f_types.LP1FDocument) types.Validator {
	return &Validator{
		doc:                     doc,
		dates:                   []time.Time{},
		applicantSignatureDates: []time.Time{},
	}
}

func (v *Validator) Validate() error {
	// Collect dates from other sections
	if err := v.collectValidDates(); err != nil {
		return err
	}

	// Validate applicant signatures
	return v.applicantSignatureValidator()
}

// Gathers dates from donor, certificate provider, and attorney sections.
func (v *Validator) collectValidDates() error {
	// Donor signature date
	donorDate, err := validateSignatureDate(v.doc.Page10.Section9.Donor.Signature, v.doc.Page10.Section9.Donor.Date, "Donor")
	if err == nil {
		v.dates = append(v.dates, donorDate)
	} else {
		return err
	}

	// Certificate provider signature date
	certProviderDate, err := validateSignatureDate(v.doc.Page11.Section10.Signature, v.doc.Page11.Section10.Date, "Certificate Provider")
	if err == nil {
		v.dates = append(v.dates, certProviderDate)
	} else {
		return err
	}

	// Attorney signature date
	attorneyDate, err := validateSignatureDate(v.doc.Page12.Section11.Attorney.Signature, v.doc.Page12.Section11.Attorney.Date, "Attorney")
	if err == nil {
		v.dates = append(v.dates, attorneyDate)
	} else {
		return err
	}

	return nil
}

// Helper to validate the presence and format of a signature date
func validateSignatureDate(signature bool, dateStr string, label string) (time.Time, error) {
	if !signature {
		return time.Time{}, errors.New(label + " Signature not set")
	}

	parsedDate, err := util.ParseDate(dateStr)
	if err != nil {
		return time.Time{}, errors.New("invalid " + label + " date format, expected multiple formats like DD/MM/YYYY or YYYY-MM-DD")
	}

	if parsedDate.After(time.Now()) {
		return time.Time{}, errors.New(label + " date cannot be in the future")
	}

	return parsedDate, nil
}

// Gathers applicant signature dates and validates them against other form dates.
func (v *Validator) applicantSignatureValidator() error {
	// Collect applicant signature dates
	for _, applicant := range v.doc.Page20.Section15.Applicant {
		if applicant.Signature {
			signatureDate, err := util.ParseDate(applicant.Date)
			if err == nil {
				v.applicantSignatureDates = append(v.applicantSignatureDates, signatureDate)
			}
		}
	}

	// Ensure there is at least one valid applicant signature date
	if len(v.applicantSignatureDates) == 0 {
		return errors.New("no valid applicant signature dates found")
	}

	// Get the earliest applicant signature date
	earliestApplicantSignatureDate, err := v.getEarliestApplicantSignatureDate()
	if err != nil {
		return err
	}

	// Ensure all form dates are before the earliest applicant signature date
	for _, date := range v.dates {
		if date.After(earliestApplicantSignatureDate) {
			return errors.New("all form dates must be before the earliest applicant signature date")
		}
	}

	return nil
}

// Returns the earliest date among applicant signature dates.
func (v *Validator) getEarliestApplicantSignatureDate() (time.Time, error) {
	if len(v.applicantSignatureDates) == 0 {
		return time.Time{}, errors.New("no applicant signature dates available")
	}

	// Sort dates to find the earliest
	sort.Slice(v.applicantSignatureDates, func(i, j int) bool {
		return v.applicantSignatureDates[i].Before(v.applicantSignatureDates[j])
	})

	return v.applicantSignatureDates[0], nil
}
