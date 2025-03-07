package lpa115_parser

import (
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lpa115_types"
)

func Parse(data []byte) (any, error) {
	doc := &lpa115_types.LPA115Document{}
	return parser.DocumentParser(data, doc)
}
