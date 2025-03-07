package parser

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

// DocumentParser unmarshals XML into doc, validates required fields,
// and checks for unexpected XML elements.
func DocumentParser(data []byte, doc interface{}) (interface{}, error) {
	// Unmarshal the XML into the provided document.
	if err := xml.Unmarshal(data, doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func BaseParserXml(data []byte) (*types.BaseSet, error) {
	var parsed types.BaseSet
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("invalid base XML format: %w", err)
	}

	// Validate required fields based on "required" struct tags
	// and check for unexpected XML elements.
	if err := ValidateStruct(&parsed); err != nil {
		return nil, err
	}

	return &parsed, nil
}

// ValidateStruct checks if the provided struct or its nested structs
// have all fields marked with the "required" tag present and non-empty.
func ValidateStruct(s any) error {
	val := reflect.ValueOf(s)

	// If its a pointer, dereference it.
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	// If its a slice or array, iterate over each element.
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		for i := range val.Len() {
			if err := ValidateStruct(val.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	}
	// If its not a struct, nothing to validate.
	if val.Kind() != reflect.Struct {
		return nil
	}
	typeOfS := val.Type()

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typeOfS.Field(i)
		requiredTag := fieldType.Tag.Get("required")

		// Check if the field is a slice of types.AnyElement.
		if field.Type() == reflect.TypeOf([]types.AnyElement{}) {
			// Look for the ",any" option in the tag.
			if tag, ok := fieldType.Tag.Lookup("xml"); ok && strings.Contains(tag, ",any") {
				if field.Len() > 0 {
					return fmt.Errorf("unexpected XML elements found in field %s: %v", fieldType.Name, field.Interface())
				}
			}
		}

		// Recursively validate nested structs.
		if field.Kind() == reflect.Struct {
			if err := ValidateStruct(field.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// If a field is marked as required but is set to its zero value, return an error.
		if requiredTag == "true" {
			if isZeroOfUnderlyingType(field) {
				return fmt.Errorf("validation error: field %s is required but is missing or empty", fieldType.Name)
			}
		}
	}
	return nil
}

// isZeroOfUnderlyingType checks if a field is set to its zero value.
func isZeroOfUnderlyingType(field reflect.Value) bool {
	return field.Interface() == reflect.Zero(field.Type()).Interface()
}
