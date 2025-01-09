package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAttachDocument_Correspondence(t *testing.T) {
	// Mock dependencies
	mockConfig := config.Config{
		Auth: config.Auth{
			ApiUsername:    "opg_document_and_d@publicguardian.gsi.gov.uk",
			JWTSecretARN:   "local/jwt-key",
			CredentialsARN: "local/local-credentials",
			JWTExpiration:  3600,
		},
	}
	logger := *logger.NewLogger(&mockConfig)

	mockHttpClient, _, _, tokenGenerator := auth.PrepareMocks(&mockConfig, &logger)
	httpMiddleware, _ := httpclient.NewMiddleware(mockHttpClient, tokenGenerator)

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

	mockHttpClient.On("HTTPRequest",
		mock.Anything,
		mock.Anything,
		"POST",
		mock.Anything,
		mock.Anything).Return(mockResponseBytes, nil)

	ctx := context.Background()
	response, err := service.AttachDocuments(ctx, caseResponse)
	if err != nil {
		t.Fatalf("AttachDocuments returned error: %v", err)
	}
	assert.NotNil(t, response, "Expected non-nil response")
	assert.Equal(t, mockResponse, response, "Expected response to match mock response")
	mockHttpClient.AssertExpectations(t)
}
