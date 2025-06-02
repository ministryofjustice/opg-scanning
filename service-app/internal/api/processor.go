package api

import (
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/sirius"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

func determineCaseRequest(set *types.BaseSet) (*sirius.ScannedCaseRequest, error) {
	now := time.Now().Format(time.RFC3339)

	for _, doc := range set.Body.Documents {
		if util.Contains(constants.CreateLPADocuments, doc.Type) {
			return &sirius.ScannedCaseRequest{
				BatchID:     set.Header.Schedule,
				CaseType:    "lpa",
				ReceiptDate: formatScannedDate(set.Header.ScanTime),
				CreatedDate: now,
			}, nil
		} else if util.Contains(constants.CreateEPADocuments, doc.Type) {
			return &sirius.ScannedCaseRequest{
				BatchID:     set.Header.Schedule,
				CaseType:    "epa",
				ReceiptDate: formatScannedDate(set.Header.ScanTime),
				CreatedDate: now,
			}, nil
		} else if doc.Type == constants.DocumentTypeCOPORD && set.Header.CaseNo != "" {
			return &sirius.ScannedCaseRequest{
				CourtReference: set.Header.CaseNo,
				BatchID:        set.Header.Schedule,
				CaseType:       "order",
				ReceiptDate:    formatScannedDate(set.Header.ScanTime),
				CreatedDate:    now,
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
