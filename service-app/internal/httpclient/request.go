package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

type HttpClient struct {
	HttpClient *http.Client
	Config     config.Config
	Logger     logger.Logger
}

func NewHttpClient(config config.Config, logger logger.Logger) *HttpClient {
	return &HttpClient{
		HttpClient: &http.Client{
			Timeout: time.Duration(config.HTTP.Timeout) * time.Second,
		},
		Config: config,
		Logger: logger,
	}
}

func (r *HttpClient) HTTPRequest(url, method string, payload []byte, headers map[string]string) ([]byte, error) {
	// Log the request details before sending it
	r.Logger.Info(fmt.Sprintf("Sending request to URL: %s", url))

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
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
	defer resp.Body.Close()

	// Handle non-2xx responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("unexpected status code: %d - failed to read error response body: %w", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("unexpected status code: %d - Response: %s", resp.StatusCode, string(body))
	}

	// Handle successful responses
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}
