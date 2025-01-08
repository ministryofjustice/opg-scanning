package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"regexp"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/mocks"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAttachDocument_Correspondence(t *testing.T) {
	// Mock dependencies
	mockConfig := config.Config{}
	logger := *logger.NewLogger(&mockConfig)

	_, httpMiddleware, _ := mocks.PrepareMocks(&mockConfig, &logger)
	mockClient := new(httpclient.MockHttpClient)

	// Load PDF from the test file
	data, err := os.ReadFile("../../pdf/dummy.pdf")
	if data == nil || err != nil {
		t.Fatal("failed to load dummy PDF")
	}

	// Load XML data from the test file
	xmlStringData := util.LoadXMLFileTesting(t, "../../xml/Correspondence-valid.xml")
	xmlData := base64.StdEncoding.EncodeToString(xmlStringData)
	if xmlData == "" {
		t.Fatal("failed to load sample XML")
	}

	// Prepare service instance
	service := &Service{
		Client: &Client{Middleware: httpMiddleware},
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
