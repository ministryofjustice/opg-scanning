package ingestion

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/lestrrat-go/libxml2"
	"github.com/lestrrat-go/libxml2/parser"
	"github.com/lestrrat-go/libxml2/xsd"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type XSDValidator struct {
	schema     *xsd.Schema
	xmlContent string
}

type Root struct {
	SchemaLocation string `xml:"http://www.w3.org/2001/XMLSchema-instance noNamespaceSchemaLocation,attr"`
}

func NewXSDValidator(xsdPath string, xmlContent string) (*XSDValidator, error) {
	validPath, err := util.ValidatePath(xsdPath)
	if err != nil {
		return nil, err
	}
	xsdContent, err := os.ReadFile(validPath) //#nosec G304 false positive: we check the path above
	if err != nil {
		return nil, err
	}
	schema, err := xsd.Parse(xsdContent)
	if err != nil {
		return nil, err
	}

	return &XSDValidator{schema: schema, xmlContent: xmlContent}, nil
}

func (v *XSDValidator) ValidateXsd() error {
	doc, err := libxml2.ParseString(v.xmlContent, parser.XMLParseNoNet)
	if err != nil {
		return err
	}
	defer doc.Free()
	return v.schema.Validate(doc)
}

func ExtractSchemaLocation(xmlContent string) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))

	// Search for the root element with the specified attribute
	var root Root
	if err := decoder.Decode(&root); err != nil {
		return "", fmt.Errorf("failed to parse XML: %w", err)
	}

	// Validate the schema location
	if root.SchemaLocation == "" {
		return "", errors.New("import error: no schema location found")
	}

	// Ensure the schema location has no path separators or parent directory references
	if strings.Contains(root.SchemaLocation, "/") || strings.Contains(root.SchemaLocation, "\\") || strings.Contains(root.SchemaLocation, "..") {
		return "", errors.New("import error: invalid schema location")
	}

	return root.SchemaLocation, nil
}
