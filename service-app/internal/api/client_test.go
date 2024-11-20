package api

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type requestCaseStub struct {
	name        string
	xmlPayload  string
	expectedReq *types.ScannedCaseRequest
	expectedErr error
}

func TestCreateStubCase(t *testing.T) {
	docTypes := []string{"COPORD", "EPA", "EP2PG", "LP1F", "LPA002", "LP1H", "LP2"}

	// Test with CaseNo
	for _, docType := range docTypes {
		withCaseNoPayload := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
	<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
		<Header CaseNo="123" Scanner="9" ScanTime="2014-09-26T12:38:53" ScannerOperator="Administrator" Schedule="02-0001112-20160909185000" FeeNumber="1234"/>
		<Body>
			<Document Type="%s" Encoding="UTF-8" NoPages="19"></Document>
		</Body>
	</Set>`, docType)

		runStubCaseTest(t, withCaseNoPayload, true, docType)
	}

	// Test without CaseNo
	for _, docType := range docTypes {
		withoutCaseNoPayload := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
	<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
		<Header CaseNo="" Scanner="9" ScanTime="2014-09-26T12:38:53" ScannerOperator="Administrator" Schedule="02-0001112-20160909185000" FeeNumber="1234"/>
		<Body>
			<Document Type="%s" Encoding="UTF-8" NoPages="19"></Document>
		</Body>
	</Set>`, docType)

		runStubCaseTest(t, withoutCaseNoPayload, false, docType)
	}
}

func runStubCaseTest(t *testing.T, payload string, withCaseNo bool, docType string) {
	tests := buildTestCases(withCaseNo, docType, payload)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Parse XML payload into types.BaseSet
			var set types.BaseSet
			if err := xml.Unmarshal([]byte(tt.xmlPayload), &set); err != nil {
				t.Fatalf("failed to parse XML payload: %v", err)
			}

			// Mock server to simulate /scanned-case endpoint and validate request body
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()

				if r.Method != http.MethodPost {
					t.Errorf("expected POST method, got %s", r.Method)
				}
				if r.URL.Path != "/api/public/v1/scanned-cases" {
					t.Errorf("unexpected URL path: %s", r.URL.Path)
				}

				var receivedRequest types.ScannedCaseRequest
				if err := json.NewDecoder(r.Body).Decode(&receivedRequest); err != nil {
					t.Fatalf("failed to decode request body: %v", err)
				}

				if receivedRequest.CaseType != tt.expectedReq.CaseType {
					t.Errorf("expected case type %s, but got %s", tt.expectedReq.CaseType, receivedRequest.CaseType)
				}
				if tt.expectedReq.CourtReference != "" && receivedRequest.CourtReference != tt.expectedReq.CourtReference {
					t.Errorf("expected court reference %s, but got %s", tt.expectedReq.CourtReference, receivedRequest.CourtReference)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"UID": "dummy-uid-1234"}`))
			}))
			defer mockServer.Close()

			// Mock logger
			logger := *logger.NewLogger()
			mockConfig := config.Config{
				App: config.App{
					SiriusBaseURL: mockServer.URL,
					SiriusScanURL: "api/public/v1/scanned-cases",
				},
			}

			// Instantiate dependencies and call the method
			httpClient := httpclient.NewHttpClient(mockConfig, logger)
			stubCase := NewCreateStubCase(httpClient)
			_, err := stubCase.CreateStubCase(set)

			// Validate error behavior
			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}
			}
		})
	}
}

func buildTestCases(withCaseNo bool, docType string, payload string) []requestCaseStub {
	tests := []requestCaseStub{}

	if withCaseNo {
		if docType == "COPORD" {
			tests = append(tests, requestCaseStub{
				name:       "Order Case with CaseNo",
				xmlPayload: payload,
				expectedReq: &types.ScannedCaseRequest{
					CourtReference: "123",
					CaseType:       "order",
				},
				expectedErr: nil,
			})
		} else {
			tests = append(tests, requestCaseStub{
				name:        "Invalid CaseNo for non-COPORD",
				xmlPayload:  payload,
				expectedReq: nil,
				expectedErr: fmt.Errorf("expected error"),
			})
		}
	} else {
		if docType == "LPA002" || docType == "LP1F" || docType == "LP1H" || docType == "LP2" {
			tests = append(tests, requestCaseStub{
				name:       "LPA Case without CaseNo",
				xmlPayload: payload,
				expectedReq: &types.ScannedCaseRequest{
					CaseType: "lpa",
				},
				expectedErr: nil,
			})
		} else if docType == "EP2PG" || docType == "EPA" {
			tests = append(tests, requestCaseStub{
				name:       "EPA Case without CaseNo",
				xmlPayload: payload,
				expectedReq: &types.ScannedCaseRequest{
					CaseType: "epa",
				},
				expectedErr: nil,
			})
		} else {
			tests = append(tests, requestCaseStub{
				name:        "Invalid Document Type without CaseNo",
				xmlPayload:  payload,
				expectedReq: &types.ScannedCaseRequest{},
				expectedErr: fmt.Errorf("expected error"),
			})
		}
	}

	return tests
}
