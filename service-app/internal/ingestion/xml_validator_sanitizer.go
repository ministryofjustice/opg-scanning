package ingestion

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

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

	sanitizedXmlData, err := XmlSanitize(xmlData)
	if err != nil {
		return nil, fmt.Errorf("sanitization failed: %w", err)
	}

	parsedBaseXml, err := parser.BaseParserXml([]byte(sanitizedXmlData))
	if err != nil {
		return nil, fmt.Errorf("base XML parsing failed: %w", err)
	}

	return parsedBaseXml, err
}

func XmlSanitize(xmlData string) (string, error) {
	if xmlData == "" {
		return "", errors.New("empty XML data provided")
	}

	// Remove comments from XML
	commentRegex := regexp.MustCompile(`(?s)<!--.*?-->`)
	sanitized := commentRegex.ReplaceAllString(xmlData, "")

	// Trim unnecessary whitespaces
	sanitized = strings.TrimSpace(sanitized)

	// Remove any potentially harmful attributes, e.g., "script"
	sanitized = strings.ReplaceAll(sanitized, "<script", "<removed_script")
	sanitized = strings.ReplaceAll(sanitized, "</script>", "</removed_script>")

	// Remove unwanted elements (example: potentially dangerous elements)
	unwantedTags := []string{"<iframe", "</iframe>", "<object", "</object>", "<embed", "</embed>"}
	for _, tag := range unwantedTags {
		sanitized = strings.ReplaceAll(sanitized, tag, "")
	}

	return sanitized, nil
}
