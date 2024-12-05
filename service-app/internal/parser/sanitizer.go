package parser

import (
	"errors"
	"reflect"
	"strings"
)

type Sanitizer struct {
	doc *interface{}
}

func NewSanitizer(doc interface{}) *Sanitizer {
	return &Sanitizer{
		doc: &doc,
	}
}

// SanitizeStruct recursively sanitizes a struct, sanitizing string fields and
// recursively sanitizing nested structs. The function uses reflection to access
// the fields of the struct and sanitizes only the exported fields. If a field
// has a 'sanitize' tag set to 'false', the field is skipped.
func (s *Sanitizer) SanitizeStruct(input interface{}) error {
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
			if err := s.SanitizeStruct(field.Addr().Interface()); err != nil {
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
