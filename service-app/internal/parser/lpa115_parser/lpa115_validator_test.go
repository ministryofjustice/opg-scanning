package lpa115_parser

import (
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidXML(t *testing.T) {
	validator := getValidator(t, "LPA115-valid.xml")
	assert.Len(t, validator.Validate(), 0, "Expected no validation errors")
}

func TestInvalidXML(t *testing.T) {
	fileName := "LPA115-invalid-dates.xml"
	validator := getValidator(t, fileName)

	parser.DocumentValidationTestHelper(t, fileName, []string{
		"(?i)^Failed to parse Page 3 Option A date:",
		"(?i)^Failed to parse Page 3 Option B date:",
		"(?i)^Failed to parse Page 6 Part B date:",
	}, validator)
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
