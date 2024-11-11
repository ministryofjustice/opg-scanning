package parser

import (
	"errors"
	"fmt"
	"reflect"
)

type CommonValidator struct {
	doc           interface{}
	errorMessages []string
}

func NewCommonValidator(doc interface{}) *CommonValidator {
	return &CommonValidator{
		doc:           doc,
		errorMessages: []string{},
	}
}

func (v *CommonValidator) Validate(page string, section string) bool {
	v.errorMessages = []string{}

	if !v.formHasWitnessSignature(page, section) {
		v.addValidatorErrorMessage(fmt.Sprintf("Section %s Witness Signature not set.", section))
	}

	if !v.formHasWitnessFullName(page, section) {
		v.addValidatorErrorMessage(fmt.Sprintf("Section %s Witness Full Name not set.", section))
	}

	if !v.formHasWitnessAddress(page, section) {
		v.addValidatorErrorMessage(fmt.Sprintf("Section %s Witness Address not valid.", section))
	}

	return len(v.errorMessages) == 0
}

func (v *CommonValidator) formHasWitnessSignature(page, section string) bool {
	signature, err := v.getFieldByPath(page, section, "Witness", "Signature")
	if err != nil || signature != "true" {
		return false
	}
	return true
}

func (v *CommonValidator) formHasWitnessFullName(page, section string) bool {
	fullName, err := v.getFieldByPath(page, section, "Witness", "FullName")
	if err != nil || fullName == "" {
		return false
	}
	return true
}

func (v *CommonValidator) formHasWitnessAddress(page, section string) bool {
	addressLine1, err1 := v.getFieldByPath(page, section, "Witness", "Address", "Address1")
	postcode, err2 := v.getFieldByPath(page, section, "Witness", "Address", "Postcode")
	if err1 != nil || err2 != nil || addressLine1 == "" || postcode == "" {
		return false
	}
	return true
}

func (v *CommonValidator) addValidatorErrorMessage(msg string) {
	v.errorMessages = append(v.errorMessages, msg)
}

// Uses reflection to dynamically access nested struct fields
func (v *CommonValidator) getFieldByPath(fields ...string) (string, error) {
	current := reflect.ValueOf(v.doc).Elem()

	for _, field := range fields {
		current = current.FieldByName(field)
		if !current.IsValid() {
			return "", fmt.Errorf("field %s does not exist in path %v", field, fields)
		}
		// Check if current field is nil for pointer types
		if current.Kind() == reflect.Ptr && current.IsNil() {
			return "", fmt.Errorf("field %s in path %v is nil", field, fields)
		}
		// Dereference pointer fields
		if current.Kind() == reflect.Ptr {
			current = current.Elem()
		}
	}

	// Convert final field value to string if it's a string type
	if current.Kind() == reflect.String {
		return current.String(), nil
	} else if current.Kind() == reflect.Bool {
		return fmt.Sprintf("%t", current.Bool()), nil
	}
	return "", errors.New("field is neither a string nor a bool")
}

func (v *CommonValidator) GetValidatorErrorMessages() []string {
	return v.errorMessages
}
