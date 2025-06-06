package lp1h_parser

import (
	"os"
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidXML(t *testing.T) {
	validator := getValidator(t, "LP1H-valid.xml")
	assert.Len(t, validator.Validate(), 0, "Expected no validation errors")
}

func getValidator(t *testing.T, fileName string) parser.CommonValidator {
	xml, err := os.ReadFile("../../../testdata/xml/" + fileName)
	require.NoError(t, err)
	doc, err := Parse(xml)
	require.NoError(t, err)
	validator := NewValidator()

	err = validator.Setup(doc)
	assert.Nil(t, err)

	return validator
}
