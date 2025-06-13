package util

import (
	"encoding/base64"
	"fmt"
)

func DecodeEmbeddedXML(encodedXML string) ([]byte, error) {
	if encodedXML == "" {
		return nil, fmt.Errorf("embedded XML is empty")
	}

	decodedXML, err := base64.StdEncoding.DecodeString(encodedXML)
	if err != nil {
		return nil, fmt.Errorf("failed to decode embedded XML: %w", err)
	}

	return decodedXML, nil
}
