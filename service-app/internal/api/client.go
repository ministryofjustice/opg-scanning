package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
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
func (c *Client) CreateCaseStub(ctx context.Context, set types.BaseSet) (*types.ScannedCaseResponse, error) {
	scannedCaseRequest, err := determineCaseRequest(set)
	if err != nil {
		return nil, err
	}

	if scannedCaseRequest == nil && set.Header.CaseNo != "" {
		return &types.ScannedCaseResponse{
			UID: set.Header.CaseNo,
		}, nil
	}

	return c.requestCreateScannedCase(ctx, scannedCaseRequest)
}

func determineCaseRequest(set types.BaseSet) (*types.ScannedCaseRequest, error) {
	now := time.Now().Format(time.RFC3339)

	parsedScanTime, err := time.Parse("2006-01-02 15:04:05", set.Header.ScanTime)
	if err != nil {
		return nil, fmt.Errorf("invalid ScanTime format: %w", err)
	}
	// Add timezone (UTC) and format as ISO 8601
	formattedScanTime := parsedScanTime.UTC().Format(time.RFC3339)

	for _, doc := range set.Body.Documents {
		if util.Contains(constants.LPATypeDocuments, doc.Type) {
			return &types.ScannedCaseRequest{
				BatchID:     set.Header.Schedule,
				CaseType:    "lpa",
				ReceiptDate: formattedScanTime,
				CreatedDate: now,
			}, nil
		} else if util.Contains(constants.EPATypeDocuments, doc.Type) {
			return &types.ScannedCaseRequest{
				BatchID:     set.Header.Schedule,
				CaseType:    "epa",
				ReceiptDate: formattedScanTime,
				CreatedDate: now,
			}, nil
		} else if doc.Type == constants.DocumentTypeCOPORD && set.Header.CaseNo != "" {
			return &types.ScannedCaseRequest{
				CourtReference: set.Header.CaseNo,
				BatchID:        set.Header.Schedule,
				CaseType:       "order",
				ReceiptDate:    formattedScanTime,
			}, nil
		}
	}

	return nil, nil
}

func (c *Client) requestCreateScannedCase(ctx context.Context, reqData *types.ScannedCaseRequest) (*types.ScannedCaseResponse, error) {
	if reqData == nil {
		return nil, fmt.Errorf("request data is nil")
	}

	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	url := fmt.Sprintf("%s/%s", c.Middleware.Config.App.SiriusBaseURL, c.Middleware.Config.App.SiriusScanURL)

	responseBody, err := c.Middleware.HTTPRequest(ctx, url, "POST", body, nil)
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
