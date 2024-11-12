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

func (v *CommonValidator) WitnessSignatureFullNameAddressValidator(page string, section string) bool {
	v.errorMessages = []string{}

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

func (v *CommonValidator) formHasWitnessSignature(page, section string) bool {
	signature, err := v.GetFieldByPath(page, section, "Witness", "Signature")
	if err != nil || len(signature) == 0 || signature[0] != true {
		return false
	}
	return true
}

func (v *CommonValidator) formHasWitnessFullName(page, section string) bool {
	fullName, err := v.GetFieldByPath(page, section, "Witness", "FullName")
	if err != nil || len(fullName) == 0 || fullName[0] == "" {
		return false
	}
	return true
}

func (v *CommonValidator) formHasWitnessAddress(page, section string) bool {
	addressLine1, err1 := v.GetFieldByPath(page, section, "Witness", "Address", "Address1")
	postcode, err2 := v.GetFieldByPath(page, section, "Witness", "Address", "Postcode")
	if err1 != nil || err2 != nil || len(addressLine1) == 0 || addressLine1[0] == "" || len(postcode) == 0 || postcode[0] == "" {
		return false
	}
	return true
}

func (v *CommonValidator) AddValidatorErrorMessage(msg string) {
	v.errorMessages = append(v.errorMessages, msg)
}

// Uses reflection to dynamically access nested struct fields and handles array/slice
func (v *CommonValidator) GetFieldByPath(page, section string, fields ...string) ([]interface{}, error) {
	current := reflect.ValueOf(v.doc).Elem()

	for _, field := range append([]string{page, section}, fields...) {
		if field == "" {
			continue
		}

		current = current.FieldByName(field)
		if !current.IsValid() {
			return nil, fmt.Errorf("field %s does not exist in path %v", field, fields)
		}
		// Check if current field is nil for pointer types
		if current.Kind() == reflect.Ptr && current.IsNil() {
			return nil, fmt.Errorf("field %s in path %v is nil", field, fields)
		}
		// Dereference pointer fields
		if current.Kind() == reflect.Ptr {
			current = current.Elem()
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

func (v *CommonValidator) GetValidatorErrorMessages() []string {
	errorMessages := []string{}
	for _, msg := range v.errorMessages {
		errorMessages = append(errorMessages, msg+"\n")
	}
	return errorMessages
}
