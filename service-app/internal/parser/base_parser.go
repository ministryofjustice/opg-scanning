package parser

import (
	"encoding/xml"
	"fmt"
	"reflect"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type BaseParser interface {
	ParseDocument(data []byte) (interface{}, error)
	ParseImage(data []byte) (interface{}, error)
}

func BaseParserXml(data []byte) (*types.BaseSet, error) {
	var parsed types.BaseSet
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("invalid base XML format: %w", err)
	}

	return &parsed, nil
}

// ValidateStruct checks for required fields based on `required:"true"` tags
func ValidateStruct(s interface{}) error {
	val := reflect.ValueOf(s)

	// Check if val is a pointer and dereference it
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typeOfS := val.Type()

	for i := 0; i < val.NumField(); i++ {
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

// isZeroOfUnderlyingType checks if a field is set to its zero value
func isZeroOfUnderlyingType(field reflect.Value) bool {
	return field.Interface() == reflect.Zero(field.Type()).Interface()
}
