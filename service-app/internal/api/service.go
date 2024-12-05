package api

import (
	"context"
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
	Client       *Client
	set          *types.BaseSet
	originalDoc  *types.BaseDocument
	processedDoc interface{}
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

	// Encode parsed document and replace the embedded XML
	if s.processedDoc != nil {
		encoded, err := json.Marshal(s.processedDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to encode parsed document: %w", err)
		}
		s.originalDoc.EmbeddedXML = string(encoded)
	}

	// Check for Correspondence or SupCorrespondence and extract SubType
	if util.Contains([]string{"Correspondence", "SupCorrespondence"}, s.originalDoc.Type) {
		correspDoc, err := corresp_parser.Parse([]byte(s.originalDoc.EmbeddedXML))
		if err != nil {
			return nil, fmt.Errorf("failed to extract SubType for document %s: %w", s.originalDoc.Type, err)
		}
		documentSubType = correspDoc.SubType
	}

	// Prepare the request payload
	request := types.ScannedDocumentRequest{
		CaseReference:   caseResponse.UID,
		Content:         s.originalDoc.EmbeddedXML,
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
