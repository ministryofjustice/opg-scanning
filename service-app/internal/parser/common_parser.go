package parser

import (
	"encoding/xml"
	"fmt"
	"reflect"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

func DocumentParser(data []byte, doc interface{}) (interface{}, error) {
	if err := xml.Unmarshal(data, doc); err != nil {
		return nil, err
	}

	// Validate required fields based on struct tags
	if err := ValidateStruct(doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func BaseParserXml(data []byte) (*types.BaseSet, error) {
	var parsed types.BaseSet
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("invalid base XML format: %w", err)
	}

	return &parsed, nil
}

// ValidateStruct checks if the provided struct or its nested structs
// have all fields marked with the "required" tag present and non-empty.
// It supports pointer dereferencing and recursive validation for nested structs.
// Returns an error if any required field is missing or empty; otherwise, returns nil.
func ValidateStruct(s any) error {
	val := reflect.ValueOf(s)

	// Check if val is a pointer and dereference it
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typeOfS := val.Type()

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typeOfS.Field(i)
		requiredTag := fieldType.Tag.Get("required")

		// Handle nested structs by recursively validating them
		if field.Kind() == reflect.Struct {
			if err := ValidateStruct(field.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// If the field has a required tag and it's true, validate presence
		if requiredTag == "true" {
			if isZeroOfUnderlyingType(field) {
				return fmt.Errorf("validation error: field %s is required but is missing or empty", fieldType.Name)
			}
		}
	}
	return nil
}

// Checks if a field is set to its zero value
func isZeroOfUnderlyingType(field reflect.Value) bool {
	return field.Interface() == reflect.Zero(field.Type()).Interface()
}
