package lp1f_parser

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
	lp1f_types "github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
)

type Sanitizer struct {
	doc *lp1f_types.LP1FDocument
}

func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		doc: &lp1f_types.LP1FDocument{},
	}
}

func (v *Sanitizer) Setup(doc interface{}) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.doc = doc.(*lpf1_types.LP1FDocument)

	return nil
}

func (s *Sanitizer) Sanitize() (interface{}, error) {
	// Sanitize the entire struct dynamically
	if err := s.sanitizeStruct(s.doc); err != nil {
		return nil, err
	}

	return s.doc, nil
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
