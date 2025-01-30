package lpc_parser

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lpc_types"
)

type Sanitizer struct {
	doc           *lpc_types.LPCDocument
	baseSanitizer *parser.BaseSanitizer
}

func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		doc: &lpc_types.LPCDocument{},
	}
}

func (v *Sanitizer) Setup(doc interface{}) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	v.doc = doc.(*lpc_types.LPCDocument)
	v.baseSanitizer = parser.NewBaseSanitizer(v.doc)

	return nil
}

func (s *Sanitizer) Sanitize() (interface{}, error) {
	if err := s.baseSanitizer.SanitizeStruct(s.doc); err != nil {
		return nil, err
	}

	return s.doc, nil
}
