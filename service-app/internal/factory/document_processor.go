package factory

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type DocumentProcessor struct {
	doc       interface{}
	validator types.Validator
	sanitizer types.Sanitizer
}

func (p *DocumentProcessor) Process() (interface{}, error) {
	if err := p.validator.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sanitizedDoc, err := p.sanitizer.Sanitize(p.doc)
	if err != nil {
		return nil, fmt.Errorf("sanitization failed: %w", err)
	}

	p.doc = sanitizedDoc
	return p.doc, nil
}

func (p *DocumentProcessor) GetDocument() interface{} {
	return p.doc
}
