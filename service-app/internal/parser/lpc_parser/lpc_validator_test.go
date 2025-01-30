package lpc_parser

import (
	"regexp"
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/require"
)

func TestValidLPCXML(t *testing.T) {
	validator := getLPCValidator(t, "LPC-valid.xml")
	err := validator.Validate()
	require.NoError(t, err, "Expected no errors for valid LPC XML")
}

func TestInvalidLPCXML(t *testing.T) {
	validator := getLPCValidator(t, "LPC-invalid.xml")
	err := validator.Validate()
	require.Error(t, err, "Expected validation errors for invalid LPC XML, but got none")

	messages := validator.(*Validator).baseValidator.GetValidatorErrorMessages()

	expectedErrMsgs := []string{
		"(?i)found only 1 attorney block; 2 required",     // minOccurs=2 but only 1 present
		"(?i)missing Donor element in ContinuationSheet1", // required <Donor> is missing
		"(?i)invalid element BURNN instead of BURN",       // if there is a typo in the element name
		"(?i)missing PhysicalPage element",                // required element not found
	}

	t.Log("Actual messages from validation:")
	t.Log(messages)

	for _, pattern := range expectedErrMsgs {
		regex, compErr := regexp.Compile(pattern)
		require.NoError(t, compErr, "Failed to compile regex for pattern: %s", pattern)

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

func TestInvalidLPCDateOrderXML(t *testing.T) {
	validator := getLPCValidator(t, "LPC-invalid-dates.xml")
	err := validator.Validate()
	require.Error(t, err, "Expected validation errors due to date ordering but got none")

	messages := validator.(*Validator).baseValidator.GetValidatorErrorMessages()

	found := util.Contains(messages, "all form dates must be before the earliest applicant signature date")
	require.True(t, found, "Expected date ordering validation error not found in messages")
}

func getLPCValidator(t *testing.T, fileName string) parser.CommonValidator {
	xml := util.LoadXMLFileTesting(t, "../../../xml/"+fileName)

	doc, err := Parse(xml)
	require.NoError(t, err, "Failed to parse %s", fileName)

	validator := NewValidator()
	validator.Setup(doc)
	return validator
}
