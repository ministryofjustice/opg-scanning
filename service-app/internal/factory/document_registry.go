package factory

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/corresp_parser"
	"github.com/ministryofjustice/opg-scanning/internal/parser/lp1f_parser"
)

// Component defines a registry entry for a document type.
type Component struct {
	Parser    func([]byte) (interface{}, error)
	Validator parser.CommonValidator
	Sanitizer parser.CommonSanitizer
}

// Defines the behavior for a document registry.
type RegistryInterface interface {
	GetParser(docType string) (func([]byte) (interface{}, error), error)
	GetValidator(docType string) (parser.CommonValidator, error)
	GetSanitizer(docType string) (parser.CommonSanitizer, error)
}

// Registry manages parsers, validators, and sanitizers for document types.
type Registry struct {
	components map[string]Component
}

// Initializes the registry with doc type handlers.
func NewRegistry() *Registry {
	return &Registry{
		components: map[string]Component{
			"LP1F": {
				Parser: func(data []byte) (interface{}, error) {
					return lp1f_parser.Parse(data)
				},
				Validator: lp1f_parser.NewValidator(),
				Sanitizer: lp1f_parser.NewSanitizer(),
			},
			"Correspondence": {
				Parser: func(data []byte) (interface{}, error) {
					return corresp_parser.Parse(data)
				},
				// ValidatorFactory: func(doc interface{}) parser.CommonValidator {
				// 	return corresp_parser.NewValidator(doc.(*corresp_types.Correspondence))
				// },
				// Sanitizer: corresp_parser.NewSanitizer(),
			},
		},
	}
}

func (r *Registry) GetParser(docType string) (func([]byte) (interface{}, error), error) {
	component, exists := r.components[docType]
	if !exists {
		return nil, fmt.Errorf("parser for document type '%s' not found", docType)
	}
	return component.Parser, nil
}

func (r *Registry) GetValidator(docType string) (parser.CommonValidator, error) {
	component, exists := r.components[docType]
	if !exists {
		return nil, fmt.Errorf("validator for document type '%s' not found", docType)
	}
	return component.Validator, nil
}

func (r *Registry) GetSanitizer(docType string) (parser.CommonSanitizer, error) {
	component, exists := r.components[docType]
	if !exists {
		return nil, fmt.Errorf("sanitizer for document type '%s' not found", docType)
	}
	return component.Sanitizer, nil
}
