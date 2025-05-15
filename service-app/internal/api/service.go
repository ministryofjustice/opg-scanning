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

type service struct {
	Client      *client
	set         *types.BaseSet
	originalDoc *types.BaseDocument
}

func newService(client *client, set *types.BaseSet) *service {
	return &service{
		Client: client,
		set:    set,
	}
}

// Attach documents to cases
func (s *service) AttachDocuments(ctx context.Context, caseResponse *types.ScannedCaseResponse) (*types.ScannedDocumentResponse, []byte, error) {
	var documentSubType string
	var originalDocType = s.originalDoc.Type

	// Decode the base64-encoded XML
	decodedXML, err := base64.StdEncoding.DecodeString(s.originalDoc.EmbeddedXML)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode base64-encoded XML: %w", err)
	}

	// Check for Correspondence or SupCorrespondence and extract SubType
	if util.Contains([]string{"Correspondence", "SupCorrespondence"}, originalDocType) {
		// Parse the XML
		correspInterface, err := corresp_parser.Parse([]byte(decodedXML))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse correspondence for document %s: %w", originalDocType, err)
		}
		corresp, ok := correspInterface.(*corresp_types.Correspondence)
		if !ok {
			return nil, nil, fmt.Errorf("failed to cast correspInterface to corresp_types.Correspondence")
		}
		documentSubType = corresp.SubType
	}

	// Prepare the request payload
	request := types.ScannedDocumentRequest{
		CaseReference:   caseResponse.UID,
		Content:         s.originalDoc.EmbeddedPDF,
		DocumentType:    originalDocType,
		DocumentSubType: documentSubType,
		ScannedDate:     formatScannedDate(s.set.Header.ScanTime),
	}

	// Send the request
	url := fmt.Sprintf("%s/%s", s.Client.Middleware.Config.App.SiriusBaseURL, s.Client.Middleware.Config.App.SiriusAttachDocURL)

	resp, err := s.Client.clientRequest(ctx, request, url)
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
func (s *service) CreateCaseStub(ctx context.Context) (*types.ScannedCaseResponse, error) {
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

	resp, err := s.Client.clientRequest(ctx, scannedCaseRequest, url)
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
