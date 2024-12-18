package lp1f_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
	lp1f_types "github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
)

type Sanitizer struct {
	doc             *lp1f_types.LP1FDocument
	commonSanitizer *parser.Sanitizer
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
