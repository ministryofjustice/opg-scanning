package corresp_parser

import (
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/corresp_types"
)

func Parse(data []byte) (interface{}, error) {
	doc := &corresp_types.Correspondence{}
	return parser.DocumentParser(data, doc)
}
