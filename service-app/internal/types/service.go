package types

type ScannedCaseRequest struct {
	BatchID        string `json:"batchId"`
	CaseType       string `json:"caseType"`
	CourtReference string `json:"courtReference,omitempty"`
	ReceiptDate    string `json:"receiptDate"`
	CreatedDate    string `json:"createdDate"`
}

type ScannedCaseResponse struct {
	UID string `json:"uid"`
}

type ScannedDocumentRequest struct {
	CaseReference   string `json:"caseReference"`
	Content         string `json:"content"`
	DocumentType    string `json:"documentType"`
	DocumentSubType string `json:"documentSubType,omitempty"`
	ScannedDate     string `json:"scannedDate"`
}

type ScannedDocumentResponse struct {
	ID                  int    `json:"id"`
	UUID                string `json:"uuid"`
	Type                string `json:"type"`
	FriendlyDescription string `json:"friendlyDescription"`
	Title               string `json:"title"`
	SourceDocumentType  string `json:"sourceDocumentType"`
	Subtype             string `json:"subtype"`
}

// For handling error responses according to the OpenAPI spec
type ErrorResponse struct {
	Type             string            `json:"type"`
	Title            string            `json:"title"`
	Status           string            `json:"status"`
	Detail           string            `json:"detail"`
	ValidationErrors map[string]string `json:"validation_errors"`
}
