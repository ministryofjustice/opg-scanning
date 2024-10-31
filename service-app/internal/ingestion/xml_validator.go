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
	xsdValidator, err := NewXSDValidator(v.config.XSDDirectory, xmlData)
	if err != nil {
		return nil, fmt.Errorf("Failed to create XSD validator: " + err.Error())
	}

	if err := xsdValidator.ValidateXsd(); err != nil {
		return nil, fmt.Errorf("XSD Validation failed: " + err.Error())
	}

	sanitizedXmlData, err := v.sanitizer.XmlSanitize(xmlData)
	if err != nil {
		return nil, fmt.Errorf("Sanitization failed: " + err.Error())
	}

	parsedBaseXml, err := parser.BaseParserXml([]byte(sanitizedXmlData))
	if err != nil {
		return nil, fmt.Errorf("Base XML parsing failed: " + err.Error())
	}

	return parsedBaseXml, err
}
