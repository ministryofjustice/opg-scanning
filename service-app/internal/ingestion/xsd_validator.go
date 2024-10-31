package ingestion

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/lestrrat-go/libxml2"
	"github.com/lestrrat-go/libxml2/xsd"
)

type XSDValidator struct {
	schema     *xsd.Schema
	xmlContent string
}

type Root struct {
	SchemaLocation string `xml:"http://www.w3.org/2001/XMLSchema-instance noNamespaceSchemaLocation,attr"`
}

func NewXSDValidator(xsdPath string, xmlContent string) (*XSDValidator, error) {
	xsdContent, err := os.ReadFile(xsdPath)
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
	doc, err := libxml2.ParseString(v.xmlContent)
	if err != nil {
		return err
	}
	defer doc.Free()
	return v.schema.Validate(doc)
}

func (v *XSDValidator) getXsdName() (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(v.xmlContent))

	// Search for the root element with the specified attribute
	var root Root
	if err := decoder.Decode(&root); err != nil {
		return "", fmt.Errorf("failed to parse XML: %w", err)
	}

	// Validate the schema location
	if root.SchemaLocation == "" {
		return "", errors.New("import error: no schema location found")
	}

	return root.SchemaLocation, nil
}
