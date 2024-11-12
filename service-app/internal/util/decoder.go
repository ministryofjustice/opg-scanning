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

func DecodeEmbeddedImage(encodedImage string) ([]byte, error) {
	if encodedImage == "" {
		return nil, fmt.Errorf("embedded image is empty")
	}

	decodedImage, err := base64.StdEncoding.DecodeString(encodedImage)
	if err != nil {
		return nil, fmt.Errorf("failed to decode embedded image: %w", err)
	}

	return decodedImage, nil
}
