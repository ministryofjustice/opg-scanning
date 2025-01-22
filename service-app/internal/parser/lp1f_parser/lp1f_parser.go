package lp1f_parser

import (
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	lp1f_types "github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
)

func Parse(data []byte) (interface{}, error) {
	doc := &lp1f_types.LP1FDocument{}
	return parser.DocumentParser(data, doc)
}
