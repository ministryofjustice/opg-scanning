package lp2_parser

import (
	"regexp"
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var err error

func TestValidXML(t *testing.T) {
	validator := getValidator(t, "LP2-valid.xml")
	err = validator.Validate()
	require.NoError(t, err, "Expected no errors")
}

func TestInvalidXML(t *testing.T) {
	validator := getValidator(t, "LP2-invalid-dates.xml")
	err := validator.Validate()
	require.Error(t, err, "Expected validation errors due to date ordering but got none")

	messages := validator.(*Validator).baseValidator.GetValidatorErrorMessages()

	expectedErrMsgs := []string{
		"(?i)^Failed to parse attorney signature date:",
		"(?i)^Failed to parse attorney date of birth:",
		"(?i)^Both LPA sub-types are selected",
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

func getValidator(t *testing.T, fileName string) parser.CommonValidator {
	xml := util.LoadXMLFileTesting(t, "../../../xml/"+fileName)
	doc, err := Parse(xml)
	require.NoError(t, err)
	validator := NewValidator()

	err = validator.Setup(doc)
	assert.Nil(t, err)

	return validator
}
