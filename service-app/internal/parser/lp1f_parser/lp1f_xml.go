package lpf1_parser

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	lp1f_types "github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
)

func ParseLP1FXml(data []byte) (*lp1f_types.LP1FDocument, error) {
	doc := &lp1f_types.LP1FDocument{}

	err := xml.Unmarshal(data, doc)
	if err != nil {
		return nil, err
	}

	// Validate required fields based on struct tags
	if err := parser.ValidateStruct(doc); err != nil {
		return nil, err
	}

	return doc, nil
}
