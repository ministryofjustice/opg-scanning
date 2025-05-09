package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
)

func TestAttachDocument_Correspondence(t *testing.T) {
	// Set up Pact
	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: "scanning",
		Provider: "sirius",
	})
	assert.Nil(t, err)

	// Load PDF from the test file
	pdfRaw, err := os.ReadFile("../../pdf/dummy.pdf")
	if pdfRaw == nil || err != nil {
		t.Fatal("failed to load dummy PDF")
	}
	pdfBase64 := base64.StdEncoding.EncodeToString(pdfRaw)

	// Load XML data from the test file
	xmlStringData := util.LoadXMLFileTesting(t, "../../xml/Correspondence-valid.xml")
	xmlBase64 := base64.StdEncoding.EncodeToString(xmlStringData)
	if xmlBase64 == "" {
		t.Fatal("failed to load sample XML")
	}

	// Set up expected interactions
	mockProvider.
		AddInteraction().
		Given("An LPA with UID 7000-3764-4871 exists").
		Given("I am a DDC user").
		UponReceiving("A request to attach a scanned document").
		WithRequest("POST", "/api/public/v1/scanned-documents", func(b *consumer.V4RequestBuilder) {
			b.
				Header("Content-Type", matchers.String("application/json")).
				JSONBody(matchers.Map{
					"caseReference":   matchers.String("7000-3764-4871"),
					"content":         matchers.String(pdfBase64),
					"documentType":    matchers.String("Correspondence"),
					"documentSubType": matchers.String("Legal"),
					"scannedDate":     matchers.DateTimeGenerated("2025-02-16T11:40:45Z", "yyyy-MM-dd'T'HH:mm:ss'Z'"),
				})
		}).
		WillRespondWith(201, func(b *consumer.V4ResponseBuilder) {
			b.
				Header("Content-Type", matchers.String("application/json")).
				JSONBody(matchers.Map{
					"uuid": matchers.UUID(),
				})
		})

	err = mockProvider.ExecuteTest(t, func(pactConfig consumer.MockServerConfig) error {
		baseURL := fmt.Sprintf("http://%s:%d", pactConfig.Host, pactConfig.Port)

		// Mock dependencies
		mockConfig := config.NewConfig()
		mockConfig.App.SiriusBaseURL = baseURL
		logger := logger.GetLogger(mockConfig)

		httpClient := httpclient.NewHttpClient(*mockConfig, *logger)
		httpMiddleware, _ := httpclient.NewMiddleware(httpClient)

		// Prepare service instance
		service := &service{
			Client: &client{Middleware: httpMiddleware},
			originalDoc: &types.BaseDocument{
				EmbeddedXML: xmlBase64,
				EmbeddedPDF: pdfBase64,
				Type:        "Correspondence",
			},
			set: &types.BaseSet{
				Header: &types.BaseHeader{
					ScanTime: "2024-12-05 12:34:56Z",
				},
			},
		}

		caseResponse := &types.ScannedCaseResponse{
			UID: "7000-3764-4871",
		}

		ctx := context.WithValue(context.Background(), constants.UserContextKey, "my-token")

		response, decodedXML, err := service.AttachDocuments(ctx, caseResponse)
		if err != nil {
			t.Fatalf("AttachDocuments returned error: %v", err)
		}
		assert.NotNil(t, response)
		assert.Equal(t, xmlStringData, decodedXML)
		assert.Equal(t, "fc763eba-0905-41c5-a27f-3934ab26786c", response.UUID)

		return err
	})

	assert.NoError(t, err)
}
