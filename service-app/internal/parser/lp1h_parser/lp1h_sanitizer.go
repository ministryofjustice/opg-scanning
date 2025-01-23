package lp1h_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1h_types"
)

type Sanitizer struct {
	doc             *lp1h_types.LP1HDocument
	commonSanitizer *parser.Sanitizer
}

func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		doc: &lp1h_types.LP1HDocument{},
	}
}

func (v *Sanitizer) Setup(doc interface{}) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.doc = doc.(*lp1h_types.LP1HDocument)
	v.commonSanitizer = parser.NewSanitizer(v.doc)

	return nil
}

func (s *Sanitizer) Sanitize() (interface{}, error) {
	// Sanitize the entire struct dynamically
	if err := s.commonSanitizer.SanitizeStruct(s.doc); err != nil {
		return nil, err
	}

	return s.doc, nil
}
