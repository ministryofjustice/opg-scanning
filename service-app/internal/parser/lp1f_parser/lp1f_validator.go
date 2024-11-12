package lp1f_parser

import (
	"fmt"
	"reflect"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	lp1f_types "github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type Validator struct {
	doc                     *lp1f_types.LP1FDocument
	commonValidator         *parser.CommonValidator
	dates                   []time.Time
	applicantSignatureDates []time.Time
}

func NewValidator(doc *lp1f_types.LP1FDocument) types.Validator {
	return &Validator{
		doc:             doc,
		commonValidator: parser.NewCommonValidator(doc),
	}
}

func (v *Validator) Validate() error {
	// Common witness validations
	v.commonValidator.WitnessSignatureFullNameAddressValidator("Page10", "Section9")
	v.commonValidator.WitnessSignatureFullNameAddressValidator("Page12", "Section11")

	// Section validations
	v.validateSection("Page10", "Section9", "Donor")
	v.validateSection("Page11", "Section10", "")
	v.validateSection("Page12", "Section11", "Attorney")

	// Applicant validations
	v.applicantSignatureValidator()

	// Return errors if any
	if messages := v.commonValidator.GetValidatorErrorMessages(); len(messages) > 0 {
		return fmt.Errorf("failed to validate LP1F document: %v", messages)
	}
	return nil
}

// Validates the presence and format of a signature date for a specific section
func (v *Validator) validateSection(page, section, field string) {
	_, dateStr, err := v.getFieldValues(page, section, field)
	if err != nil {
		v.commonValidator.AddValidatorErrorMessage(err.Error())
		return
	}

	if date, err := validateSignatureDate(dateStr, field); err != nil {
		v.commonValidator.AddValidatorErrorMessage(err.Error())
	} else {
		v.dates = append(v.dates, date)
	}
}

// Retrieves and validates the signature and date fields for a section
func (v *Validator) getFieldValues(page, section, field string) (bool, string, error) {
	signatureVal, err := v.commonValidator.GetFieldByPath(page, section, field, "Signature")
	if err != nil || len(signatureVal) == 0 || !signatureVal[0].(bool) {
		return false, "", fmt.Errorf("%s signature not set or invalid", field)
	}

	dateVal, err := v.commonValidator.GetFieldByPath(page, section, field, "Date")
	if err != nil || len(dateVal) == 0 {
		return true, "", fmt.Errorf("missing %s date", field)
	}

	dateStr, ok := dateVal[0].(string)
	if !ok {
		return true, "", fmt.Errorf("invalid %s date format", field)
	}

	return true, dateStr, nil
}

// Checks if the date string is valid and not in the future
func validateSignatureDate(dateStr, label string) (time.Time, error) {
	parsedDate, err := util.ParseDate(dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s date format: %w", label, err)
	}

	if parsedDate.After(time.Now()) {
		return time.Time{}, fmt.Errorf("%s date cannot be in the future", label)
	}

	return parsedDate, nil
}

// Gathers applicant signature dates and validates them
func (v *Validator) applicantSignatureValidator() {
	// Retrieve and validate applicant signature dates
	applicants, err := v.commonValidator.GetFieldByPath("Page20", "Section15", "Applicant")
	if err != nil {
		v.commonValidator.AddValidatorErrorMessage(fmt.Sprintf("failed to retrieve applicant data: %v", err))
		return
	}

	for _, applicant := range applicants {
		applicantVal := reflect.ValueOf(applicant)
		if applicantVal.Kind() == reflect.Ptr {
			applicantVal = applicantVal.Elem()
		}

		signatureField := applicantVal.FieldByName("Signature")
		dateField := applicantVal.FieldByName("Date")

		if !signatureField.IsValid() || !signatureField.Bool() {
			continue
		}
		if !dateField.IsValid() || dateField.Kind() != reflect.String {
			v.commonValidator.AddValidatorErrorMessage("applicant date is missing or invalid")
			continue
		}

		dateStr := dateField.String()
		signatureDate, err := util.ParseDate(dateStr)
		if err != nil {
			v.commonValidator.AddValidatorErrorMessage("applicant date is invalid")
			continue
		}

		v.applicantSignatureDates = append(v.applicantSignatureDates, signatureDate)
	}

	// Ensure applicant dates follow corerct ordering rules
	if len(v.applicantSignatureDates) == 0 {
		v.commonValidator.AddValidatorErrorMessage("no valid applicant signature dates found")
	} else {
		v.checkDatesAgainstEarliestApplicantDate()
	}
}

func (v *Validator) checkDatesAgainstEarliestApplicantDate() {
	earliestDate := v.getEarliestDate(v.applicantSignatureDates)
	for _, date := range v.dates {
		if date.After(earliestDate) {
			v.commonValidator.AddValidatorErrorMessage("all form dates must be before the earliest applicant signature date")
			return
		}
	}
}

func (v *Validator) getEarliestDate(dates []time.Time) time.Time {
	if len(dates) == 0 {
		return time.Time{}
	}
	earliest := dates[0]
	for _, date := range dates[1:] {
		if date.Before(earliest) {
			earliest = date
		}
	}
	return earliest
}
