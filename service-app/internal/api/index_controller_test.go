package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/ingestion"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/mocks"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var xmlPayload = `
<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
    <Header CaseNo="" Scanner="9" ScanTime="2014-09-26T12:38:53" ScannerOperator="Administrator" Schedule="02-0001112-20160909185000" FeeNumber="1234"/>
    <Body>
        <Document Type="LP2" Encoding="UTF-8" NoPages="19">
            <XML>PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz48TFAyIHhtbG5zOnZjPSJodHRwOi8vd3d3LnczLm9yZy8yMDA3L1hNTFNjaGVtYS12ZXJzaW9uaW5nIiB4c2k6bm9OYW1lc3BhY2VTY2hlbWFMb2NhdGlvbj0iTFAyLnhzZCIgeG1sbnM6eHNpPSJodHRwOi8vd3d3LnczLm9yZy8yMDAxL1hNTFNjaGVtYS1pbnN0YW5jZSI+PFBhZ2UxPjxTZWN0aW9uMT48VGl0bGU+UHJvZjwvVGl0bGU+PEZpcnN0TmFtZT5GbGF2aW88L0ZpcnN0TmFtZT48TGFzdE5hbWU+TWlsbGVyPC9MYXN0TmFtZT48UHJvcGVydHlGaW5hbmNpYWxBZmZhaXJzPjA8L1Byb3BlcnR5RmluYW5jaWFsQWZmYWlycz48SGVhbHRoV2VsZmFyZT4xPC9IZWFsdGhXZWxmYXJlPjwvU2VjdGlvbjE+PEJVUk4+c3RyaW5nPC9CVVJOPjxQaHlzaWNhbFBhZ2U+MTwvUGh5c2ljYWxQYWdlPjwvUGFnZTE+PFBhZ2UyPjxTZWN0aW9uMj48RG9ub3JSZWdpc3RlcmF0aW9uPnRydWU8L0Rvbm9yUmVnaXN0ZXJhdGlvbj48QXR0b3JuZXlSZWdpc3RlcmF0aW9uPmZhbHNlPC9BdHRvcm5leVJlZ2lzdGVyYXRpb24+PEF0dG9ybmV5PjxUaXRsZT5NcnM8L1RpdGxlPjxGaXJzdE5hbWU+TWFyZ3JldDwvRmlyc3ROYW1lPjxMYXN0TmFtZT5KZW5raW5zLUJhcnJvd3M8L0xhc3ROYW1lPjxET0I+MjQwNjE5NjM8L0RPQj48L0F0dG9ybmV5PjxBdHRvcm5leT48VGl0bGU+TXI8L1RpdGxlPjxGaXJzdE5hbWU+SnVsaXVzIEphZGVuPC9GaXJzdE5hbWU+PExhc3ROYW1lPkhlaWRlbnJlaWNoPC9MYXN0TmFtZT48RE9CPjE4MDIxOTk2PC9ET0I+PC9BdHRvcm5leT48QXR0b3JuZXk+PFRpdGxlPjwvVGl0bGU+PEZpcnN0TmFtZT48L0ZpcnN0TmFtZT48TGFzdE5hbWU+PC9MYXN0TmFtZT48RE9CPjwvRE9CPjwvQXR0b3JuZXk+PEF0dG9ybmV5PjxUaXRsZT48L1RpdGxlPjxGaXJzdE5hbWU+PC9GaXJzdE5hbWU+PExhc3ROYW1lPjwvTGFzdE5hbWU+PERPQj48L0RPQj48L0F0dG9ybmV5PjwvU2VjdGlvbjI+PEJVUk4+c3RyaW5nPC9CVVJOPjxQaHlzaWNhbFBhZ2U+MjwvUGh5c2ljYWxQYWdlPjwvUGFnZTI+PFBhZ2UzPjxTZWN0aW9uMz48VGhlRG9ub3I+dHJ1ZTwvVGhlRG9ub3I+PEFuQXR0b3JuZXk+ZmFsc2U8L0FuQXR0b3JuZXk+PE90aGVyPmZhbHNlPC9PdGhlcj48VGl0bGU+PC9UaXRsZT48Rmlyc3ROYW1lPjwvRmlyc3ROYW1lPjxMYXN0TmFtZT48L0xhc3ROYW1lPjxDb21wYW55PjwvQ29tcGFueT48QWRkcmVzcz48QWRkcmVzczE+PC9BZGRyZXNzMT48QWRkcmVzczI+PC9BZGRyZXNzMj48QWRkcmVzczM+PC9BZGRyZXNzMz48UG9zdGNvZGU+PC9Qb3N0Y29kZT48L0FkZHJlc3M+PFBvc3Q+dHJ1ZTwvUG9zdD48UGhvbmU+ZmFsc2U8L1Bob25lPjxQaG9uZU51bWJlcj48L1Bob25lTnVtYmVyPjxFbWFpbD5mYWxzZTwvRW1haWw+PEVtYWlsQWRkcmVzcz48L0VtYWlsQWRkcmVzcz48V2Vsc2g+ZmFsc2U8L1dlbHNoPjwvU2VjdGlvbjM+PEJVUk4+c3RyaW5nPC9CVVJOPjxQaHlzaWNhbFBhZ2U+MzwvUGh5c2ljYWxQYWdlPjwvUGFnZTM+PFBhZ2U0PjxTZWN0aW9uND48Q2hlcXVlPmZhbHNlPC9DaGVxdWU+PENhcmQ+dHJ1ZTwvQ2FyZD48UGhvbmVOdW1iZXI+MDE4MzEwIDMxMjk1PC9QaG9uZU51bWJlcj48UmVkdWNlZEZlZT5mYWxzZTwvUmVkdWNlZEZlZT48L1NlY3Rpb240PjxCVVJOPnN0cmluZzwvQlVSTj48UGh5c2ljYWxQYWdlPjQ8L1BoeXNpY2FsUGFnZT48L1BhZ2U0PjxQYWdlNT48U2VjdGlvbjU+PEF0dG9ybmV5PjxTaWduYXR1cmU+dHJ1ZTwvU2lnbmF0dXJlPjxEYXRlPjE4MDIyMDI1PC9EYXRlPjwvQXR0b3JuZXk+PEF0dG9ybmV5PjxTaWduYXR1cmU+dHJ1ZTwvU2lnbmF0dXJlPjxEYXRlPjE4MDIyMDI1PC9EYXRlPjwvQXR0b3JuZXk+PEF0dG9ybmV5PjxTaWduYXR1cmU+ZmFsc2U8L1NpZ25hdHVyZT48RGF0ZT48L0RhdGU+PC9BdHRvcm5leT48QXR0b3JuZXk+PFNpZ25hdHVyZT5mYWxzZTwvU2lnbmF0dXJlPjxEYXRlPjwvRGF0ZT48L0F0dG9ybmV5PjwvU2VjdGlvbjU+PEJVUk4+c3RyaW5nPC9CVVJOPjxQaHlzaWNhbFBhZ2U+NTwvUGh5c2ljYWxQYWdlPjwvUGFnZTU+PFBhZ2U2PjxTZWN0aW9uNj48QWRkcmVzc2VzPjxUaXRsZT5NcnM8L1RpdGxlPjxGaXJzdE5hbWU+TWFyZ3JldDwvRmlyc3ROYW1lPjxMYXN0TmFtZT5KZW5raW5zLUJhcnJvd3M8L0xhc3ROYW1lPjxBZGRyZXNzPjxBZGRyZXNzMT43NyBKYXNrb2xza2kgRmllbGQ8L0FkZHJlc3MxPjxBZGRyZXNzMj5Lb25vcGVsc2tpLW9uLUNocmlzdGlhbnNlbjwvQWRkcmVzczI+PEFkZHJlc3MzPkNyb25hbGV5PC9BZGRyZXNzMz48UG9zdGNvZGU+S0IyNCA0S1k8L1Bvc3Rjb2RlPjwvQWRkcmVzcz48RW1haWxBZGRyZXNzPjwvRW1haWxBZGRyZXNzPjwvQWRkcmVzc2VzPjxBZGRyZXNzZXM+PFRpdGxlPk1yPC9UaXRsZT48Rmlyc3ROYW1lPkp1bGl1cyBKYWRlbjwvRmlyc3ROYW1lPjxMYXN0TmFtZT5IZWlkZW5yZWljaDwvTGFzdE5hbWU+PEFkZHJlc3M+PEFkZHJlc3MxPjMwNCBEYWtvdGEgQnJhZTwvQWRkcmVzczE+PEFkZHJlc3MyPlN0LiBEaWNraWhpbGw8L0FkZHJlc3MyPjxBZGRyZXNzMz48L0FkZHJlc3MzPjxQb3N0Y29kZT5JSzUgN1FUPC9Qb3N0Y29kZT48L0FkZHJlc3M+PEVtYWlsQWRkcmVzcz5qamhlaWRlbnJlaWNoQGJ1c2luZXNzLmV4YW1wbGU8L0VtYWlsQWRkcmVzcz48L0FkZHJlc3Nlcz48QWRkcmVzc2VzPjxUaXRsZT48L1RpdGxlPjxGaXJzdE5hbWU+PC9GaXJzdE5hbWU+PExhc3ROYW1lPjwvTGFzdE5hbWU+PEFkZHJlc3M+PEFkZHJlc3MxPjwvQWRkcmVzczE+PEFkZHJlc3MyPjwvQWRkcmVzczI+PEFkZHJlc3MzPjwvQWRkcmVzczM+PFBvc3Rjb2RlPjwvUG9zdGNvZGU+PC9BZGRyZXNzPjxFbWFpbEFkZHJlc3M+PC9FbWFpbEFkZHJlc3M+PC9BZGRyZXNzZXM+PEFkZHJlc3Nlcz48VGl0bGU+PC9UaXRsZT48Rmlyc3ROYW1lPjwvRmlyc3ROYW1lPjxMYXN0TmFtZT48L0xhc3ROYW1lPjxBZGRyZXNzPjxBZGRyZXNzMT48L0FkZHJlc3MxPjxBZGRyZXNzMj48L0FkZHJlc3MyPjxBZGRyZXNzMz48L0FkZHJlc3MzPjxQb3N0Y29kZT48L1Bvc3Rjb2RlPjwvQWRkcmVzcz48RW1haWxBZGRyZXNzPjwvRW1haWxBZGRyZXNzPjwvQWRkcmVzc2VzPjwvU2VjdGlvbjY+PEJVUk4+c3RyaW5nPC9CVVJOPjxQaHlzaWNhbFBhZ2U+NjwvUGh5c2ljYWxQYWdlPjwvUGFnZTY+PEluZm9QYWdlPjxCVVJOPnN0cmluZzwvQlVSTj48UGh5c2ljYWxQYWdlPjc8L1BoeXNpY2FsUGFnZT48L0luZm9QYWdlPjwvTFAyPg==</XML>
            <PDF>SGVsbG8gd29ybGQ=</PDF>
        </Document>
    </Body>
</Set>
`

func setupController() *IndexController {
	appConfig := config.NewConfig()
	logger := logger.GetLogger(appConfig)

	// Create mock dependencies
	mockHttpClient, mockAuthMiddleware, awsClient, _ := auth.PrepareMocks(appConfig, logger)
	httpMiddleware, _ := httpclient.NewMiddleware(mockHttpClient)

	mockHttpClient.On("HTTPRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]byte(`{"UID": "7000-1234-1234"}`), nil)

	controller := &IndexController{
		config:         appConfig,
		logger:         logger,
		validator:      ingestion.NewValidator(),
		httpMiddleware: httpMiddleware,
		authMiddleware: mockAuthMiddleware,
		Queue:          ingestion.NewJobQueue(appConfig),
		AwsClient:      awsClient,
	}

	return controller
}

func TestIngestHandler_SetValid(t *testing.T) {
	controller := setupController()

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer([]byte(xmlPayload)))
	req.Header.Set("Content-Type", "application/xml")
	w := httptest.NewRecorder()

	reqCtx := context.WithValue(context.Background(), constants.UserContextKey, "my-token")
	req = req.WithContext(reqCtx)

	controller.IngestHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected status %d; got %d", http.StatusAccepted, resp.StatusCode)
	}

	responseBody, _ := io.ReadAll(resp.Body)
	var responseObj response

	err := json.Unmarshal(responseBody, &responseObj)
	assert.Nil(t, err)
	assert.True(t, responseObj.Data.Success)
	assert.Equal(t, "700012341234", responseObj.Data.Uid)
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

func TestIngestHandler_InvalidXMLExplainsXSDErrors(t *testing.T) {
	controller := setupController()

	xmlPayloadMalformed := `<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
		<Header CaseNo="1234"></Header>
	</Set>`

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer([]byte(xmlPayloadMalformed)))
	req.Header.Set("Content-Type", "application/xml")
	w := httptest.NewRecorder()

	controller.IngestHandler(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	responseBody, _ := io.ReadAll(resp.Body)
	var responseObj response

	err := json.Unmarshal(responseBody, &responseObj)
	assert.Nil(t, err)
	assert.False(t, responseObj.Data.Success)
	assert.Contains(t, responseObj.Data.ValidationErrors, "Element 'Set': Missing child element(s). Expected is ( Body ).")
}

func TestIngestHandler_InvalidEmbeddedXMLProvidesDetails(t *testing.T) {
	controller := setupController()

	xmlPayloadMalformed := `<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
		<Header CaseNo="" Scanner="9" ScanTime="2014-09-26T12:38:53" ScannerOperator="Administrator" Schedule="02-0001112-20160909185000" />
		<Body>
			<Document Type="LP1F" Encoding="UTF-8" NoPages="19">
				<XML>PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiIHN0YW5kYWxvbmU9Im5vIj8+CjxMUDIgeG1sbnM6eHNpPSJodHRwOi8vd3d3LnczLm9yZy8yMDAxL1hNTFNjaGVtYS1pbnN0YW5jZSIgeHNpOm5vTmFtZXNwYWNlU2NoZW1hTG9jYXRpb249IkxQMi54c2QiPjwvTFAyPg==</XML>
				<PDF>SGVsbG8gd29ybGQ=</PDF>
			</Document>
		</Body>
	</Set>`

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer([]byte(xmlPayloadMalformed)))
	req.Header.Set("Content-Type", "application/xml")
	w := httptest.NewRecorder()

	controller.IngestHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d; got %d", http.StatusBadRequest, resp.StatusCode)
	}

	responseBody, _ := io.ReadAll(resp.Body)
	var responseObj response

	err := json.Unmarshal(responseBody, &responseObj)
	assert.Nil(t, err)
	assert.False(t, responseObj.Data.Success)
	assert.Contains(t, responseObj.Data.ValidationErrors, "Element 'LP2': Missing child element(s). Expected is ( Page1 ).")
}

func TestIngestHandler_SiriusErrors(t *testing.T) {
	xmlPayloadCorrespondence := `<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
		<Header CaseNo="700012341234" Scanner="9" ScanTime="2014-09-26T12:38:53" ScannerOperator="Administrator" Schedule="02-0001112-20160909185000" />
		<Body>
			<Document Type="Correspondence" Encoding="UTF-8" NoPages="19">
				<XML>PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPENvcnJlc3BvbmRlbmNlIHhtbG5zOnhzaT0iaHR0cDovL3d3dy53My5vcmcvMjAwMS9YTUxTY2hlbWEtaW5zdGFuY2UiCiAgeHNpOm5vTmFtZXNwYWNlU2NoZW1hTG9jYXRpb249IkNvcnJlc3BvbmRlbmNlLnhzZCI+CiAgPFN1YlR5cGU+TGVnYWw8L1N1YlR5cGU+CiAgPENhc2VOdW1iZXI+MTIzNDU8L0Nhc2VOdW1iZXI+CiAgPENhc2VOdW1iZXI+Njc4OTA8L0Nhc2VOdW1iZXI+CiAgPFBhZ2U+CiAgICA8QlVSTj4xMjNBQkM8L0JVUk4+CiAgICA8UGh5c2ljYWxQYWdlPjE8L1BoeXNpY2FsUGFnZT4KICA8L1BhZ2U+CiAgPFBhZ2U+CiAgICA8QlVSTj40NTZERUY8L0JVUk4+CiAgICA8UGh5c2ljYWxQYWdlPjI8L1BoeXNpY2FsUGFnZT4KICA8L1BhZ2U+CjwvQ29ycmVzcG9uZGVuY2U+Cg==</XML>
				<PDF>SGVsbG8gd29ybGQ=</PDF>
			</Document>
		</Body>
	</Set>`

	testCases := map[string]struct {
		siriusError        httpclient.SiriusClientError
		expectedStatusCode int
		expectedMessage    string
	}{
		"404": {
			siriusError: httpclient.SiriusClientError{
				StatusCode: 404,
			},
			expectedStatusCode: 400,
			expectedMessage:    "Case not found with UID 700012341234",
		},
		"400 on case reference": {
			siriusError: httpclient.SiriusClientError{
				StatusCode: 400,
				ValidationErrors: map[string]map[string]string{
					"caseReference": {
						"regexNotMatch": "The input does not match against pattern",
					},
				},
			},
			expectedStatusCode: 400,
			expectedMessage:    "700012341234 is not a valid case UID",
		},
		"other 400": {
			siriusError: httpclient.SiriusClientError{
				StatusCode: 400,
			},
			expectedStatusCode: 500,
			expectedMessage:    "Failed to persist document to Sirius",
		},
		"500": {
			siriusError: httpclient.SiriusClientError{
				StatusCode: 500,
			},
			expectedStatusCode: 500,
			expectedMessage:    "Failed to persist document to Sirius",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			controller := setupController()

			mockHttpClient := new(mocks.MockHttpClient)
			mockHttpClient.On("GetConfig").Return(controller.config)
			mockHttpClient.On("GetLogger").Return(controller.logger)
			mockHttpClient.On("HTTPRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return([]byte{}, tc.siriusError)

			httpMiddleware, _ := httpclient.NewMiddleware(mockHttpClient)
			controller.httpMiddleware = httpMiddleware

			req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer([]byte(xmlPayloadCorrespondence)))
			req.Header.Set("Content-Type", "application/xml")
			w := httptest.NewRecorder()

			reqCtx := context.WithValue(context.Background(), constants.UserContextKey, "my-token")
			req = req.WithContext(reqCtx)

			controller.IngestHandler(w, req)

			resp := w.Result()
			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			responseBody, _ := io.ReadAll(resp.Body)
			var responseObj response

			err := json.Unmarshal(responseBody, &responseObj)
			assert.Nil(t, err)
			assert.False(t, responseObj.Data.Success)
			assert.Equal(t, tc.expectedMessage, responseObj.Data.Message)
		})
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
		On("PersistFormData", mock.Anything, mock.AnythingOfType("*bytes.Reader"), mock.Anything).
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

func TestValidateDocumentWarnsOnUnsupportedDocumentType(t *testing.T) {

	c := setupController()

	document := types.BaseDocument{
		Type:        "BadDocumentType",
		EmbeddedXML: "",
	}

	err := c.validateDocument(document)

	problem, ok := err.(Problem)
	assert.True(t, ok)

	assert.Equal(t, "Document type BadDocumentType is not supported", problem.Title)
}

func TestValidateDocumentHandlesErrorCases(t *testing.T) {
	testCases := []struct {
		name string
		XML  string
		err  string
	}{
		{
			name: "not XML",
			XML:  "not XML",
			err:  "failed to extract schema from EPA",
		},
		{
			name: "no schema",
			XML:  "<my-doc></my-doc>",
			err:  "failed to extract schema from EPA",
		},
		{
			name: "invalid schema",
			XML:  `<my-doc xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="MY-DOC.xsd"></my-doc>`,
			err:  "failed to load schema MY-DOC.xsd",
		},
		{
			name: "does not match schema",
			XML:  `<my-doc xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="LP2.xsd"></my-doc>`,
			err:  "XML for EPA failed XSD validation",
		},
		{
			name: "ok",
			XML:  `<EPA xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="EPA.xsd"><Page><BURN/><PhysicalPage>1</PhysicalPage></Page></EPA>`,
			err:  "",
		},
	}

	c := setupController()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encodedXML := base64.StdEncoding.EncodeToString([]byte(tc.XML))

			document := types.BaseDocument{
				Type:        "EPA",
				EmbeddedXML: encodedXML,
			}

			err := c.validateDocument(document)

			if tc.err == "" {
				assert.Nil(t, err)
			} else {
				assert.Contains(t, err.Error(), tc.err)
			}
		})
	}
}

func TestRespondWithErrorHandle5XX(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()

	c := setupController()

	outBuf := bytes.NewBuffer([]byte{})
	c.logger.SlogLogger = slog.New(slog.NewJSONHandler(outBuf, nil))

	c.respondWithError(ctx, w, 500, "something went wrong", errors.New("what really went wrong"))

	var logMessage map[string]string
	jsonUnmarshalReader(outBuf, &logMessage)
	assert.Equal(t, "ERROR", logMessage["level"])
	assert.Equal(t, "something went wrong", logMessage["msg"])
	assert.Equal(t, "what really went wrong", logMessage["error"])

	resp := w.Result()
	assert.Equal(t, 500, resp.StatusCode)

	respBody := response{}
	jsonUnmarshalReader(resp.Body, &respBody)

	assert.Equal(t, false, respBody.Data.Success)
	assert.Equal(t, "something went wrong", respBody.Data.Message)
}

func TestRespondWithErrorHandle4XX(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()

	c := setupController()

	outBuf := bytes.NewBuffer([]byte{})
	c.logger.SlogLogger = slog.New(slog.NewJSONHandler(outBuf, nil))

	c.respondWithError(ctx, w, 400, "you sent us something wrong", errors.New("what really went wrong"))

	var logMessage map[string]string
	jsonUnmarshalReader(outBuf, &logMessage)
	assert.Equal(t, "INFO", logMessage["level"])
	assert.Equal(t, "you sent us something wrong", logMessage["msg"])
	assert.Equal(t, "what really went wrong", logMessage["error"])

	resp := w.Result()
	assert.Equal(t, 400, resp.StatusCode)

	respBody := response{}
	jsonUnmarshalReader(resp.Body, &respBody)

	assert.Equal(t, false, respBody.Data.Success)
	assert.Equal(t, "you sent us something wrong", respBody.Data.Message)
}

func jsonUnmarshalReader(reader io.Reader, v any) {
	body, err := io.ReadAll(reader)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, v)

	if err != nil {
		panic(err)
	}
}
