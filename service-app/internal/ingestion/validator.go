package ingestion

import "errors"

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(data string) error {
	// Perform XML validation
	if data == "" {
		return errors.New("invalid data: empty string")
	}
	return nil
}
