package factory

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type DocumentProcessor struct {
	logger    logger.Logger
	doc       interface{}
	validator types.Validator
	sanitizer types.Sanitizer
}

func (p *DocumentProcessor) Process() (interface{}, error) {
	if err := p.validator.Validate(); err != nil {
		// TODO: processors can choose to mark docs that have validation issues
		// for someone to manually review, for now log the issue and continue.
		p.logger.Error("Document validation failed: " + err.Error())
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
