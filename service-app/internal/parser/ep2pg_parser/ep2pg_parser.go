package ep2pg_parser

import (
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/ep2pg_types"
)

func Parse(data []byte) (interface{}, error) {
	doc := &ep2pg_types.EP2PGDocument{}
	return parser.DocumentParser(data, doc)
}
