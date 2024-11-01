package ingestion

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type XmlValidator struct {
	config    config.Config
	sanitizer *Sanitizer
}

func NewXmlValidator(config config.Config) *XmlValidator {
	return &XmlValidator{
		config:    config,
		sanitizer: NewSanitizer(),
	}
}

func (v *XmlValidator) XmlValidate(xmlData string) (*types.Set, error) {
	if xmlData == "" {
		return nil, fmt.Errorf("empty XML data provided")
	}

	xsdValidator, err := NewXSDValidator(v.config.XSDDirectory, xmlData)
	if err != nil {
		return nil, fmt.Errorf("dailed to create XSD validator: " + err.Error())
	}

	if err := xsdValidator.ValidateXsd(); err != nil {
		return nil, fmt.Errorf("xsd Validation failed: " + err.Error())
	}

	sanitizedXmlData, err := v.sanitizer.XmlSanitize(xmlData)
	if err != nil {
		return nil, fmt.Errorf("sanitization failed: " + err.Error())
	}

	parsedBaseXml, err := parser.BaseParserXml([]byte(sanitizedXmlData))
	if err != nil {
		return nil, fmt.Errorf("base XML parsing failed: " + err.Error())
	}

	return parsedBaseXml, err
}
