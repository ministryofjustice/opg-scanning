package lp1f_parser

import (
	"errors"
	"reflect"
	"strings"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type Sanitizer struct{}

func NewSanitizer() types.Sanitizer {
	return &Sanitizer{}
}

func (s *Sanitizer) Sanitize(doc interface{}) (interface{}, error) {
	if doc == nil {
		return nil, errors.New("cannot sanitize nil data")
	}

	// Sanitize the entire struct dynamically
	if err := s.sanitizeStruct(doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *Sanitizer) sanitizeStruct(input interface{}) error {
	val := reflect.ValueOf(input)

	// Check if val is a pointer and dereference it
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errors.New("input must be a struct or pointer to struct")
	}

	typeOfS := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typeOfS.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		sanitizeTag := fieldType.Tag.Get("sanitize")
		if sanitizeTag == "false" {
			continue
		}

		// Handle nested structs by recursively sanitizing them
		if field.Kind() == reflect.Struct {
			if err := s.sanitizeStruct(field.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// Sanitize string fields
		if field.Kind() == reflect.String {
			sanitizedValue := sanitizeString(field.String())
			field.SetString(sanitizedValue)
		}
	}

	return nil
}

func sanitizeString(input string) string {
	// Sanitization logic here, e.g., trimming spaces, removing special characters, etc.
	return strings.TrimSpace(input)
}
