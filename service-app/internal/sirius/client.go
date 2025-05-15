package sirius

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	httpClient        doer
	attachDocumentURL string
	caseStubURL       string
}

func NewClient(config *config.Config) *Client {
	httpClient := &http.Client{
		Timeout: time.Duration(config.HTTP.Timeout) * time.Second,
	}
	httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)

	return &Client{
		httpClient:        httpClient,
		attachDocumentURL: fmt.Sprintf("%s/%s", config.App.SiriusBaseURL, config.App.SiriusAttachDocURL),
		caseStubURL:       fmt.Sprintf("%s/%s", config.App.SiriusBaseURL, config.App.SiriusCaseStubURL),
	}
}

type ScannedDocumentRequest struct {
	CaseReference   string `json:"caseReference"`
	Content         string `json:"content"`
	DocumentType    string `json:"documentType"`
	DocumentSubType string `json:"documentSubType,omitempty"`
	ScannedDate     string `json:"scannedDate"`
}

type ScannedDocumentResponse struct {
	UUID string `json:"uuid"`
}

func (c *Client) AttachDocument(ctx context.Context, data *ScannedDocumentRequest) (*ScannedDocumentResponse, error) {
	req, err := newRequest(ctx, c.attachDocumentURL, data)
	if err != nil {
		return nil, err
	}

	v := &ScannedDocumentResponse{}
	if err := c.do(req, v); err != nil {
		return nil, fmt.Errorf("client attach document: %w", err)
	}

	return v, nil
}

type ScannedCaseRequest struct {
	BatchID        string `json:"batchId"`
	CaseType       string `json:"caseType"`
	CourtReference string `json:"courtReference,omitempty"`
	ReceiptDate    string `json:"receiptDate"`
	CreatedDate    string `json:"createdDate"`
}

type ScannedCaseResponse struct {
	UID string `json:"uId"`
}

func (c *Client) CreateCaseStub(ctx context.Context, data *ScannedCaseRequest) (*ScannedCaseResponse, error) {
	req, err := newRequest(ctx, c.caseStubURL, data)
	if err != nil {
		return nil, err
	}

	v := &ScannedCaseResponse{}
	if err := c.do(req, v); err != nil {
		return nil, fmt.Errorf("client create case stub: %w", err)
	}

	return v, nil
}

func newRequest(ctx context.Context, url string, data any) (*http.Request, error) {
	if data == nil {
		return nil, fmt.Errorf("data is nil")
	}

	token, ok := ctx.Value(constants.UserContextKey).(string)
	if !ok {
		return nil, fmt.Errorf("could not fetch user token from context")
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	return req, nil
}

func (c *Client) do(req *http.Request, v any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	// Handle non-2xx responses
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		errOut := Error{
			StatusCode: resp.StatusCode,
		}

		if resp.StatusCode == 400 {
			var respBody struct {
				ValidationErrors map[string]map[string]string `json:"validation_errors"`
			}

			if json.Unmarshal(body, &respBody) == nil {
				errOut.ValidationErrors = respBody.ValidationErrors
			}
		}

		return errOut
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d - Response: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, &v); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	return nil
}
