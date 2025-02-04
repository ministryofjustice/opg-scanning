package api

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/ingestion"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var xmlPayload = `
<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
    <Header CaseNo="" Scanner="9" ScanTime="2014-09-26T12:38:53" ScannerOperator="Administrator" Schedule="02-0001112-20160909185000" FeeNumber="1234"/>
    <Body>
        <Document Type="LP1F" Encoding="UTF-8" NoPages="19">
            <XML>SGVsbG8gd29ybGQ=</XML>
            <PDF>SGVsbG8gd29ybGQ=</PDF>
        </Document>
    </Body>
</Set>
`

func setupController() *IndexController {
	appConfig := config.NewConfig()
	logger := logger.NewLogger(appConfig)

	// Create mock dependencies
	mockHttpClient, mockAuthMiddleware, awsClient, tokenGenerator := auth.PrepareMocks(appConfig, logger)
	httpMiddleware, _ := httpclient.NewMiddleware(mockHttpClient, tokenGenerator)

	mockHttpClient.On("HTTPRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]byte(`{"UID": "dummy-uid-1234"}`), nil)

	return &IndexController{
		config:         appConfig,
		logger:         logger,
		validator:      ingestion.NewValidator(),
		httpMiddleware: httpMiddleware,
		authMiddleware: mockAuthMiddleware,
		Queue:          ingestion.NewJobQueue(appConfig),
		AwsClient:      awsClient,
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

func TestProcessAndPersist_IncludesXMLDeclaration(t *testing.T) {
	controller := setupController()

	var setPayload types.BaseSet
	err := xml.Unmarshal([]byte(xmlPayload), &setPayload)
	require.NoError(t, err, "failed to unmarshal xmlPayload")

	originalDoc := &types.BaseDocument{
		Type: "LP1F",
	}

	var capturedXML []byte

	mockAws := new(aws.MockAwsClient)
	controller.AwsClient = mockAws

	// Set expectation on PersistFormData.
	mockAws.
		On("PersistFormData", mock.Anything, mock.AnythingOfType("*bytes.Reader"), "LP1F").
		Return("testFileName", nil).
		Run(func(args mock.Arguments) {
			bodyReader := args.Get(1).(io.Reader)
			var err error
			capturedXML, err = io.ReadAll(bodyReader)
			require.NoError(t, err)
		})
	fileName, err := controller.processAndPersist(context.Background(), setPayload, originalDoc)
	require.NoError(t, err)
	require.Equal(t, "testFileName", fileName)

	// Verify that the captured XML starts with the XML declaration.
	expectedHeader := `<?xml version="1.0" encoding="UTF-8" standalone="no"?>`
	if !strings.HasPrefix(string(capturedXML), expectedHeader) {
		t.Errorf("expected XML to begin with header %q, got: %s", expectedHeader, string(capturedXML))
	}
}
