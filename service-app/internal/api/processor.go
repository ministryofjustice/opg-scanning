package api

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type Correspondence struct {
	SubType string `xml:"SubType"`
}

func DecodeAndExtractSubType(encodedXML string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedXML)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 content: %w", err)
	}

	var correspondence Correspondence
	err = xml.Unmarshal(decodedBytes, &correspondence)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal XML content: %w", err)
	}

	return correspondence.SubType, nil
}

func determineCaseRequest(set types.BaseSet) (*types.ScannedCaseRequest, error) {
	now := time.Now().Format(time.RFC3339)

	for _, doc := range set.Body.Documents {
		if util.Contains(constants.LPATypeDocuments, doc.Type) {
			return &types.ScannedCaseRequest{
				BatchID:     set.Header.Schedule,
				CaseType:    "lpa",
				ReceiptDate: formatScannedDate(set.Header.ScanTime),
				CreatedDate: now,
			}, nil
		} else if util.Contains(constants.EPATypeDocuments, doc.Type) {
			return &types.ScannedCaseRequest{
				BatchID:     set.Header.Schedule,
				CaseType:    "epa",
				ReceiptDate: formatScannedDate(set.Header.ScanTime),
				CreatedDate: now,
			}, nil
		} else if doc.Type == constants.DocumentTypeCOPORD && set.Header.CaseNo != "" {
			return &types.ScannedCaseRequest{
				CourtReference: set.Header.CaseNo,
				BatchID:        set.Header.Schedule,
				CaseType:       "order",
				ReceiptDate:    formatScannedDate(set.Header.ScanTime),
			}, nil
		}
	}

	return nil, nil
}

func formatScannedDate(scanTime string) string {
	parsedScanTime, err := time.Parse("2006-01-02 15:04:05", scanTime)
	if err != nil {
		return time.Now().UTC().Format(time.RFC3339)
	}
	return parsedScanTime.UTC().Format(time.RFC3339)
}
