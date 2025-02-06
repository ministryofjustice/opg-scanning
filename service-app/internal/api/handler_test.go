package api

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/ingestion"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var xmlPayload = `
<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
    <Header CaseNo="" Scanner="9" ScanTime="2014-09-26T12:38:53" ScannerOperator="Administrator" Schedule="02-0001112-20160909185000" FeeNumber="1234"/>
    <Body>
        <Document Type="LP1F" Encoding="UTF-8" NoPages="19">
            <XML>PD94bWwgdmVyc2lvbj0iMS4wIj8+CjxjYXRhbG9nPgogICA8Ym9vayBpZD0iYmsxMDEiPgogICAg
ICA8YXV0aG9yPkdhbWJhcmRlbGxhLCBNYXR0aGV3PC9hdXRob3I+CiAgICAgIDx0aXRsZT5YTUwg
RGV2ZWxvcGVyJ3MgR3VpZGU8L3RpdGxlPgogICAgICA8Z2VucmU+Q29tcHV0ZXI8L2dlbnJlPgog
ICAgICA8cHJpY2U+NDQuOTU8L3ByaWNlPgogICAgICA8cHVibGlzaF9kYXRlPjIwMDAtMTAtMDE8
L3B1Ymxpc2hfZGF0ZT4KICAgICAgPGRlc2NyaXB0aW9uPkFuIGluLWRlcHRoIGxvb2sgYXQgY3Jl
YXRpbmcgYXBwbGljYXRpb25zIAogICAgICB3aXRoIFhNTC48L2Rlc2NyaXB0aW9uPgogICA8L2Jv
b2s+CiAgIDxib29rIGlkPSJiazEwMiI+CiAgICAgIDxhdXRob3I+UmFsbHMsIEtpbTwvYXV0aG9y
PgogICAgICA8dGl0bGU+TWlkbmlnaHQgUmFpbjwvdGl0bGU+CiAgICAgIDxnZW5yZT5GYW50YXN5
PC9nZW5yZT4KICAgICAgPHByaWNlPjUuOTU8L3ByaWNlPgogICAgICA8cHVibGlzaF9kYXRlPjIw
MDAtMTItMTY8L3B1Ymxpc2hfZGF0ZT4KICAgICAgPGRlc2NyaXB0aW9uPkEgZm9ybWVyIGFyY2hp
dGVjdCBiYXR0bGVzIGNvcnBvcmF0ZSB6b21iaWVzLCAKICAgICAgYW4gZXZpbCBzb3JjZXJlc3Ms
IGFuZCBoZXIgb3duIGNoaWxkaG9vZCB0byBiZWNvbWUgcXVlZW4gCiAgICAgIG9mIHRoZSB3b3Js
ZC48L2Rlc2NyaXB0aW9uPgogICA8L2Jvb2s+CiAgIDxib29rIGlkPSJiazEwMyI+CiAgICAgIDxh
dXRob3I+Q29yZXRzLCBFdmE8L2F1dGhvcj4KICAgICAgPHRpdGxlPk1hZXZlIEFzY2VuZGFudDwv
dGl0bGU+CiAgICAgIDxnZW5yZT5GYW50YXN5PC9nZW5yZT4KICAgICAgPHByaWNlPjUuOTU8L3By
aWNlPgogICAgICA8cHVibGlzaF9kYXRlPjIwMDAtMTEtMTc8L3B1Ymxpc2hfZGF0ZT4KICAgICAg
PGRlc2NyaXB0aW9uPkFmdGVyIHRoZSBjb2xsYXBzZSBvZiBhIG5hbm90ZWNobm9sb2d5IAogICAg
ICBzb2NpZXR5IGluIEVuZ2xhbmQsIHRoZSB5b3VuZyBzdXJ2aXZvcnMgbGF5IHRoZSAKICAgICAg
Zm91bmRhdGlvbiBmb3IgYSBuZXcgc29jaWV0eS48L2Rlc2NyaXB0aW9uPgogICA8L2Jvb2s+CiAg
IDxib29rIGlkPSJiazEwNCI+CiAgICAgIDxhdXRob3I+Q29yZXRzLCBFdmE8L2F1dGhvcj4KICAg
ICAgPHRpdGxlPk9iZXJvbidzIExlZ2FjeTwvdGl0bGU+CiAgICAgIDxnZW5yZT5GYW50YXN5PC9n
ZW5yZT4KICAgICAgPHByaWNlPjUuOTU8L3ByaWNlPgogICAgICA8cHVibGlzaF9kYXRlPjIwMDEt
MDMtMTA8L3B1Ymxpc2hfZGF0ZT4KICAgICAgPGRlc2NyaXB0aW9uPkluIHBvc3QtYXBvY2FseXBz
ZSBFbmdsYW5kLCB0aGUgbXlzdGVyaW91cyAKICAgICAgYWdlbnQga25vd24gb25seSBhcyBPYmVy
b24gaGVscHMgdG8gY3JlYXRlIGEgbmV3IGxpZmUgCiAgICAgIGZvciB0aGUgaW5oYWJpdGFudHMg
b2YgTG9uZG9uLiBTZXF1ZWwgdG8gTWFldmUgCiAgICAgIEFzY2VuZGFudC48L2Rlc2NyaXB0aW9u
PgogICA8L2Jvb2s+CiAgIDxib29rIGlkPSJiazEwNSI+CiAgICAgIDxhdXRob3I+Q29yZXRzLCBF
dmE8L2F1dGhvcj4KICAgICAgPHRpdGxlPlRoZSBTdW5kZXJlZCBHcmFpbDwvdGl0bGU+CiAgICAg
IDxnZW5yZT5GYW50YXN5PC9nZW5yZT4KICAgICAgPHByaWNlPjUuOTU8L3ByaWNlPgogICAgICA8
cHVibGlzaF9kYXRlPjIwMDEtMDktMTA8L3B1Ymxpc2hfZGF0ZT4KICAgICAgPGRlc2NyaXB0aW9u
PlRoZSB0d28gZGF1Z2h0ZXJzIG9mIE1hZXZlLCBoYWxmLXNpc3RlcnMsIAogICAgICBiYXR0bGUg
b25lIGFub3RoZXIgZm9yIGNvbnRyb2wgb2YgRW5nbGFuZC4gU2VxdWVsIHRvIAogICAgICBPYmVy
b24ncyBMZWdhY3kuPC9kZXNjcmlwdGlvbj4KICAgPC9ib29rPgogICA8Ym9vayBpZD0iYmsxMDYi
PgogICAgICA8YXV0aG9yPlJhbmRhbGwsIEN5bnRoaWE8L2F1dGhvcj4KICAgICAgPHRpdGxlPkxv
dmVyIEJpcmRzPC90aXRsZT4KICAgICAgPGdlbnJlPlJvbWFuY2U8L2dlbnJlPgogICAgICA8cHJp
Y2U+NC45NTwvcHJpY2U+CiAgICAgIDxwdWJsaXNoX2RhdGU+MjAwMC0wOS0wMjwvcHVibGlzaF9k
YXRlPgogICAgICA8ZGVzY3JpcHRpb24+V2hlbiBDYXJsYSBtZWV0cyBQYXVsIGF0IGFuIG9ybml0
aG9sb2d5IAogICAgICBjb25mZXJlbmNlLCB0ZW1wZXJzIGZseSBhcyBmZWF0aGVycyBnZXQgcnVm
ZmxlZC48L2Rlc2NyaXB0aW9uPgogICA8L2Jvb2s+CiAgIDxib29rIGlkPSJiazEwNyI+CiAgICAg
IDxhdXRob3I+VGh1cm1hbiwgUGF1bGE8L2F1dGhvcj4KICAgICAgPHRpdGxlPlNwbGlzaCBTcGxh
c2g8L3RpdGxlPgogICAgICA8Z2VucmU+Um9tYW5jZTwvZ2VucmU+CiAgICAgIDxwcmljZT40Ljk1
PC9wcmljZT4KICAgICAgPHB1Ymxpc2hfZGF0ZT4yMDAwLTExLTAyPC9wdWJsaXNoX2RhdGU+CiAg
ICAgIDxkZXNjcmlwdGlvbj5BIGRlZXAgc2VhIGRpdmVyIGZpbmRzIHRydWUgbG92ZSB0d2VudHkg
CiAgICAgIHRob3VzYW5kIGxlYWd1ZXMgYmVuZWF0aCB0aGUgc2VhLjwvZGVzY3JpcHRpb24+CiAg
IDwvYm9vaz4KICAgPGJvb2sgaWQ9ImJrMTA4Ij4KICAgICAgPGF1dGhvcj5Lbm9yciwgU3RlZmFu
PC9hdXRob3I+CiAgICAgIDx0aXRsZT5DcmVlcHkgQ3Jhd2xpZXM8L3RpdGxlPgogICAgICA8Z2Vu
cmU+SG9ycm9yPC9nZW5yZT4KICAgICAgPHByaWNlPjQuOTU8L3ByaWNlPgogICAgICA8cHVibGlz
aF9kYXRlPjIwMDAtMTItMDY8L3B1Ymxpc2hfZGF0ZT4KICAgICAgPGRlc2NyaXB0aW9uPkFuIGFu
dGhvbG9neSBvZiBob3Jyb3Igc3RvcmllcyBhYm91dCByb2FjaGVzLAogICAgICBjZW50aXBlZGVz
LCBzY29ycGlvbnMgIGFuZCBvdGhlciBpbnNlY3RzLjwvZGVzY3JpcHRpb24+CiAgIDwvYm9vaz4K
ICAgPGJvb2sgaWQ9ImJrMTA5Ij4KICAgICAgPGF1dGhvcj5LcmVzcywgUGV0ZXI8L2F1dGhvcj4K
ICAgICAgPHRpdGxlPlBhcmFkb3ggTG9zdDwvdGl0bGU+CiAgICAgIDxnZW5yZT5TY2llbmNlIEZp
Y3Rpb248L2dlbnJlPgogICAgICA8cHJpY2U+Ni45NTwvcHJpY2U+CiAgICAgIDxwdWJsaXNoX2Rh
dGU+MjAwMC0xMS0wMjwvcHVibGlzaF9kYXRlPgogICAgICA8ZGVzY3JpcHRpb24+QWZ0ZXIgYW4g
aW5hZHZlcnRhbnQgdHJpcCB0aHJvdWdoIGEgSGVpc2VuYmVyZwogICAgICBVbmNlcnRhaW50eSBE
ZXZpY2UsIEphbWVzIFNhbHdheSBkaXNjb3ZlcnMgdGhlIHByb2JsZW1zIAogICAgICBvZiBiZWlu
ZyBxdWFudHVtLjwvZGVzY3JpcHRpb24+CiAgIDwvYm9vaz4KICAgPGJvb2sgaWQ9ImJrMTEwIj4K
ICAgICAgPGF1dGhvcj5PJ0JyaWVuLCBUaW08L2F1dGhvcj4KICAgICAgPHRpdGxlPk1pY3Jvc29m
dCAuTkVUOiBUaGUgUHJvZ3JhbW1pbmcgQmlibGU8L3RpdGxlPgogICAgICA8Z2VucmU+Q29tcHV0
ZXI8L2dlbnJlPgogICAgICA8cHJpY2U+MzYuOTU8L3ByaWNlPgogICAgICA8cHVibGlzaF9kYXRl
PjIwMDAtMTItMDk8L3B1Ymxpc2hfZGF0ZT4KICAgICAgPGRlc2NyaXB0aW9uPk1pY3Jvc29mdCdz
IC5ORVQgaW5pdGlhdGl2ZSBpcyBleHBsb3JlZCBpbiAKICAgICAgZGV0YWlsIGluIHRoaXMgZGVl
cCBwcm9ncmFtbWVyJ3MgcmVmZXJlbmNlLjwvZGVzY3JpcHRpb24+CiAgIDwvYm9vaz4KICAgPGJv
b2sgaWQ9ImJrMTExIj4KICAgICAgPGF1dGhvcj5PJ0JyaWVuLCBUaW08L2F1dGhvcj4KICAgICAg
PHRpdGxlPk1TWE1MMzogQSBDb21wcmVoZW5zaXZlIEd1aWRlPC90aXRsZT4KICAgICAgPGdlbnJl
PkNvbXB1dGVyPC9nZW5yZT4KICAgICAgPHByaWNlPjM2Ljk1PC9wcmljZT4KICAgICAgPHB1Ymxp
c2hfZGF0ZT4yMDAwLTEyLTAxPC9wdWJsaXNoX2RhdGU+CiAgICAgIDxkZXNjcmlwdGlvbj5UaGUg
TWljcm9zb2Z0IE1TWE1MMyBwYXJzZXIgaXMgY292ZXJlZCBpbiAKICAgICAgZGV0YWlsLCB3aXRo
IGF0dGVudGlvbiB0byBYTUwgRE9NIGludGVyZmFjZXMsIFhTTFQgcHJvY2Vzc2luZywgCiAgICAg
IFNBWCBhbmQgbW9yZS48L2Rlc2NyaXB0aW9uPgogICA8L2Jvb2s+CiAgIDxib29rIGlkPSJiazEx
MiI+CiAgICAgIDxhdXRob3I+R2Fsb3MsIE1pa2U8L2F1dGhvcj4KICAgICAgPHRpdGxlPlZpc3Vh
bCBTdHVkaW8gNzogQSBDb21wcmVoZW5zaXZlIEd1aWRlPC90aXRsZT4KICAgICAgPGdlbnJlPkNv
bXB1dGVyPC9nZW5yZT4KICAgICAgPHByaWNlPjQ5Ljk1PC9wcmljZT4KICAgICAgPHB1Ymxpc2hf
ZGF0ZT4yMDAxLTA0LTE2PC9wdWJsaXNoX2RhdGU+CiAgICAgIDxkZXNjcmlwdGlvbj5NaWNyb3Nv
ZnQgVmlzdWFsIFN0dWRpbyA3IGlzIGV4cGxvcmVkIGluIGRlcHRoLAogICAgICBsb29raW5nIGF0
IGhvdyBWaXN1YWwgQmFzaWMsIFZpc3VhbCBDKyssIEMjLCBhbmQgQVNQKyBhcmUgCiAgICAgIGlu
dGVncmF0ZWQgaW50byBhIGNvbXByZWhlbnNpdmUgZGV2ZWxvcG1lbnQgCiAgICAgIGVudmlyb25t
ZW50LjwvZGVzY3JpcHRpb24+CiAgIDwvYm9vaz4KPC9jYXRhbG9nPg==</XML>
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
	// Decode the embedded XML
	decodedXML, err := util.DecodeEmbeddedXML(setPayload.Body.Documents[0].EmbeddedXML)
	require.NoError(t, err, "failed to decode embedded XML")

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
	fileName, err := controller.processAndPersist(context.Background(), decodedXML, originalDoc)
	require.NoError(t, err)
	require.Equal(t, "testFileName", fileName)

	// Verify that the captured XML starts with the XML declaration.
	expectedHeader := regexp.MustCompile(`^<\?xml\s+version="1\.0"(?:\s+encoding="UTF-8")?.*\?>\n?`)
	if !expectedHeader.Match(capturedXML) {
		t.Errorf("expected XML header to match regex %q, got: %s", expectedHeader.String(), string(capturedXML))
	}
}
