package lp1f_parser

import (
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidXML(t *testing.T) {
	validator := getValidator(t, "LP1F-valid.xml")
	errMessages := validator.Validate()
	errMessagesLen := len(errMessages)
	if errMessagesLen > 0 {
		t.Errorf("Expected no errors but got %d", errMessagesLen)
	}
}

func TestInvalidXML(t *testing.T) {
	fileName := "LP1F-invalid-dates.xml"
	validator := getValidator(t, fileName)

	expectedErrMsgs := []string{
		"(?i)^Page10 Section9 Witness Signature not set",
		"(?i)^Page10 Section9 Donor Signature not set",
		"(?i)^Page10 Section9 Witness Full Name not set",
		"(?i)^Page10 Section9 Witness Address not valid",
		"(?i)^Page10 Section9 Donor signature not set or invalid",
		"(?i)^Page12\\[0] Section11 Witness Signature not set",
		"(?i)^applicant date is invalid",
		"(?i)^all form dates must be before the earliest applicant signature date",
	}

	parser.DocumentValidationTestHelper(t, fileName, expectedErrMsgs, validator)
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
