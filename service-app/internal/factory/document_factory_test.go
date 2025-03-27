package factory

import (
	"context"
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"github.com/lestrrat-go/libxml2"
	"github.com/lestrrat-go/libxml2/xsd"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/corresp_types"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1f_types"
	"github.com/ministryofjustice/opg-scanning/internal/types/lp1h_types"
	"github.com/ministryofjustice/opg-scanning/internal/types/lpa120_types"
	"github.com/ministryofjustice/opg-scanning/internal/types/lpc_types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessDocument_LP1F(t *testing.T) {
	processedDoc := prepareDocument(t, "LP1F", "LP1F-valid")

	lp1fDoc, ok := processedDoc.(*lp1f_types.LP1FDocument)
	require.True(t, ok, "Expected processedDoc to be of type *lp1f_types.LP1FDocument")

	assert.Equal(t, "Charles", lp1fDoc.Page1.Section1.FirstName, "FirstName mismatch")
	assert.Equal(t, "Anderson", lp1fDoc.Page1.Section1.LastName, "LastName mismatch")
}

func TestProcessDocument_LP1H(t *testing.T) {
	processedDoc := prepareDocument(t, "LP1H", "LP1H-valid")

	lp1hDoc, ok := processedDoc.(*lp1h_types.LP1HDocument)
	require.True(t, ok, "Expected processedDoc to be of type *lp1f_types.LP1HDocument")

	assert.Equal(t, "John", lp1hDoc.Page1.Section1.FirstName, "FirstName mismatch")
	assert.Equal(t, "Doe", lp1hDoc.Page1.Section1.LastName, "LastName mismatch")
}

func TestProcessDocument_Correspondence(t *testing.T) {
	processedDoc := prepareDocument(t, "Correspondence", "Correspondence-valid")

	corresp, ok := processedDoc.(*corresp_types.Correspondence)
	require.True(t, ok, "Expected processedDoc to be of type *corresp_types.Correspondence")

	assert.Equal(t, "Legal", corresp.SubType, "SubType mismatch")
	assert.Equal(t, "12345", corresp.CaseNumber[0], "CaseNumber mismatch")
}

func TestProcessDocument_LPC(t *testing.T) {
	processedDoc := prepareDocument(t, "LPC", "LPC-valid")

	lpc, ok := processedDoc.(*lpc_types.LPCDocument)
	require.True(t, ok, "Expected processedDoc to be of type *lpc_types.LPCDocument")

	assert.Equal(t, "Jack", lpc.Page1[0].ContinuationSheet1.Attorneys[0].FirstName, "FirstName mismatch")
	assert.Equal(t, "Jones", lpc.Page1[0].ContinuationSheet1.Attorneys[0].LastName, "LastName mismatch")
}

func TestProcessDocument_LPA120(t *testing.T) {
	processedDoc := prepareDocument(t, "LPA120", "LPA120-valid")

	lpa120, ok := processedDoc.(*lpa120_types.LPA120Document)
	require.True(t, ok, "Expected processedDoc to be of type *lpa120_types.LPA120Document")

	assert.Equal(t, "John Doe", lpa120.Page3.Section1.FullName, "FullName mismatch")
}

func TestProcessGenericDocuments(t *testing.T) {
	testCases := []struct {
		name    string
		docType string
		valid   string
	}{
		{"LP2", "LP2", "LP2-valid"},
		{"LPA002", "LPA002", "LPA002-valid"},
		{"LPA-PA", "LPA-PA", "LPA-PA-valid"},
		{"LPA-PW", "LPA-PW", "LPA-PW-valid"},
		{"LPA114", "LPA114", "LPA114-valid"},
		{"LPA117", "LPA117", "LPA117-valid"},
		{"EP2PG", "EP2PG", "EP2PG-valid"},
		{"EPA", "EPA", "EPA-valid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prepareDocument(t, tc.docType, tc.valid)
		})
	}
}

// Verifies that valid data returns proper XML given LP1F.
func TestGenerateXMLFromProcessedDocument_Success(t *testing.T) {
	processedDoc := prepareDocument(t, "LP1F", "LP1F-valid")

	lp1fDoc, ok := processedDoc.(*lp1f_types.LP1FDocument)
	require.True(t, ok, "Expected processedDoc to be of type *lp1f_types.LP1FDocument")

	input := lp1fDoc
	result, err := GenerateXMLFromProcessedDocument(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedHeader := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
	if !strings.HasPrefix(string(result), expectedHeader) {
		t.Errorf("Expected XML to start with header %q, got %q", expectedHeader, result)
	}

	if !strings.Contains(string(result), "<LP1F>") {
		t.Errorf("Expected XML to contain <LP1F>, got %q", result)
	}
}

// Verifies that an unmarshalable input causes an error.
func TestGenerateXMLFromProcessedDocument_Error(t *testing.T) {
	// A channel cannot be marshalled to XML.
	input := make(chan int)
	_, err := GenerateXMLFromProcessedDocument(input)
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}

func TestProcessedXMLValidation(t *testing.T) {
	mockConfig := config.NewConfig()

	processedDoc := prepareDocument(t, "LP1F", "LP1F-valid")

	lp1fDoc, ok := processedDoc.(*lp1f_types.LP1FDocument)
	require.True(t, ok, "Expected processedDoc to be of type *lp1f_types.LP1FDocument")

	input := lp1fDoc
	result, err := GenerateXMLFromProcessedDocument(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Run XSD validation on the processed XML.
	// Assume schemaPath is the path to your XSD.
	schemaPath := mockConfig.App.ProjectFullPath + "/xsd/LP1F.xsd"
	xsdContent, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("failed to load schemaPath: %v", err)
	}
	schema, err := xsd.Parse(xsdContent)
	if err != nil {
		t.Fatalf("failed to parse xsdContent: %v", err)
	}

	doc, err := libxml2.ParseString(string(result))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Free()

	if err := schema.Validate(doc); err != nil {
		t.Fatalf("XSD validation failed: %v", err)
	}
}

func prepareDocument(t *testing.T, docType string, fileName string) interface{} {
	// Load the sample XML from the xml directory
	encodedXML := loadXMLFile(t, "../../xml/"+fileName+".xml")
	if encodedXML == "" {
		t.Fatal("failed to load sample XML")
	}

	// Create a base document with the loaded XML
	doc := &types.BaseDocument{
		Type:        docType,
		Encoding:    "base64",
		NoPages:     1,
		EmbeddedXML: encodedXML,
	}

	// Create a new DocumentProcessor using the factory
	registry, err := NewRegistry()
	require.NoError(t, err, "Failed create Registry")

	cfg := config.NewConfig()
	logger := logger.GetLogger(cfg)
	processor, err := NewDocumentProcessor(doc, doc.Type, "XML", registry, logger)
	require.NoError(t, err, "NewDocumentProcessor returned an error")

	// Process the document
	ctx := context.Background()
	processedDoc, err := processor.Process(ctx)
	require.NoError(t, err, "Document processing failed")

	return processedDoc
}

func loadXMLFile(t *testing.T, filepath string) string {
	validPath, err := util.ValidatePath(filepath)
	if err != nil {
		require.FailNow(t, "Invalid file path", err.Error())
	}
	data, err := os.ReadFile(validPath)
	require.NoError(t, err, "Failed to read XML file")
	return base64.StdEncoding.EncodeToString(data)
}
