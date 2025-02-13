package lpa120_parser

import (
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lpa120_types"
)

func Parse(data []byte) (interface{}, error) {
	doc := &lpa120_types.LPA120Document{}
	return parser.DocumentParser(data, doc)
}
