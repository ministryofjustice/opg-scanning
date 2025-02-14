package corresp_parser

import (
	"encoding/xml"
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types/corresp_types"
)

func Parse(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("correspondence data is empty")
	}

	doc := &corresp_types.Correspondence{}
	if err := xml.Unmarshal(data, doc); err != nil {
		return nil, err
	}
	// Validate required fields based on struct tags
	if err := parser.ValidateStruct(doc); err != nil {
		return nil, err
	}

	return doc, nil
}
