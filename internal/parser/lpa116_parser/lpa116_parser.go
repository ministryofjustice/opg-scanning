package lpa116_parser

import (
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lpa116_types"
)

func Parse(data []byte) (any, error) {
	doc := &lpa116_types.LPA116Document{}
	return parser.DocumentParser(data, doc)
}
