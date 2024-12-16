package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"regexp"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
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

	mockLogger := logger.NewLogger(mockConfig)
	mockClient.On("GetLogger").Return(mockLogger)

	// Create middleware instance
	middleware, err := httpclient.NewMiddleware(mockClient, mockAwsClient)
	if err != nil {
		t.Fatalf("failed to create middleware: %v", err)
	}

	// Load PDF from the test file
	data, err := os.ReadFile("../../pdf/dummy.pdf")
	if data == nil || err != nil {
		t.Fatal("failed to load dummy PDF")
	}

	// Load XML data from the test file
	xmlData := util.LoadXMLFileTesting(t, "../../xml/Correspondence-valid.xml")
	if xmlData == "" {
		t.Fatal("failed to load sample XML")
	}

	// Prepare service instance
	service := &Service{
		Client: &Client{Middleware: middleware},
		originalDoc: &types.BaseDocument{
			EmbeddedXML: xmlData,
			EmbeddedPDF: base64.StdEncoding.EncodeToString(data),
			Type:        "Correspondence",
		},
		set: &types.BaseSet{
			Header: &types.BaseHeader{
				ScanTime: "2024-12-05 12:34:56Z",
			},
		},
	}

	caseResponse := &types.ScannedCaseResponse{
		UID: "CASE123",
	}

	// Mock the response from HTTPRequest
	mockResponse := &types.ScannedDocumentResponse{
		UUID:                "CASE123",
		Type:                "Correspondence",
		Subtype:             "Application Related",
		SourceDocumentType:  "Correspondence",
		Title:               "Case Document",
		FriendlyDescription: "Scanned document for case",
		ID:                  1,
	}
	mockResponseBytes, _ := json.Marshal(mockResponse)

	// Mock the HTTPRequest method
	mockClient.On("HTTPRequest", mock.Anything, mock.MatchedBy(func(url string) bool {
		// Remove domain from url using regex pattern
		urlWithoutDomain := regexp.MustCompile(`^https?://[^/]+`).ReplaceAllString(url, "")
		return urlWithoutDomain == "/api/public/v1/scanned-documents"
	}), "POST", mock.Anything, mock.Anything).Return(mockResponseBytes, nil)

	ctx := context.Background()
	response, err := service.AttachDocuments(ctx, caseResponse)
	if err != nil {
		t.Fatalf("AttachDocuments returned error: %v", err)
	}
	assert.NotNil(t, response, "Expected non-nil response")
	assert.Equal(t, mockResponse, response, "Expected response to match mock response")

	// Assert HTTPRequest was called with the expected parameters
	mockClient.AssertCalled(t, "HTTPRequest", mock.Anything, mock.MatchedBy(func(url string) bool {
		urlWithoutDomain := regexp.MustCompile(`^https?://[^/]+`).ReplaceAllString(url, "")
		return urlWithoutDomain == "/api/public/v1/scanned-documents"
	}), "POST", mock.Anything, mock.Anything)
}
