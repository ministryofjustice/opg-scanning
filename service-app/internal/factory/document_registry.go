package factory

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/parser"
)

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
func NewRegistry() (*Registry, error) {
	components := make(map[string]Component)

	// List of supported document types
	docTypes := constants.SupportedDocumentTypes

	// Populate the registry using the utility function
	for _, docType := range docTypes {
		component, err := GetComponent(docType)
		if err != nil {
			return nil, fmt.Errorf("error getting component for %s: %v", docType, err)
		}
		components[docType] = component
	}

	return &Registry{
		components: components,
	}, nil
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
