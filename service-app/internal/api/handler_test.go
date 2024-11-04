package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/ingestion"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

var xmlPayload = `
<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
    <Header CaseNo="" Scanner="9" ScanTime="2014-09-26T12:38:53" ScannerOperator="Administrator" Schedule="02-0001112-20160909185000" FeeNumber="1234"/>
    <Body>
        <Document Type="LP1F" Encoding="UTF-8" NoPages="19">
            <XML>SGVsbG8gd29ybGQ=</XML>
            <Image>SGVsbG8gd29ybGQ=</Image>
        </Document>
    </Body>
</Set>
`

// Helper to create an IndexController instance for testing
func setupController() *IndexController {
	return &IndexController{
		config:    config.NewConfig(),
		validator: ingestion.NewValidator(),
		Queue:     ingestion.NewJobQueue(),
		logger:    logger.NewLogger(),
	}
}

func TestIngestHandler_SetValid(t *testing.T) {
	controller := setupController()

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer([]byte(xmlPayload)))
	req.Header.Set("Content-Type", "application/xml")
	w := httptest.NewRecorder()

	controller.IngestHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected status %d; got %d", http.StatusAccepted, resp.StatusCode)
	}
}

func TestIngestHandler_InvalidContentType(t *testing.T) {
	controller := setupController()

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer([]byte(xmlPayload)))
	// Invalid Content-Type
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	controller.IngestHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d; got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestIngestHandler_InvalidXML(t *testing.T) {
	controller := setupController()

	xmlPayloadMalformed := `<Set>
		<Header CaseNo="1234"><Body></Body>`

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer([]byte(xmlPayloadMalformed)))
	req.Header.Set("Content-Type", "application/xml")
	w := httptest.NewRecorder()

	controller.IngestHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d; got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// TODO: Complete this test
func TestIngestHandler_QueueingDocuments(t *testing.T) {
	controller := setupController()

	// Mock valid XML body with multiple documents
	xmlPayloadQueued := `<Set>
		<Header CaseNo="1234" Scanner="Scanner1" ScanTime="2021-09-01T12:34:56" ScannerOperator="Operator" />
		<Body>
			<Document Type="LP1F" Encoding="UTF-8" NoPages="6">
				<XML>...base64encodedXML...</XML>
				<Image>...base64encodedImage...</Image>
			</Document>
			<Document Type="LP1F" Encoding="UTF-8" NoPages="4">
				<XML>...base64encodedXML...</XML>
				<Image>...base64encodedImage...</Image>
			</Document>
		</Body>
	</Set>`

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer([]byte(xmlPayloadQueued)))
	req.Header.Set("Content-Type", "application/xml")
	w := httptest.NewRecorder()

	controller.IngestHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected status %d; got %d", http.StatusAccepted, resp.StatusCode)
	}

	expectedJobs := 2
	actualJobs := len(controller.Queue.Jobs)
	if actualJobs != expectedJobs {
		t.Errorf("expected %d jobs to be queued; got %d", expectedJobs, actualJobs)
	}
}
