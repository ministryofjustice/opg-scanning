package generic_parser

import (
	"github.com/ministryofjustice/opg-scanning/internal/parser"
)

func Parse(data []byte) (any, error) {
	doc := &struct{}{}

	return parser.DocumentParser(data, doc)
}
