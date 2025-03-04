package lpc_parser

import (
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidLPCXML(t *testing.T) {
	validator := getLPCValidator(t, "LPC-valid.xml")
	assert.Len(t, validator.Validate(), 0, "Expected no validation errors")
}

func TestInvalidLPCXML(t *testing.T) {
	fileName := "LPC-invalid.xml"
	validator := getLPCValidator(t, fileName)
	expectedErrMsgs := []string{
		`(?i)Page1\[\d+\] requires exactly 2 Attorney blocks, found`,
		`(?i)Page3\[\d+\] requires exactly 2 Witness blocks, found`,
		`(?i)Page4\[\d+\] requires exactly 2 AuthorisedPerson blocks, found`,
	}
	parser.DocumentValidationTestHelper(t, fileName, expectedErrMsgs, validator)
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
