package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type SiriusClientError struct {
	StatusCode       int
	ValidationErrors map[string]map[string]string
}

func (sce SiriusClientError) Error() string {
	return fmt.Sprintf("received %d response from Sirius", sce.StatusCode)
}

type HttpClient struct {
	HttpClient *http.Client
	Config     *config.Config
	Logger     *logger.Logger
}

func NewHttpClient(config config.Config, logger logger.Logger) *HttpClient {
	httpClient := &HttpClient{
		HttpClient: &http.Client{
			Timeout: time.Duration(config.HTTP.Timeout) * time.Second,
		},
		Config: &config,
		Logger: &logger,
	}

	httpClient.HttpClient.Transport = otelhttp.NewTransport(httpClient.HttpClient.Transport)

	return httpClient
}

func (r *HttpClient) HTTPRequest(ctx context.Context, url, method string, payload []byte, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Use the shared HTTP client
	resp, err := r.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-2xx responses
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		errOut := SiriusClientError{
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

		return nil, errOut
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d - Response: %s", resp.StatusCode, string(body))
	}

	// Handle successful responses
	return body, nil
}
