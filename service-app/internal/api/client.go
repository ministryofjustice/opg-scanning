package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

func CreateStubCase(url string, set types.BaseSet) (*types.ScannedCaseResponse, error) {
	var scannedCaseRequest types.ScannedCaseRequest
	now := time.Now().Format(time.RFC3339)

	if set.Header.CaseNo == "" {
		// Check for LPA cases
		for _, doc := range set.Body.Documents {
			if doc.Type == "LPA002" || doc.Type == "LP1F" || doc.Type == "LP1H" || doc.Type == "LP2" {
				// Create a new LPA case
				scannedCaseRequest = types.ScannedCaseRequest{
					BatchID:     set.Header.Schedule,
					CaseType:    "lpa",
					ReceiptDate: set.Header.ScanTime,
					CreatedDate: now,
				}
				break
			} else if doc.Type == "EP2PG" || doc.Type == "EPA" {
				// Create a new EPA case
				scannedCaseRequest = types.ScannedCaseRequest{
					BatchID:     set.Header.Schedule,
					CaseType:    "epa",
					ReceiptDate: set.Header.ScanTime,
					CreatedDate: now,
				}
				break
			}
		}
	} else if set.Header.CaseNo != "" {
		// Check for COPORD case with CaseNo
		for _, doc := range set.Body.Documents {
			if doc.Type == "COPORD" {
				scannedCaseRequest = types.ScannedCaseRequest{
					CourtReference: set.Header.CaseNo,
					BatchID:        set.Header.Schedule,
					CaseType:       "order",
					ReceiptDate:    set.Header.ScanTime,
				}
				break
			}
		}
	}

	return requestCreateScannedCase(url, scannedCaseRequest)
}

func requestCreateScannedCase(url string, reqData types.ScannedCaseRequest) (*types.ScannedCaseResponse, error) {
	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	req, err := http.NewRequest("POST", url+"/scanned-cases", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request (dummy for testing purposes)
	// Placeholder for actual HTTP call
	// resp, err := http.DefaultClient.Do(req)
	// Instead return a dummy UUID for now
	return &types.ScannedCaseResponse{UUID: "dummy-uuid-1234"}, nil
}
