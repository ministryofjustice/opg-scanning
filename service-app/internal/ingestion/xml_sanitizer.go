package ingestion

import (
	"errors"
	"regexp"
	"strings"
)

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
