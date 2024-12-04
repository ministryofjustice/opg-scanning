package corresp_parser

import (
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/corresp_types"
)

func Parse(data []byte) (*corresp_types.Correspondence, error) {
	doc := &corresp_types.Correspondence{}
	if err := parser.UnmarshalXML(data, doc); err != nil {
		return nil, err
	}
	// Validate required fields based on struct tags
	if err := parser.ValidateStruct(doc); err != nil {
		return nil, err
	}

	return doc, nil
}
