package api

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/corresp_types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAttachDocument_Correspondence(t *testing.T) {
	// Mock dependencies
	mockAwsClient := new(aws.MockAwsClient)
	mockAwsClient.On("GetSecretValue", mock.Anything, mock.AnythingOfType("string")).
		Return("mock-signing-secret", nil)

	mockClient := new(httpclient.MockHttpClient)

	mockConfig := config.NewConfig()
	mockClient.On("GetConfig").Return(mockConfig)

	mockLogger := logger.NewLogger()
	mockClient.On("GetLogger").Return(mockLogger)

	middleware, err := httpclient.NewMiddleware(mockClient, mockAwsClient)
	if err != nil {
		t.Fatalf("failed to create middleware: %v", err)
	}

	// XML representation of the processed document
	xmlData := `
		<Correspondence xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		                xsi:noNamespaceSchemaLocation="Correspondence.xsd">
		    <SubType>Unknown</SubType>
		    <CaseNumber>INVALID_CASE123</CaseNumber>
		</Correspondence>`

	// Prepare service instance
	service := &Service{
		Client: &Client{Middleware: middleware}, // Inject mock client
		originalDoc: &types.BaseDocument{
			EmbeddedXML: xmlData, // Use valid XML
			Type:        "Correspondence",
		},
		processedDoc: &corresp_types.Correspondence{
			XMLName: xml.Name{
				Local: "Correspondence",
			},
			CaseNumber: "CASE123",
			SubType:    "Application Related",
		},
		set: &types.BaseSet{
			Header: &types.BaseHeader{
				ScanTime: "2024-12-05 12:34:56Z",
			},
		},
	}

	// Simulate the request
	caseResponse := &types.ScannedCaseResponse{
		UID: "CASE123",
	}
	ctx := context.Background()

	// Mock Sirius response
	mockResponse := &types.ScannedDocumentResponse{
		UID: "CASE123",
	}
	mockResponseBytes, _ := json.Marshal(mockResponse)

	// Mock HTTPRequest method with correct argument matchers
	mockClient.On("HTTPRequest",
		mock.Anything,                            // Context
		mock.AnythingOfType("string"),            // URL
		mock.AnythingOfType("string"),            // Method
		mock.AnythingOfType("[]uint8"),           // Payload ([]byte)
		mock.AnythingOfType("map[string]string"), // Headers
	).Return(mockResponseBytes, nil)

	// Call AttachDocuments
	response, err := service.AttachDocuments(ctx, caseResponse)
	assert.NoError(t, err, "Expected no error from AttachDocuments")
	assert.Equal(t, mockResponse, response, "Expected response to match mock response")

	// Validate that HTTPRequest was called with expected arguments
	mockClient.AssertCalled(t, "HTTPRequest", mock.Anything,
		mock.MatchedBy(func(url string) bool {
			return url == "/api/public/v1/scanned-documents"
		}),
		"POST",
		mock.AnythingOfType("[]uint8"),
		mock.AnythingOfType("map[string]string"),
	)
}
