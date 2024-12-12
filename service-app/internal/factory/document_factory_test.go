package factory

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/lpf1_types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessDocument_LP1F(t *testing.T) {
	// Load the sample XML from the xml directory
	encodedXML := loadXMLFile(t, "../../xml/LP1F-valid.xml")
	if encodedXML == "" {
		t.Fatal("failed to load sample XML")
	}

	// Create a base document with the loaded XML
	doc := &types.BaseDocument{
		Type:        "LP1F",
		Encoding:    "base64",
		NoPages:     1,
		EmbeddedXML: encodedXML,
	}

	// Create a new DocumentProcessor using the factory
	registry := NewRegistry()
	logger := logger.NewLogger()
	processor, err := NewDocumentProcessor(doc, doc.Type, "XML", registry, logger)
	require.NoError(t, err, "NewDocumentProcessor returned an error")

	// Process the document
	processedDoc, err := processor.Process()
	require.NoError(t, err, "Document processing failed")

	lp1fDoc, ok := processedDoc.(*lpf1_types.LP1FDocument)
	require.True(t, ok, "expected processedDoc to be of type *lp1f_types.LP1FDocument")

	assert.Equal(t, "ANDREW ROBERT", lp1fDoc.Page1.Section1.FirstName, "FirstName mismatch")
	assert.Equal(t, "HEPBURN", lp1fDoc.Page1.Section1.LastName, "LastName mismatch")
}

func loadXMLFile(t *testing.T, filepath string) string {
	data, err := os.ReadFile(filepath)
	require.NoError(t, err, "failed to read XML file")
	return base64.StdEncoding.EncodeToString(data)
}
