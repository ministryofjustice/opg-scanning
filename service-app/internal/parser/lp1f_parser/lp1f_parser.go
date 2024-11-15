package lp1f_parser

import (
	"encoding/xml"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	lp1f_types "github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
)

type Parser struct{}

func NewParser() types.Parser {
	return &Parser{}
}

func (p *Parser) Parse(data []byte) (interface{}, error) {
	doc := &lp1f_types.LP1FDocument{}

	// We assume XML for now
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
