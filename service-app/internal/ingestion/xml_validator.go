package ingestion

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/config"
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

func (v *XmlValidator) XmlValidateSanitize(xmlData string) (*types.BaseSet, error) {
	if xmlData == "" {
		return nil, fmt.Errorf("empty XML data provided")
	}

	parsedBaseXml, err := parser.BaseParserXml([]byte(xmlData))
	if err != nil {
		return nil, fmt.Errorf("base XML parsing failed: %w", err)
	}

	return parsedBaseXml, err
}
