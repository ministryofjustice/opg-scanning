package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser/corresp_parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type ServiceInterface interface {
	AttachDocuments(ctx context.Context, set types.BaseSet) (*types.ScannedDocumentResponse, error)
	CreateCaseStub(ctx context.Context, set types.BaseSet) (*types.ScannedCaseResponse, error)
}

type Service struct {
	Client      *Client
	set         *types.BaseSet
	originalDoc *types.BaseDocument
}

func NewService(client *Client, set *types.BaseSet) *Service {
	return &Service{
		Client: client,
		set:    set,
	}
}

// Attach documents to cases
func (s *Service) AttachDocuments(ctx context.Context, caseResponse *types.ScannedCaseResponse) (*types.ScannedDocumentResponse, error) {
	var documentSubType string

	// Check for Correspondence or SupCorrespondence and extract SubType
	if util.Contains([]string{"Correspondence", "SupCorrespondence"}, s.originalDoc.Type) {
		// Decode the base64-encoded XML
		decodedXML, err := base64.StdEncoding.DecodeString(s.originalDoc.EmbeddedXML)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64-encoded XML: %w", err)
		}
		// Parse the XML
		correspDoc, err := corresp_parser.Parse([]byte(decodedXML))
		if err != nil {
			return nil, fmt.Errorf("failed to parse correspondence for document %s: %w", s.originalDoc.Type, err)
		}
		documentSubType = correspDoc.SubType
	}

	// Prepare the request payload
	request := types.ScannedDocumentRequest{
		CaseReference:   caseResponse.UID,
		Content:         s.originalDoc.EmbeddedPDF,
		DocumentType:    s.originalDoc.Type,
		DocumentSubType: documentSubType,
		ScannedDate:     formatScannedDate(s.set.Header.ScanTime),
	}

	// Send the request
	url := fmt.Sprintf("%s/%s", s.Client.Middleware.Config.App.SiriusBaseURL, s.Client.Middleware.Config.App.SiriusAttachDocURL)

	resp, err := s.Client.ClientRequest(ctx, request, url)
	if err != nil {
		return nil, fmt.Errorf("failed to attach document %s: %w", s.originalDoc.Type, err)
	}

	var scannedResponse types.ScannedDocumentResponse
	err = json.Unmarshal(*resp, &scannedResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &scannedResponse, nil
}

// Create a case stub
func (s *Service) CreateCaseStub(ctx context.Context) (*types.ScannedCaseResponse, error) {
	scannedCaseRequest, err := determineCaseRequest(s.set)
	if err != nil {
		return nil, err
	}

	if scannedCaseRequest == nil && s.set.Header.CaseNo == "" {
		return nil, fmt.Errorf("CaseNo cannot be empty with unmatched document type")
	}

	if scannedCaseRequest == nil && s.set.Header.CaseNo != "" {
		return &types.ScannedCaseResponse{
			UID: s.set.Header.CaseNo,
		}, nil
	}

	url := fmt.Sprintf("%s/%s", s.Client.Middleware.Config.App.SiriusBaseURL, s.Client.Middleware.Config.App.SiriusCaseStubURL)

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