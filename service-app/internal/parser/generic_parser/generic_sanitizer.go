package generic_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
)

type Sanitizer struct {
	doc           interface{}
	baseSanitizer *parser.BaseSanitizer
}

func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		doc: struct{}{},
	}
}

func (v *Sanitizer) Setup(doc interface{}) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.baseSanitizer = parser.NewBaseSanitizer(v.doc)

	return nil
}

func (s *Sanitizer) Sanitize() (interface{}, error) {
	if err := s.baseSanitizer.SanitizeStruct(s.doc); err != nil {
		return nil, err
	}

	return s.doc, nil
}
