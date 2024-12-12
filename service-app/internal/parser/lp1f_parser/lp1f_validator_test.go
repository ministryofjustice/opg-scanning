package lp1f_parser

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/stretchr/testify/require"
)

var err error

func TestValidXML(t *testing.T) {
	validator := getValidator(t, "LP1F-valid.xml")
	err = validator.Validate()
	require.NoError(t, err, "Expected no errors")
}

func TestInvalidXML(t *testing.T) {
	validator := getValidator(t, "LP1F-invalid.xml")
	err := validator.Validate()
	require.Error(t, err, "Expected validation errors due to date ordering but got none")

	messages := validator.(*Validator).commonValidator.GetValidatorErrorMessages()

	expectedErrMsgs := []string{
		"(?i)^Page10 Section9 Witness Signature not set",
		"(?i)^Page10 Section9 Donor Signature not set",
		"(?i)^Page10 Section9 Witness Full Name not set",
		"(?i)^Page10 Section9 Witness Address not valid",
		"(?i)^Page10 Section9 Donor signature not set or invalid",
		"(?i)^Page12\\[2\\] Section11 Witness Signature not set",
		"(?i)^no valid applicant signature/dates found",
	}

	t.Log("Actual messages from validation:")
	t.Log(messages)

	// Match each expected pattern against actual messages using regex
	for _, pattern := range expectedErrMsgs {
		regex, err := regexp.Compile(pattern)
		require.NoError(t, err, "Failed to compile regex for pattern: %s", pattern)

		// Check if any actual message matches the current regex pattern
		found := false
		for _, msg := range messages {
			if regex.MatchString(msg) {
				found = true
				break
			}
		}

		require.True(t, found, "Expected error message pattern not found: %s", pattern)
	}
}

func TestInvalidDateOrderXML(t *testing.T) {
	validator := getValidator(t, "LP1F-invalid-dates.xml")
	err := validator.Validate()
	require.Error(t, err, "Expected validation errors due to date ordering but got none")

	messages := validator.(*Validator).commonValidator.GetValidatorErrorMessages()
	found := containsMessage(messages, "all form dates must be before the earliest applicant signature date")
	require.True(t, found, "Expected date ordering validation error not found")
}

func getValidator(t *testing.T, fileName string) parser.CommonValidator {
	xml := loadXMLFile(t, "../../../xml/"+fileName)
	doc, err := Parse([]byte(xml))
	require.NoError(t, err)
	validator := NewValidator()
	validator.Setup(doc)
	return validator
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
