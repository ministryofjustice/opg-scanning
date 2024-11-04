package lp1f

import (
	"errors"
	"strings"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type Sanitizer struct{}

func NewSanitizer() *Sanitizer {
	return &Sanitizer{}
}

func (s *Sanitizer) Sanitize(data *types.LP1FDocument) (*types.LP1FDocument, error) {
	if data == nil {
		return nil, errors.New("cannot sanitize nil data")
	}

	// Sanitize Page1 data
	data.Page1.Section1.FirstName = sanitizeString(data.Page1.Section1.FirstName)
	data.Page1.Section1.LastName = sanitizeString(data.Page1.Section1.LastName)
	data.Page1.Section1.Address = sanitizeString(data.Page1.Section1.Address)
	data.Page1.Section1.EmailAddress = sanitizeString(data.Page1.Section1.EmailAddress)

	return data, nil
}

func sanitizeString(input string) string {
	// Just an example for now
	return strings.TrimSpace(strings.ReplaceAll(input, "<script>", ""))
}
