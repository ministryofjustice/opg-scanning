package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	if scannedCaseRequest == (types.ScannedCaseRequest{}) {
		return nil, fmt.Errorf("could not determine case type")
	}

	return requestCreateScannedCase(url, scannedCaseRequest)
}

func requestCreateScannedCase(url string, reqData types.ScannedCaseRequest) (*types.ScannedCaseResponse, error) {
	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	responseBody, err := httpRequest(url+"/scanned-case", "POST", string(body))
	if err != nil {
		return nil, err
	}

	scannedResponse := types.ScannedCaseResponse{}
	err = json.Unmarshal(responseBody, &scannedResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &scannedResponse, nil
	//return &types.ScannedCaseResponse{UID: "dummy-uid-1234"}, nil
}

func httpRequest(url string, method string, payload string) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBufferString(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non 2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil

}
