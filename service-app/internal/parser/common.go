package parser

import (
	"encoding/xml"
	"fmt"
)

type CommonValidator interface {
	Setup(doc interface{}) error
	Validate() error
}

type CommonSanitizer interface {
	Setup(doc interface{}) error
	Sanitize() (interface{}, error)
}

func UnmarshalXML[T any](data []byte, target *T) error {
	err := xml.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal XML: %w", err)
	}
	return nil
}
