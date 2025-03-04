package api

import (
	"context"
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
)

type requestCaseStub struct {
	name        string
	xmlPayload  string
	expectedReq *types.ScannedCaseRequest
	expectedErr bool
}

const (
	withCaseNoPayload = `<?xml version="1.0" encoding="UTF-8"?>
	<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
		<Header CaseNo="123" Scanner="9" ScanTime="2014-09-26 12:38:53" ScannerOperator="Administrator" Schedule="02-0001112-20160909185000" FeeNumber="1234"/>
		<Body>
			<Document Type="%s" Encoding="UTF-8" NoPages="19"></Document>
		</Body>
	</Set>`
	withoutCaseNoPayload = `<?xml version="1.0" encoding="UTF-8"?>
	<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
		<Header CaseNo="" Scanner="9" ScanTime="2014-09-26 12:38:53" ScannerOperator="Administrator" Schedule="02-0001112-20160909185000" FeeNumber="1234"/>
		<Body>
			<Document Type="%s" Encoding="UTF-8" NoPages="19"></Document>
		</Body>
	</Set>`
)

func buildTestCases() []requestCaseStub {
	return []requestCaseStub{
		{
			name:       "Order Case with CaseNo",
			xmlPayload: fmt.Sprintf(withCaseNoPayload, "COPORD"),
			expectedReq: &types.ScannedCaseRequest{
				BatchID:        "02-0001112-20160909185000",
				CaseType:       "order",
				CourtReference: "123",
			},
			expectedErr: false,
		},
		{
			name:        "Invalid CaseNo for non-COPORD",
			xmlPayload:  fmt.Sprintf(withCaseNoPayload, "EPA"),
			expectedReq: nil,
			expectedErr: false,
		},
		{
			name:       "LPA Case without CaseNo",
			xmlPayload: fmt.Sprintf(withoutCaseNoPayload, "LPA002"),
			expectedReq: &types.ScannedCaseRequest{
				BatchID:  "02-0001112-20160909185000",
				CaseType: "lpa",
			},
			expectedErr: false,
		},
		{
			name:       "EPA Case without CaseNo",
			xmlPayload: fmt.Sprintf(withoutCaseNoPayload, "EPA"),
			expectedReq: &types.ScannedCaseRequest{
				BatchID:  "02-0001112-20160909185000",
				CaseType: "epa",
			},
			expectedErr: false,
		},
		{
			name:        "Invalid Document Type without CaseNo",
			xmlPayload:  fmt.Sprintf(withoutCaseNoPayload, "INVALID"),
			expectedReq: nil,
			expectedErr: true,
		},
	}
}

func parseXMLPayload(t *testing.T, payload string) types.BaseSet {
	var set types.BaseSet
	if err := xml.Unmarshal([]byte(payload), &set); err != nil {
		t.Fatalf("failed to parse XML payload: %v", err)
	}
	return set
}

func runStubCaseTest(t *testing.T, tt requestCaseStub) {
	// Set up Pact
	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: "scanning",
		Provider: "sirius",
	})
	assert.Nil(t, err)

	t.Run(tt.name, func(t *testing.T) {
		set := parseXMLPayload(t, tt.xmlPayload)
		mockConfig := config.NewConfig()
		logger := logger.NewLogger(mockConfig)

		// Set up expected interactions
		if tt.expectedReq != nil {
			mockProvider.
				AddInteraction().
				Given("I am a DDC user").
				UponReceiving("A request to create a stub case").
				WithRequest("POST", "/api/public/v1/scanned-cases", func(b *consumer.V4RequestBuilder) {
					bodyMatcher := matchers.Map{
						"batchId":     matchers.String(tt.expectedReq.BatchID),
						"caseType":    matchers.String(tt.expectedReq.CaseType),
						"receiptDate": matchers.DateTimeGenerated("2014-09-26T12:38:53Z", "yyyy-MM-dd'T'HH:mm:ss'Z'"),
						"createdDate": matchers.DateTimeGenerated("2025-02-16T11:40:45Z", "yyyy-MM-dd'T'HH:mm:ss'Z'"),
					}

					if tt.expectedReq.CourtReference != "" {
						bodyMatcher["courtReference"] = matchers.String(tt.expectedReq.CourtReference)
					}

					b.
						Header("Content-Type", matchers.String("application/json")).
						JSONBody(bodyMatcher)
				}).
				WillRespondWith(201, func(b *consumer.V4ResponseBuilder) {
					b.
						Header("Content-Type", matchers.String("application/json")).
						JSONBody(matchers.Map{
							"uId": matchers.Regex("7000-3737-2818", `7\d{3}-\d{4}-\d{4}`),
						})
				})
		}

		err = mockProvider.ExecuteTest(t, func(pactConfig consumer.MockServerConfig) error {
			baseURL := fmt.Sprintf("http://%s:%d", pactConfig.Host, pactConfig.Port)
			mockConfig.App.SiriusBaseURL = baseURL

			// Mock dependencies
			_, _, _, tokenGenerator := auth.PrepareMocks(mockConfig, logger)
			httpClient := httpclient.NewHttpClient(*mockConfig, *logger)
			httpMiddleware, _ := httpclient.NewMiddleware(httpClient, tokenGenerator)

			client := NewClient(httpMiddleware)
			service := NewService(client, &set)

			ctx := context.Background()

			response, err := service.CreateCaseStub(ctx)

			if tt.expectedErr {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, "7000-3737-2818", response.UID)
			}

			return nil
		})

		if tt.expectedReq != nil {
			assert.NoError(t, err)
		}
	})
}

func TestCreateStubCase(t *testing.T) {
	tests := buildTestCases()
	for _, tt := range tests {
		runStubCaseTest(t, tt)
	}
}
