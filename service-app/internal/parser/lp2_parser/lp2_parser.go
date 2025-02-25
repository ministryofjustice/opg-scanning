package lp2_parser

import (
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp2_types"
)

func Parse(data []byte) (interface{}, error) {
	doc := &lp2_types.LP2Document{}
	return parser.DocumentParser(data, doc)
}
