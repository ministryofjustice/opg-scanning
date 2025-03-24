package api

import (
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"os"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
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

		_, _, _, tokenGenerator := auth.PrepareMocks(mockConfig, logger)
		httpClient := httpclient.NewHttpClient(*mockConfig, *logger)
		httpMiddleware, _ := httpclient.NewMiddleware(httpClient, tokenGenerator)

		// Prepare service instance
		service := &Service{
			Client: &Client{Middleware: httpMiddleware},
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

		ctx := context.Background()
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

func TestAttachDocument_Set_Supervision(t *testing.T) {
	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: "scanning",
		Provider: "sirius",
	})
	assert.Nil(t, err)

	// Load XML data from the test file specific to DEPREPORTS
	xmlStringData := util.LoadXMLFileTesting(t, "../../xml/set/Supervision-DEPREPORTS.xml")
	xmlBase64 := base64.StdEncoding.EncodeToString(xmlStringData)
	if xmlBase64 == "" {
		t.Fatal("failed to load sample DeputyReport XML")
	}

	var sSet types.BaseSet
	err = xml.Unmarshal(xmlStringData, &sSet)
	assert.NoError(t, err)
	// We expect two Document nodes (one DEPREPORTS and one DEPCORRES)
	assert.Equal(t, 2, len(sSet.Body.Documents), "expected 2 documents in the set")

	for idx, d := range sSet.Body.Documents {
		doc := sSet.Body.Documents[idx]

		// Set up expected interactions
		mockProvider.
			AddInteraction().
			Given(fmt.Sprintf("An %v with UID %v", d.Type, sSet.Header.CaseNo)).
			Given("I am a DDC user").
			UponReceiving("A request to attach a scanned document").
			WithRequest("POST", "/api/public/v1/scanned-documents", func(b *consumer.V4RequestBuilder) {
				b.
					Header("Content-Type", matchers.String("application/json")).
					JSONBody(matchers.Map{
						"caseReference": matchers.String(sSet.Header.CaseNo),
						"content":       matchers.String(doc.EmbeddedPDF),
						"documentType":  matchers.String(d.Type),
						"scannedDate":   matchers.DateTimeGenerated("2014-12-18T14:48:33Z", "yyyy-MM-dd'T'HH:mm:ss'Z'"),
					})
			}).
			WillRespondWith(201, func(b *consumer.V4ResponseBuilder) {
				b.
					Header("Content-Type", matchers.String("application/json")).
					JSONBody(matchers.Map{
						"uuid": matchers.UUID(),
					})
			})
	}

	err = mockProvider.ExecuteTest(t, func(pactConfig consumer.MockServerConfig) error {
		baseURL := fmt.Sprintf("http://%s:%d", pactConfig.Host, pactConfig.Port)

		fmt.Printf("Base URL: %s\n", baseURL)

		// Mock dependencies
		mockConfig := config.NewConfig()
		mockConfig.App.SiriusBaseURL = baseURL
		logger := logger.GetLogger(mockConfig)

		_, _, _, tokenGenerator := auth.PrepareMocks(mockConfig, logger)
		httpClient := httpclient.NewHttpClient(*mockConfig, *logger)
		httpMiddleware, _ := httpclient.NewMiddleware(httpClient, tokenGenerator)

		// For each document in the set, create a Service instance and call AttachDocuments
		svc := &Service{
			Client: &Client{Middleware: httpMiddleware},
			set:    &sSet,
		}
		for _, d := range sSet.Body.Documents {
			// Create a BaseDocument for this individual document.
			svc.originalDoc = &d
			caseResp := &types.ScannedCaseResponse{
				UID: sSet.Header.CaseNo,
			}

			ctx := context.Background()
			resp, decodedXML, err := svc.AttachDocuments(ctx, caseResp)
			if err != nil {
				t.Fatalf("AttachDocuments returned error for document type %s: %v", d.Type, err)
			}

			// Decode the embedded XML
			decodedEmbeddedXML, err := base64.StdEncoding.DecodeString(d.EmbeddedXML)
			if err != nil {
				t.Fatalf("failed to decode embedded XML for document type %s: %v", d.Type, err)
			}

			assert.Equal(t, decodedEmbeddedXML, decodedXML)
			assert.NotEqual(t, "", resp.UUID)
		}

		return err
	})

	assert.NoError(t, err)
}
