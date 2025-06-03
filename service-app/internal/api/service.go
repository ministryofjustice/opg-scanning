package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"

	"github.com/ministryofjustice/opg-scanning/internal/parser/corresp_parser"
	"github.com/ministryofjustice/opg-scanning/internal/sirius"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/corresp_types"
)

type SiriusClient interface {
	AttachDocument(ctx context.Context, data *sirius.ScannedDocumentRequest) (*sirius.ScannedDocumentResponse, error)
	CreateCaseStub(ctx context.Context, data *sirius.ScannedCaseRequest) (*sirius.ScannedCaseResponse, error)
}

type service struct {
	siriusClient SiriusClient
	set          *types.BaseSet
	originalDoc  *types.BaseDocument
}

func newService(client SiriusClient, set *types.BaseSet) *service {
	return &service{
		siriusClient: client,
		set:          set,
	}
}

// Attach documents to cases
func (s *service) AttachDocuments(ctx context.Context, caseResponse *sirius.ScannedCaseResponse) (*sirius.ScannedDocumentResponse, []byte, error) {
	var documentSubType string
	var originalDocType = s.originalDoc.Type

	// Decode the base64-encoded XML
	decodedXML, err := base64.StdEncoding.DecodeString(s.originalDoc.EmbeddedXML)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode base64-encoded XML: %w", err)
	}

	// Check for Correspondence or SupCorrespondence and extract SubType
	if slices.Contains([]string{"Correspondence", "SupCorrespondence"}, originalDocType) {
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
	request := &sirius.ScannedDocumentRequest{
		CaseReference:   caseResponse.UID,
		Content:         s.originalDoc.EmbeddedPDF,
		DocumentType:    originalDocType,
		DocumentSubType: documentSubType,
		ScannedDate:     formatScannedDate(s.set.Header.ScanTime),
	}

	// Send the request
	scannedResponse, err := s.siriusClient.AttachDocument(ctx, request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to attach document %s: %w", s.originalDoc.Type, err)
	}

	return scannedResponse, decodedXML, nil
}

// Create a case stub
func (s *service) CreateCaseStub(ctx context.Context) (*sirius.ScannedCaseResponse, error) {
	scannedCaseRequest, err := determineCaseRequest(s.set)
	if err != nil {
		return nil, err
	}

	if scannedCaseRequest == nil && s.set.Header.CaseNo == "" {
		return nil, fmt.Errorf("CaseNo cannot be empty with unmatched document type")
	}

	if scannedCaseRequest == nil && s.set.Header.CaseNo != "" {
		return &sirius.ScannedCaseResponse{
			UID: s.set.Header.CaseNo,
		}, nil
	}

	scannedResponse, err := s.siriusClient.CreateCaseStub(ctx, scannedCaseRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Sirius: %w", err)
	}

	return scannedResponse, nil
}
