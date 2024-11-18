package api

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

const (
	LPAPayload = `<?xml version="1.0" encoding="UTF-8"?>
<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
    <Header CaseNo="" Scanner="9" ScanTime="2014-09-26T12:38:53" ScannerOperator="Administrator" Schedule="02-0001112-20160909185000" FeeNumber="1234"/>
    <Body>
        <Document Type="LP1F" Encoding="UTF-8" NoPages="19"></Document>
    </Body>
</Set>`

	EPAPayload = `<?xml version="1.0" encoding="UTF-8"?>
<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
    <Header CaseNo="" Scanner="9" ScanTime="2014-09-26T12:38:53" ScannerOperator="Administrator" Schedule="02-0002222-20160909185000" FeeNumber="5678"/>
    <Body>
        <Document Type="EPA" Encoding="UTF-8" NoPages="10"></Document>
    </Body>
</Set>`

	OrderPayload = `<?xml version="1.0" encoding="UTF-8"?>
<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
    <Header CaseNo="12345" Scanner="9" ScanTime="2014-09-26T12:38:53" ScannerOperator="Administrator" Schedule="02-0003333-20160909185000" FeeNumber="91011"/>
    <Body>
        <Document Type="COPORD" Encoding="UTF-8" NoPages="5"></Document>
    </Body>
</Set>`
)

func TestCreateStubCase(t *testing.T) {
	tests := []struct {
		name        string
		xmlPayload  string
		expectedReq types.ScannedCaseRequest
	}{
		{
			name:       "LPA Case",
			xmlPayload: LPAPayload,
			expectedReq: types.ScannedCaseRequest{
				BatchID:     "02-0001112-20160909185000",
				CaseType:    "lpa",
				ReceiptDate: "2014-09-26T12:38:53",
				CreatedDate: time.Now().Format(time.RFC3339), // Dynamically set
			},
		},
		{
			name:       "EPA Case",
			xmlPayload: EPAPayload,
			expectedReq: types.ScannedCaseRequest{
				BatchID:     "02-0002222-20160909185000",
				CaseType:    "epa",
				ReceiptDate: "2014-09-26T12:38:53",
				CreatedDate: time.Now().Format(time.RFC3339),
			},
		},
		{
			name:       "Order Case",
			xmlPayload: OrderPayload,
			expectedReq: types.ScannedCaseRequest{
				CourtReference: "12345",
				BatchID:        "02-0003333-20160909185000",
				CaseType:       "order",
				ReceiptDate:    "2014-09-26T12:38:53",
			},
		},
	}

	var wg sync.WaitGroup

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg.Add(1)

			// Parse XML payload into types.BaseSet
			var set types.BaseSet
			if err := xml.Unmarshal([]byte(tt.xmlPayload), &set); err != nil {
				t.Fatalf("failed to parse XML payload: %v", err)
			}

			// Mock server to simulate /scanned-case endpoint and validate request body
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer wg.Done()

				if r.URL.Path != "/scanned-case" {
					t.Fatalf("unexpected URL path: %s", r.URL.Path)
				}

				var receivedRequest types.ScannedCaseRequest
				if err := json.NewDecoder(r.Body).Decode(&receivedRequest); err != nil {
					t.Fatalf("failed to decode request body: %v", err)
				}

				if !reflect.DeepEqual(receivedRequest, tt.expectedReq) {
					t.Errorf("received request does not match expected request.\nReceived: %+v\nExpected: %+v", receivedRequest, tt.expectedReq)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"UID": "dummy-uid-1234"}`))
			}))
			defer mockServer.Close()

			_, err := CreateStubCase(mockServer.URL, set)
			if err != nil {
				t.Fatalf("failed to create stub case: %v", err)
			}
		})
	}

	wg.Wait()
}
