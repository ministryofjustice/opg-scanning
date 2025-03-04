package ep2pg_parser

import (
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidXML(t *testing.T) {
	validator := getValidator(t, "EP2PG-valid.xml")
	errMessages := validator.Validate()
	errMessagesLen := len(errMessages)
	if errMessagesLen > 0 {
		t.Errorf("Expected no errors but got %d", errMessagesLen)
	}
}

func TestInvalidXML(t *testing.T) {
	fileName := "EP2PG-invalid-dates.xml"
	validator := getValidator(t, fileName)
	expectedErrMsgs := []string{
		"(?i)^Failed to parse date of birth for Donor:",
	}
	parser.TestHelperDocumentValidation(t, fileName, expectedErrMsgs, validator)
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
