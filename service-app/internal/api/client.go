package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type Client struct {
	Middleware *httpclient.Middleware
}

func NewClient(middleware *httpclient.Middleware) *Client {
	return &Client{
		Middleware: middleware,
	}
}

// Determines case type and sends the request to Sirius
func (c *Client) CreateCaseStub(set types.BaseSet) (*types.ScannedCaseResponse, error) {
	scannedCaseRequest, err := determineCaseRequest(set)
	if err != nil {
		return nil, fmt.Errorf("failed to determine case type: %w", err)
	}

	return c.requestCreateScannedCase(scannedCaseRequest)
}

func determineCaseRequest(set types.BaseSet) (*types.ScannedCaseRequest, error) {
	now := time.Now().Format(time.RFC3339)

	parsedScanTime, err := time.Parse("2006-01-02T15:04:05", set.Header.ScanTime)
	if err != nil {
		return nil, fmt.Errorf("invalid ScanTime format: %w", err)
	}
	// Add timezone (UTC) and format as ISO 8601
	formattedScanTime := parsedScanTime.UTC().Format(time.RFC3339)

	for _, doc := range set.Body.Documents {
		switch doc.Type {
		case "LPA002", "LP1F", "LP1H", "LP2":
			return &types.ScannedCaseRequest{
				BatchID:     set.Header.Schedule,
				CaseType:    "lpa",
				ReceiptDate: formattedScanTime,
				CreatedDate: now,
			}, nil
		case "EP2PG", "EPA":
			return &types.ScannedCaseRequest{
				BatchID:     set.Header.Schedule,
				CaseType:    "epa",
				ReceiptDate: formattedScanTime,
				CreatedDate: now,
			}, nil
		case "COPORD":
			if set.Header.CaseNo != "" {
				return &types.ScannedCaseRequest{
					CourtReference: set.Header.CaseNo,
					BatchID:        set.Header.Schedule,
					CaseType:       "order",
					ReceiptDate:    formattedScanTime,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("could not determine case type")
}

func (c *Client) requestCreateScannedCase(reqData *types.ScannedCaseRequest) (*types.ScannedCaseResponse, error) {
	if reqData == nil {
		return nil, fmt.Errorf("request data is nil")
	}

	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	url := fmt.Sprintf("%s/%s", c.Middleware.Client.Config.App.SiriusBaseURL, c.Middleware.Client.Config.App.SiriusScanURL)

	responseBody, err := c.Middleware.HTTPRequest(url, "POST", body, nil)
	if err != nil {
		return nil, fmt.Errorf("request to Sirius API failed: %w", err)
	}

	var scannedResponse types.ScannedCaseResponse
	err = json.Unmarshal(responseBody, &scannedResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &scannedResponse, nil
}
