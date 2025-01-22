package lp1h_parser

import (
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/parser"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/require"
)

var err error

func TestValidXML(t *testing.T) {
	validator := getValidator(t, "LP1H-valid.xml")
	err = validator.Validate()
	require.NoError(t, err, "Expected no errors")
}

func getValidator(t *testing.T, fileName string) parser.CommonValidator {
	xml := util.LoadXMLFileTesting(t, "../../../xml/"+fileName)
	doc, err := Parse(xml)
	require.NoError(t, err)
	validator := NewValidator()
	validator.Setup(doc)
	return validator
}
