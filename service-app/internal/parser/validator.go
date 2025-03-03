package parser

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/require"
)

// Validator is a struct that holds document data and validation error messages
type BaseValidator struct {
	doc           interface{}
	errorMessages []string
	dates         []time.Time
}

// NewValidator creates a new instance of Validator
func NewBaseValidator(doc interface{}) *BaseValidator {
	return &BaseValidator{
		doc:           doc,
		errorMessages: []string{},
	}
}

// AddValidatorErrorMessage adds an error message to the validator's list of error messages
func (v *BaseValidator) AddValidatorErrorMessage(msg string) {
	v.errorMessages = append(v.errorMessages, msg)
}

// GetValidatorErrorMessages returns a copy of the validator's error messages
func (v *BaseValidator) GetValidatorErrorMessages() []string {
	errorMessages := []string{}
	return append(errorMessages, v.errorMessages...)
}

// Helper functions for working with field values

// GetFieldByPath retrieves a field by its path in the document
func (v *BaseValidator) GetFieldByPath(page string, section string, fields ...string) ([]interface{}, error) {
	current := reflect.ValueOf(v.doc).Elem()

	for _, field := range append([]string{page, section}, fields...) {
		if field == "" {
			continue
		}

		fieldName, index, err := parseFieldWithIndex(field)
		if err != nil {
			return nil, err
		}

		current = current.FieldByName(fieldName)
		if !current.IsValid() {
			return nil, fmt.Errorf("field %s does not exist in path %v", field, fields)
		}

		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				return nil, fmt.Errorf("field %s in path %v is nil", field, fields)
			}
			current = current.Elem()
		}

		if current.Kind() == reflect.Slice || current.Kind() == reflect.Array {
			if index == nil {
				continue
			}

			if *index < 0 || *index >= current.Len() {
				return nil, fmt.Errorf("index %d out of bounds for field %s", *index, field)
			}
			current = current.Index(*index)
		}
	}

	switch current.Kind() {
	case reflect.String:
		return []interface{}{current.String()}, nil
	case reflect.Bool:
		return []interface{}{current.Bool()}, nil
	case reflect.Slice, reflect.Array:
		var result []interface{}
		for i := 0; i < current.Len(); i++ {
			result = append(result, current.Index(i).Interface())
		}
		return result, nil
	default:
		return nil, errors.New("unsupported field type")
	}
}

// Helper function to handle field names with optional indices e.g. "Page12[0]"
func parseFieldWithIndex(field string) (string, *int, error) {
	if !strings.Contains(field, "[") {
		return field, nil, nil // No index specified
	}

	name := field[:strings.Index(field, "[")]
	indexStr := field[strings.Index(field, "[")+1 : strings.Index(field, "]")]

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return "", nil, fmt.Errorf("invalid index in field %s: %v", field, err)
	}

	return name, &index, nil
}

// Validation functions for form fields

// Validates the presence of witness signature, full name, and address
func (v *BaseValidator) WitnessSignatureFullNameAddressValidator(page string, section string) bool {
	if !v.formHasWitnessSignature(page, section) {
		v.AddValidatorErrorMessage(fmt.Sprintf("%s %s Witness Signature not set.", page, section))
	}

	if !v.formHasWitnessFullName(page, section) {
		v.AddValidatorErrorMessage(fmt.Sprintf("%s %s Witness Full Name not set.", page, section))
	}

	if !v.formHasWitnessAddress(page, section) {
		v.AddValidatorErrorMessage(fmt.Sprintf("%s %s Witness Address not valid.", page, section))
	}

	return len(v.errorMessages) == 0
}

// Validates the presence and format of a signature date for a specific section
func (v *BaseValidator) ValidateSignatureDate(page, section, field string) {
	dateStr, err := v.getFieldValues(page, section, field)
	if err != nil {
		v.AddValidatorErrorMessage(err.Error())
		return
	}

	if date, err := validateSignatureDate(dateStr, field); err != nil {
		v.AddValidatorErrorMessage(err.Error())
	} else {
		v.dates = append(v.dates, date)
	}
}

// ApplicantSignatureValidator gathers applicant signature dates and validates them
func (v *BaseValidator) ApplicantSignatureValidator(page string) []time.Time {
	var applicantSignatureDates []time.Time

	// Retrieve and validate applicant signature dates
	applicants, err := v.GetFieldByPath(page, "Section15", "Applicant")
	if err != nil {
		v.AddValidatorErrorMessage(fmt.Sprintf("failed to retrieve applicant data: %v", err))
		return nil
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
			v.AddValidatorErrorMessage("applicant date is missing or invalid")
			continue
		}

		dateStr := dateField.String()
		signatureDate, err := util.ParseDate(dateStr, "")
		if err != nil {
			v.AddValidatorErrorMessage("applicant date is invalid")
			continue
		}

		applicantSignatureDates = append(applicantSignatureDates, signatureDate)
	}

	// Ensure applicant dates follow correct ordering rules
	if len(applicantSignatureDates) == 0 {
		v.AddValidatorErrorMessage("no valid applicant signature/dates found")
	} else {
		v.checkDatesAgainstEarliestApplicantDate(applicantSignatureDates)
	}

	return applicantSignatureDates
}

// Checks if the date string is valid and not in the future
func validateSignatureDate(dateStr, label string) (time.Time, error) {
	parsedDate, err := util.ParseDate(dateStr, "")
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s date format: %w", label, err)
	}

	if parsedDate.After(time.Now()) {
		return time.Time{}, fmt.Errorf("%s date cannot be in the future", label)
	}

	return parsedDate, nil
}

// Helper function to check if all form dates are before the earliest applicant signature date
func (v *BaseValidator) checkDatesAgainstEarliestApplicantDate(applicantSignatureDates []time.Time) {
	earliestDate := getEarliestDate(applicantSignatureDates)
	for _, date := range v.dates {
		if date.After(earliestDate) {
			v.AddValidatorErrorMessage("all form dates must be before the earliest applicant signature date")
			return
		}
	}
}

// Helper function to retrieve the earliest date from a list of dates
func getEarliestDate(dates []time.Time) time.Time {
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

// Helper functions for checking specific fields in the form

// formHasWitnessSignature checks if the witness signature field is present and valid
func (v *BaseValidator) formHasWitnessSignature(page, section string) bool {
	signature, err := v.GetFieldByPath(page, section, "Witness", "Signature")
	if err == nil && signature[0].(bool) {
		return true
	}
	return false
}

// formHasWitnessFullName checks if the witness full name field is present and valid
func (v *BaseValidator) formHasWitnessFullName(page, section string) bool {
	fullName, err := v.GetFieldByPath(page, section, "Witness", "FullName")
	if err == nil && fullName[0] != "" {
		return true
	}
	return false
}

// formHasWitnessAddress checks if the witness address fields are present and valid
func (v *BaseValidator) formHasWitnessAddress(page, section string) bool {
	addressLine1, err1 := v.GetFieldByPath(page, section, "Witness", "Address", "Address1")
	postcode, err2 := v.GetFieldByPath(page, section, "Witness", "Address", "Postcode")

	if err1 == nil && err2 == nil && addressLine1[0] != "" && postcode[0] != "" {
		return true
	}

	return false
}

// Retrieves and validates the signature and date fields for a section
func (v *BaseValidator) getFieldValues(page, section, field string) (string, error) {
	signatureVal, err := v.GetFieldByPath(page, section, field, "Signature")
	if err != nil || !signatureVal[0].(bool) {
		return "", fmt.Errorf("%s %s %s signature not set or invalid", page, section, field)
	}

	dateVal, err := v.GetFieldByPath(page, section, field, "Date")
	dateStr := dateVal[0].(string)
	if err != nil || dateStr == "" {
		return "", fmt.Errorf("missing %s %s %s date", page, section, field)
	}

	return dateStr, nil
}

func TestHelperDocumentValidation(
	t *testing.T,
	fileName string,
	expectError bool,
	expectedPatterns []string,
	validator CommonValidator,
) {

	err := validator.Validate()

	if expectError {
		require.Error(t, err, "Expected validation errors for %s, but got none", fileName)

		// Safely assert that the validator is of the correct type
		messages := validator.GetValidatorErrorMessages()

		t.Log("Actual messages from validation:")
		for _, msg := range messages {
			t.Log(msg)
		}

		for _, pattern := range expectedPatterns {
			regex, compErr := regexp.Compile(pattern)
			require.NoError(t, compErr, "Failed to compile regex for pattern: %s", pattern)

			found := false
			for _, msg := range messages {
				if regex.MatchString(msg) {
					found = true
					break
				}
			}

			require.True(t, found, "Expected error message pattern not found: %s", pattern)
		}
	} else {
		require.NoError(t, err, "Expected no errors for valid document XML")
	}
}
