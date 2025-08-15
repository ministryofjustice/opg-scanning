package ingestion

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/config"
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type XmlValidator struct {
	config config.Config
}

func NewXmlValidator(config config.Config) *XmlValidator {
	return &XmlValidator{
		config: config,
	}
}

func (v *XmlValidator) XmlValidate(xmlData []byte) (*types.BaseSet, error) {
	if len(xmlData) == 0 {
		return nil, fmt.Errorf("empty XML data provided")
	}

	parsedBaseXml, err := parser.BaseParserXml(xmlData)
	if err != nil {
		return nil, fmt.Errorf("base XML parsing failed: %w", err)
	}

	return parsedBaseXml, err
}
