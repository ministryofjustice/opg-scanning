package lp2_parser

import (
	"os"
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidXML(t *testing.T) {
	validator := getValidator(t, "LP2-valid.xml")
	assert.Len(t, validator.Validate(), 0, "Expected no validation errors")
}

func TestInvalidXML(t *testing.T) {
	fileName := "LP2-invalid-dates.xml"
	validator := getValidator(t, fileName)

	parser.DocumentValidationTestHelper(t, fileName, []string{
		"(?i)^Failed to parse attorney signature date:",
		"(?i)^Failed to parse attorney date of birth:",
		"(?i)^Both LPA sub-types are selected",
	}, validator)
}

func getValidator(t *testing.T, fileName string) parser.CommonValidator {
	xml, err := os.ReadFile("../../../xml/" + fileName)
	require.NoError(t, err)
	doc, err := Parse(xml)
	require.NoError(t, err)
	validator := NewValidator()

	err = validator.Setup(doc)
	assert.Nil(t, err)

	return validator
}
