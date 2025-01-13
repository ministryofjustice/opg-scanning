package api

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
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

func runStubCaseTest(t *testing.T, tt requestCaseStub) {
	t.Run(tt.name, func(t *testing.T) {
		t.Parallel()

		set := parseXMLPayload(t, tt.xmlPayload)
		mockConfig := *config.NewConfig()
		logger := *logger.NewLogger(&mockConfig)

		// Mock dependencies
		mockHttpClient, _, _, tokenGenerator := auth.PrepareMocks(&mockConfig, &logger)
		httpMiddleware, _ := httpclient.NewMiddleware(mockHttpClient, tokenGenerator)

		mockHttpClient.On("HTTPRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Maybe().
			Run(func(args mock.Arguments) {
				payload := args[3].([]byte)

				var receivedRequest types.ScannedCaseRequest
				if err := json.NewDecoder(bytes.NewReader(payload)).Decode(&receivedRequest); err != nil {
					t.Fatalf("failed to decode request body: %v", err)
				}

				// Perform assertions or checks on the received request data
				if tt.expectedReq != nil {
					if receivedRequest.CaseType != tt.expectedReq.CaseType {
						t.Errorf("expected case type %s, but got %s", tt.expectedReq.CaseType, receivedRequest.CaseType)
					}
					if tt.expectedReq.CourtReference != "" && receivedRequest.CourtReference != tt.expectedReq.CourtReference {
						t.Errorf("expected court reference %s, but got %s", tt.expectedReq.CourtReference, receivedRequest.CourtReference)
					}
				}
			}).
			Return([]byte(`{"UID": "dummy-uid-1234"}`), nil)

		client := NewClient(httpMiddleware)
		service := NewService(client, &set)

		ctx := context.Background()

		_, err := service.CreateCaseStub(ctx)

		if tt.expectedErr {
			if err == nil {
				t.Errorf("expected error, but got nil")
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}

		// Assert mock expectations
		mockHttpClient.AssertExpectations(t)
	})
}

func TestCreateStubCase(t *testing.T) {
	tests := buildTestCases()
	for _, tt := range tests {
		runStubCaseTest(t, tt)
	}
}
