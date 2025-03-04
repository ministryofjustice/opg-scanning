package lp1f_parser

import (
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var err error

func TestValidXML(t *testing.T) {
	validator := getValidator(t, "LP1F-valid.xml")
	err = validator.Validate()
	require.NoError(t, err, "Expected no errors")
}

func TestInvalidXML(t *testing.T) {
	fileName := "LP1F-invalid-dates.xml"
	validator := getValidator(t, fileName)

	expectedErrMsgs := []string{
		"(?i)^Page10 Section9 Witness Signature not set",
		"(?i)^Page10 Section9 Witness Full Name not set",
		"(?i)^Page10 Section9 Witness Address not valid",
	}

	parser.TestHelperDocumentValidation(t, fileName, true, expectedErrMsgs, validator)
}

func TestInvalidDateOrderXML(t *testing.T) {
	validator := getValidator(t, "LP1F-invalid-dates.xml")
	err := validator.Validate()
	require.Error(t, err, "Expected validation errors due to date ordering but got none")

	messages := validator.GetValidatorErrorMessages()
	found := util.Contains(messages, "all form dates must be before the earliest applicant signature date")
	require.True(t, found, "Expected date ordering validation error not found")
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
