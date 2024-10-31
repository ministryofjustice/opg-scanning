package parser

import (
	"encoding/xml"
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type BaseParser interface {
	ParseDocument(data []byte) (interface{}, error)
	ParseImage(data []byte) (interface{}, error)
}

func BaseParserXml(data []byte) (*types.Set, error) {
	var parsed types.Set
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("invalid base XML format: %w", err)
	}

	return &parsed, nil
}
