package parser

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Validator struct {
	doc           interface{}
	errorMessages []string
}

func NewValidator(doc interface{}) *Validator {
	return &Validator{
		doc:           doc,
		errorMessages: []string{},
	}
}

func (v *Validator) WitnessSignatureFullNameAddressValidator(page string, section string) bool {
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

func (v *Validator) formHasWitnessSignature(page, section string) bool {
	signature, err := v.GetFieldByPath(page, section, "Witness", "Signature")
	if err == nil && signature[0].(bool) {
		return true
	}
	return false
}

func (v *Validator) formHasWitnessFullName(page, section string) bool {
	fullName, err := v.GetFieldByPath(page, section, "Witness", "FullName")
	if err == nil && fullName[0] != "" {
		return true
	}
	return false
}

func (v *Validator) formHasWitnessAddress(page, section string) bool {
	addressLine1, err1 := v.GetFieldByPath(page, section, "Witness", "Address", "Address1")
	postcode, err2 := v.GetFieldByPath(page, section, "Witness", "Address", "Postcode")

	// Check that addressLine1 and postcode contain non-empty strings
	if err1 == nil && err2 == nil && addressLine1[0] != "" && postcode[0] != "" {
		return true
	}

	return false
}

func (v *Validator) AddValidatorErrorMessage(msg string) {
	v.errorMessages = append(v.errorMessages, msg)
}

// GetFieldByPath retrieves a field by its path in the document.
// The path is a set of strings, where each string is a field name or a field name with an optional index.
// For example, "Page1[0].Section1.Witness.FullName" is a valid path.
// If any part of the path does not exist, an error is returned.
// The function returns a slice of interfaces containing the values of the field.
// If the field is a slice or array, the function returns a slice of interfaces containing the elements of the slice or array.
// If the field is a string, bool, or unsupported type, the function returns an error.
func (v *Validator) GetFieldByPath(page, section string, fields ...string) ([]interface{}, error) {
	current := reflect.ValueOf(v.doc).Elem()

	// Start navigation through the fields path
	for _, field := range append([]string{page, section}, fields...) {
		if field == "" {
			continue
		}

		fieldName, index, err := parseFieldWithIndex(field)
		if err != nil {
			return nil, err
		}

		// Navigate to the field by name
		current = current.FieldByName(fieldName)
		if !current.IsValid() {
			return nil, fmt.Errorf("field %s does not exist in path %v", field, fields)
		}

		// Dereference pointers
		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				return nil, fmt.Errorf("field %s in path %v is nil", field, fields)
			}
			current = current.Elem()
		}

		// Handle slices/arrays with optional index
		if current.Kind() == reflect.Slice || current.Kind() == reflect.Array {
			if index == nil {
				continue
			}

			// If index is provided, access the specific element
			if *index < 0 || *index >= current.Len() {
				return nil, fmt.Errorf("index %d out of bounds for field %s", *index, field)
			}
			current = current.Index(*index)
		}
	}

	// Return the final field value as a slice of interfaces
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

// Handles field names with optional indices e.g. "Page12[0]"
func parseFieldWithIndex(field string) (string, *int, error) {
	if !strings.Contains(field, "[") {
		return field, nil, nil // No index specified
	}

	// Extract the field name and index
	name := field[:strings.Index(field, "[")]
	indexStr := field[strings.Index(field, "[")+1 : strings.Index(field, "]")]

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return "", nil, fmt.Errorf("invalid index in field %s: %v", field, err)
	}

	return name, &index, nil
}

func (v *Validator) GetValidatorErrorMessages() []string {
	errorMessages := []string{}
	for _, msg := range v.errorMessages {
		errorMessages = append(errorMessages, msg+"\n")
	}
	return errorMessages
}
