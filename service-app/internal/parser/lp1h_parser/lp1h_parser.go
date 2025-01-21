package lp1h_parser

import (
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1h_types"
)

func Parse(data []byte) (interface{}, error) {
	doc := &lp1h_types.LP1HDocument{}
	return parser.DocumentParser(data, doc)
}
