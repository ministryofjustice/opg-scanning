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

func (v *XmlValidator) XmlValidateSanitize(xmlData string) (*types.Set, error) {
	if xmlData == "" {
		return nil, fmt.Errorf("empty XML data provided")
	}

	sanitizedXmlData, err := XmlSanitize(xmlData)
	if err != nil {
		return nil, fmt.Errorf("sanitization failed: " + err.Error())
	}

	parsedBaseXml, err := parser.BaseParserXml([]byte(sanitizedXmlData))
	if err != nil {
		return nil, fmt.Errorf("base XML parsing failed: " + err.Error())
	}

	return parsedBaseXml, err
}
