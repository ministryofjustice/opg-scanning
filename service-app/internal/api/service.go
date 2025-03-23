package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/parser/corresp_parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/corresp_types"
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
func (s *Service) AttachDocuments(ctx context.Context, caseResponse *types.ScannedCaseResponse) (*types.ScannedDocumentResponse, []byte, error) {
	var documentSubType string
	var mappedDocType string

	// Decode the base64-encoded XML
	decodedXML, err := base64.StdEncoding.DecodeString(s.originalDoc.EmbeddedXML)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode base64-encoded XML: %w", err)
	}

	// Check for Correspondence or SupCorrespondence and extract SubType
	if util.Contains([]string{"Correspondence", "SupCorrespondence"}, s.originalDoc.Type) {
		// Parse the XML
		correspInterface, err := corresp_parser.Parse([]byte(decodedXML))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse correspondence for document %s: %w", s.originalDoc.Type, err)
		}
		corresp, ok := correspInterface.(*corresp_types.Correspondence)
		if !ok {
			return nil, nil, fmt.Errorf("failed to cast correspInterface to corresp_types.Correspondence")
		}
		documentSubType = corresp.SubType
		mappedDocType = deputyDocType(s.originalDoc.Type)
	} else if util.Contains([]string{"DEPREPORTS", "FINDOCS", "DEPCORRES"}, s.originalDoc.Type) {
		// For supervision report types directly map the top-level document type.
		mappedDocType = deputyDocType(s.originalDoc.Type)
	} else {
		// For any other document types, leave as is.
		mappedDocType = s.originalDoc.Type
	}

	// Prepare the request payload
	request := types.ScannedDocumentRequest{
		CaseReference:   caseResponse.UID,
		Content:         s.originalDoc.EmbeddedPDF,
		DocumentType:    mappedDocType,
		DocumentSubType: documentSubType,
		ScannedDate:     formatScannedDate(s.set.Header.ScanTime),
	}

	// Send the request
	url := fmt.Sprintf("%s/%s", s.Client.Middleware.Config.App.SiriusBaseURL, s.Client.Middleware.Config.App.SiriusAttachDocURL)

	resp, err := s.Client.ClientRequest(ctx, request, url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to attach document %s: %w", s.originalDoc.Type, err)
	}

	var scannedResponse types.ScannedDocumentResponse
	err = json.Unmarshal(*resp, &scannedResponse)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &scannedResponse, decodedXML, nil
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

	if s.Client.Middleware == nil {
		return nil, fmt.Errorf("middleware is nil")
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

func deputyDocType(docType string) string {
	switch docType {
	case "FINDOCS":
		return "Report - Financial evidence"
	case "DEPREPORTS":
		return "Report - General"
	case "DEPCORRES":
		return "Report"
	case "SupCorrespondence":
		return "Correspondence"
	default:
		return docType
	}
}
