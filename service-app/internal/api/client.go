package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type CreateStubCase struct {
	config     config.Config
	logger     logger.Logger
	httpClient *httpclient.HttpClient
}

func NewCreateStubCase(httpClient *httpclient.HttpClient) *CreateStubCase {
	return &CreateStubCase{
		config:     httpClient.Config,
		logger:     httpClient.Logger,
		httpClient: httpClient,
	}
}

// Determines case type and sends the request to Sirius
func (s CreateStubCase) CreateStubCase(set types.BaseSet) (*types.ScannedCaseResponse, error) {
	scannedCaseRequest, err := s.determineCaseRequest(set)
	if err != nil {
		return nil, fmt.Errorf("failed to determine case type: %w", err)
	}

	return s.requestCreateScannedCase(scannedCaseRequest)
}

func (s CreateStubCase) determineCaseRequest(set types.BaseSet) (*types.ScannedCaseRequest, error) {
	now := time.Now().Format(time.RFC3339)

	for _, doc := range set.Body.Documents {
		switch doc.Type {
		case "LPA002", "LP1F", "LP1H", "LP2":
			return &types.ScannedCaseRequest{
				BatchID:     set.Header.Schedule,
				CaseType:    "lpa",
				ReceiptDate: set.Header.ScanTime,
				CreatedDate: now,
			}, nil
		case "EP2PG", "EPA":
			return &types.ScannedCaseRequest{
				BatchID:     set.Header.Schedule,
				CaseType:    "epa",
				ReceiptDate: set.Header.ScanTime,
				CreatedDate: now,
			}, nil
		case "COPORD":
			if set.Header.CaseNo != "" {
				return &types.ScannedCaseRequest{
					CourtReference: set.Header.CaseNo,
					BatchID:        set.Header.Schedule,
					CaseType:       "order",
					ReceiptDate:    set.Header.ScanTime,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("could not determine case type")
}

func (s CreateStubCase) requestCreateScannedCase(reqData *types.ScannedCaseRequest) (*types.ScannedCaseResponse, error) {
	if reqData == nil {
		return nil, fmt.Errorf("request data is nil")
	}

	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	url := fmt.Sprintf("%s/%s", s.config.App.SiriusBaseURL, s.config.App.SiriusScanURL)

	// TODO: we need to include auth / middleware to handle token inclusion

	responseBody, err := s.httpClient.HTTPRequest(url, "POST", body, nil)
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
