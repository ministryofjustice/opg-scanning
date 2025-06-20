package lpc_parser

import (
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/lpc_types"
)

func Parse(data []byte) (any, error) {
	doc := &lpc_types.LPCDocument{}
	return parser.DocumentParser(data, doc)
}
