package sirius

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = context.WithValue(context.Background(), constants.UserContextKey, "testing-key")

func TestNewClient(t *testing.T) {
	config := &config.Config{}
	config.App.SiriusBaseURL = "http://example.com"
	config.App.SiriusAttachDocURL = "attach"
	config.App.SiriusCaseStubURL = "case"

	client := NewClient(config)

	assert.Equal(t, "http://example.com/attach", client.attachDocumentURL)
	assert.Equal(t, "http://example.com/case", client.caseStubURL)
}

func TestClientAttachDocument(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			body, _ := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewReader(body))

			return assert.Equal(t, http.MethodPost, req.Method) &&
				assert.Equal(t, "http://example.com/attach", req.URL.String()) &&
				assert.Equal(t, "application/json", req.Header.Get("Content-Type")) &&
				assert.Equal(t, "Bearer testing-key", req.Header.Get("Authorization")) &&
				assert.JSONEq(t, `{"caseReference":"5","content":"","documentType":"","scannedDate":""}`, string(body))
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"uuid":"hello"}`)),
		}, nil)

	client := &Client{
		httpClient:        doer,
		attachDocumentURL: "http://example.com/attach",
	}

	resp, err := client.AttachDocument(ctx, &ScannedDocumentRequest{CaseReference: "5"})
	assert.Nil(t, err)
	assert.Equal(t, &ScannedDocumentResponse{UUID: "hello"}, resp)
}

func TestClientCreateCaseStub(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			body, _ := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewReader(body))

			return assert.Equal(t, http.MethodPost, req.Method) &&
				assert.Equal(t, "http://example.com/case", req.URL.String()) &&
				assert.Equal(t, "application/json", req.Header.Get("Content-Type")) &&
				assert.Equal(t, "Bearer testing-key", req.Header.Get("Authorization")) &&
				assert.JSONEq(t, `{"batchId":"1","caseType":"","receiptDate":"","createdDate":""}`, string(body))
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"uid":"hello"}`)),
		}, nil)

	client := &Client{
		httpClient:  doer,
		caseStubURL: "http://example.com/case",
	}

	resp, err := client.CreateCaseStub(ctx, &ScannedCaseRequest{BatchID: "1"})
	assert.Nil(t, err)
	assert.Equal(t, &ScannedCaseResponse{UID: "hello"}, resp)
}

func TestNewRequest(t *testing.T) {
	req, err := newRequest(ctx, "url", "data")

	assert.Nil(t, err)
	assert.Equal(t, ctx, req.Context())
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, "url", req.URL.String())
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer testing-key", req.Header.Get("Authorization"))

	body, _ := io.ReadAll(req.Body)
	assert.Equal(t, `"data"`, string(body))
}

func TestNewRequest_NilData(t *testing.T) {
	_, err := newRequest(ctx, "url", nil)
	assert.Error(t, err)
}

func TestNewRequest_MissingUserToken(t *testing.T) {
	_, err := newRequest(context.Background(), "url", "data")
	assert.Error(t, err)
}

func TestClientDo(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "url", nil)
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`"blah"`)),
	}

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(req).
		Return(resp, nil)

	client := &Client{httpClient: doer}

	var v string
	err := client.do(req, &v)

	assert.Nil(t, err)
	assert.Equal(t, "blah", v)
}

func TestClientDo_404(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "url", nil)
	resp := &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(strings.NewReader(`"blah"`)),
	}

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(req).
		Return(resp, nil)

	client := &Client{httpClient: doer}
	err := client.do(req, nil)

	assert.Equal(t, Error{StatusCode: 404}, err)
}

func TestClientDo_400(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "url", nil)
	resp := &http.Response{
		StatusCode: 400,
		Body:       io.NopCloser(strings.NewReader(`{"validation_errors":{"a":{"b":"c"}}}`)),
	}

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(req).
		Return(resp, nil)

	client := &Client{httpClient: doer}
	err := client.do(req, nil)

	assert.Equal(t, Error{StatusCode: 400, ValidationErrors: map[string]map[string]string{"a": {"b": "c"}}}, err)
}

func TestClientDo_500(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "url", nil)
	resp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader(`"blah"`)),
	}

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(req).
		Return(resp, nil)

	client := &Client{httpClient: doer}
	err := client.do(req, nil)

	assert.ErrorContains(t, err, "unexpected status code")
}
