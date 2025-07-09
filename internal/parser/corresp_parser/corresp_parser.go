package corresp_parser

import (
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/corresp_types"
)

func Parse(data []byte) (any, error) {
	doc := &corresp_types.Correspondence{}
	return parser.DocumentParser(data, doc)
}
