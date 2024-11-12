package lp1f_parser

import (
	"os"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	lp1f_types "github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
	"github.com/stretchr/testify/require"
)

var err error

func TestValidXML(t *testing.T) {
	validator := getValidator(t, "LP1F-valid.xml")
	err = validator.Validate()
	require.NoError(t, err, "Expected no errors")
}

func TestInvalidDateOrderXML(t *testing.T) {
	validator := getValidator(t, "LP1F-invalid-dates.xml")
	err := validator.Validate()
	require.Error(t, err, "Expected validation errors due to date ordering but got none")

	messages := validator.(*Validator).commonValidator.GetValidatorErrorMessages()
	found := containsMessage(messages, "all form dates must be before the earliest applicant signature date")
	require.True(t, found, "Expected date ordering validation error not found")
}

func getValidator(t *testing.T, fileName string) types.Validator {
	xml := loadXMLFile(t, "../../../xml/"+fileName)
	doc, err := NewParser().Parse([]byte(xml))
	require.NoError(t, err)

	return NewValidator(doc.(*lp1f_types.LP1FDocument))
}

func loadXMLFile(t *testing.T, filepath string) string {
	data, err := os.ReadFile(filepath)
	if err != nil {
		require.FailNow(t, "Failed to read XML file", err.Error())
	}
	return string(data)
}

func containsMessage(messages []string, expectedMessage string) bool {
	for _, msg := range messages {
		if strings.Contains(msg, expectedMessage) {
			return true
		}
	}
	return false
}
