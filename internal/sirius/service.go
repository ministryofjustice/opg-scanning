package sirius

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"

	"github.com/ministryofjustice/opg-scanning/internal/config"
	"github.com/ministryofjustice/opg-scanning/internal/parser/corresp_parser"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/types/corresp_types"
)

type siriusClient interface {
	AttachDocument(ctx context.Context, data *ScannedDocumentRequest) (*ScannedDocumentResponse, error)
	CreateCaseStub(ctx context.Context, data *ScannedCaseRequest) (*ScannedCaseResponse, error)
}

type Service struct {
	client siriusClient
}

func NewService(config *config.Config) *Service {
	return &Service{
		client: NewClient(config),
	}
}

func (s *Service) AttachDocuments(ctx context.Context, set *types.BaseSet, originalDoc *types.BaseDocument, caseResponse *ScannedCaseResponse) (*ScannedDocumentResponse, []byte, error) {
	documentSubType := ""
	originalDocType := originalDoc.Type

	decodedXML, err := base64.StdEncoding.DecodeString(originalDoc.EmbeddedXML)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode base64-encoded XML: %w", err)
	}

	// Check for Correspondence or SupCorrespondence and extract SubType
	if slices.Contains([]string{"Correspondence", "SupCorrespondence"}, originalDocType) {
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

	request := &ScannedDocumentRequest{
		CaseReference:   caseResponse.UID,
		Content:         originalDoc.EmbeddedPDF,
		DocumentType:    originalDocType,
		DocumentSubType: documentSubType,
		ScannedDate:     formatScannedDate(set.Header.ScanTime),
	}

	scannedResponse, err := s.client.AttachDocument(ctx, request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to attach document %s: %w", originalDoc.Type, err)
	}

	return scannedResponse, decodedXML, nil
}

func (s *Service) CreateCaseStub(ctx context.Context, set *types.BaseSet) (*ScannedCaseResponse, error) {
	scannedCaseRequest, err := determineCaseRequest(set)
	if err != nil {
		return nil, err
	}

	if scannedCaseRequest == nil {
		if set.Header.CaseNo == "" {
			return nil, fmt.Errorf("CaseNo cannot be empty with unmatched document type")
		} else {
			return &ScannedCaseResponse{UID: set.Header.CaseNo}, nil
		}
	}

	scannedResponse, err := s.client.CreateCaseStub(ctx, scannedCaseRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Sirius: %w", err)
	}

	return scannedResponse, nil
}
