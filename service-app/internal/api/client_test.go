package api

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/stretchr/testify/mock"
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
				CourtReference: "123",
				CaseType:       "order",
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
				CaseType: "lpa",
			},
			expectedErr: false,
		},
		{
			name:       "EPA Case without CaseNo",
			xmlPayload: fmt.Sprintf(withoutCaseNoPayload, "EPA"),
			expectedReq: &types.ScannedCaseRequest{
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

func setupMockServer(t *testing.T, expectedReq *types.ScannedCaseRequest) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		var receivedRequest types.ScannedCaseRequest
		if err := json.NewDecoder(r.Body).Decode(&receivedRequest); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if expectedReq != nil {
			if receivedRequest.CaseType != expectedReq.CaseType {
				t.Errorf("expected case type %s, but got %s", expectedReq.CaseType, receivedRequest.CaseType)
			}
			if expectedReq.CourtReference != "" && receivedRequest.CourtReference != expectedReq.CourtReference {
				t.Errorf("expected court reference %s, but got %s", expectedReq.CourtReference, receivedRequest.CourtReference)
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"UID": "dummy-uid-1234"}`))
	}))
}

func runStubCaseTest(t *testing.T, tt requestCaseStub) {
	t.Run(tt.name, func(t *testing.T) {
		t.Parallel()

		endpoint := "api/public/v1/scanned-cases"
		set := parseXMLPayload(t, tt.xmlPayload)
		mockServer := setupMockServer(t, tt.expectedReq)
		defer mockServer.Close()

		logger := *logger.NewLogger()
		mockConfig := config.Config{
			App: config.App{
				SiriusBaseURL: mockServer.URL,
				SiriusScanURL: endpoint,
			},
			Auth: config.Auth{
				ApiUsername:   "test",
				JWTSecretARN:  "local/jwt-key",
				JWTExpiration: 3600,
			},
		}

		// Mock SecretsManager
		mockAwsClient := new(aws.MockAwsClient)
		mockAwsClient.On("GetSecretValue", mock.Anything, mock.AnythingOfType("string")).
			Return("mock-signing-secret", nil)

		httpClient := httpclient.NewHttpClient(mockConfig, logger)
		middleware, err := httpclient.NewMiddleware(httpClient, mockAwsClient)
		if err != nil {
			t.Fatalf("failed to create middleware: %v", err)
		}
		client := NewClient(middleware)

		ctx := context.Background()
		_, err = client.CreateCaseStub(ctx, set)

		if tt.expectedErr {
			if len(err.Error()) == 0 {
				t.Errorf("expected error %v, but got %v", tt.expectedErr, err)
			}
		} else {
			if err != nil {
				t.Errorf("expected no error, but got: %v", err)
			}
		}
	})
}

func TestCreateStubCase(t *testing.T) {
	tests := buildTestCases()
	for _, tt := range tests {
		runStubCaseTest(t, tt)
	}
}
