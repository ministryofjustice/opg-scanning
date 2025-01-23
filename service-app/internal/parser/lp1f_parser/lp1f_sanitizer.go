package lp1f_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
)

type Sanitizer struct {
	doc           *lp1f_types.LP1FDocument
	baseSanitizer *parser.BaseSanitizer
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

	v.doc = doc.(*lp1f_types.LP1FDocument)
	v.baseSanitizer = parser.NewBaseSanitizer(v.doc)

	return nil
}

func (s *Sanitizer) Sanitize() (interface{}, error) {
	// Sanitize the entire struct dynamically
	if err := s.baseSanitizer.SanitizeStruct(s.doc); err != nil {
		return nil, err
	}

	return s.doc, nil
}
