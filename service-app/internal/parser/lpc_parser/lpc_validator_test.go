package lpc_parser

import (
	"regexp"
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/assert"
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
		`(?i)Page1\[\d+\] requires exactly 2 Attorney blocks, found`,
		`(?i)Page3\[\d+\] requires exactly 2 Witness blocks, found`,
		`(?i)Page4\[\d+\] requires exactly 2 AuthorisedPerson blocks, found`,
	}

	t.Log("Actual messages from validation:")
	for _, msg := range messages {
		t.Log(msg)
	}

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

func getLPCValidator(t *testing.T, fileName string) parser.CommonValidator {
	xml := util.LoadXMLFileTesting(t, "../../../xml/"+fileName)

	doc, err := Parse(xml)
	require.NoError(t, err, "Failed to parse %s", fileName)

	validator := NewValidator()

	err = validator.Setup(doc)
	assert.Nil(t, err)

	return validator
}
