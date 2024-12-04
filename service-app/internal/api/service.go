package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type ServiceInterface interface {
	AttachDocuments(ctx context.Context, set types.BaseSet) ([]types.ScannedDocumentResponse, error)
	CreateCaseStub(ctx context.Context, set types.BaseSet) (*types.ScannedCaseResponse, error)
}

type Service struct {
	Client *Client
}

func NewService(client *Client) *Service {
	return &Service{Client: client}
}

// Attach documents to cases
func (s *Service) AttachDocuments(ctx context.Context, set types.BaseSet) ([]types.ScannedDocumentResponse, error) {
	responses := []types.ScannedDocumentResponse{}

	for _, doc := range set.Body.Documents {
		var documentSubType string

		// Check for Correspondence or SupCorrespondence and extract SubType
		if util.Contains([]string{"Correspondence", "SupCorrespondence"}, doc.Type) {
			subType, err := DecodeAndExtractSubType(doc.EmbeddedXML)
			if err != nil {
				return nil, fmt.Errorf("failed to extract SubType for document %s: %w", doc.Type, err)
			}
			documentSubType = subType
		}

		// Prepare the request payload
		request := types.ScannedDocumentRequest{
			CaseReference:   set.Header.CaseNo,
			Content:         doc.EmbeddedXML,
			DocumentType:    doc.Type,
			DocumentSubType: documentSubType,
			ScannedDate:     formatScannedDate(set.Header.ScanTime),
		}

		// Send the request
		url := fmt.Sprintf("%s/scanned-documents", s.Client.Middleware.Config.App.SiriusBaseURL)

		resp, err := s.Client.ClientRequest(ctx, request, url)
		if err != nil {
			return nil, fmt.Errorf("failed to attach document %s: %w", doc.Type, err)
		}

		var scannedResponse types.ScannedDocumentResponse
		err = json.Unmarshal(*resp, &scannedResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		responses = append(responses, scannedResponse)
	}

	return responses, nil
}

// Create a case stub
func (s *Service) CreateCaseStub(ctx context.Context, set types.BaseSet) (*types.ScannedCaseResponse, error) {
	scannedCaseRequest, err := determineCaseRequest(set)
	if err != nil {
		return nil, err
	}

	if scannedCaseRequest == nil && set.Header.CaseNo == "" {
		return nil, fmt.Errorf("CaseNo cannot be empty with unmatched document type")
	}

	if scannedCaseRequest == nil && set.Header.CaseNo != "" {
		return &types.ScannedCaseResponse{
			UID: set.Header.CaseNo,
		}, nil
	}

	url := fmt.Sprintf("%s/%s", s.Client.Middleware.Config.App.SiriusBaseURL, s.Client.Middleware.Config.App.SiriusScanURL)

	resp, err := s.Client.ClientRequest(ctx, scannedCaseRequest, url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Sirius: %w", err)
	}

	var scannedResponse types.ScannedCaseResponse
	err = json.Unmarshal(*resp, &scannedResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &scannedResponse, nil
}
