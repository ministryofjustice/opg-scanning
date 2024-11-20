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
	Config config.Config
	Logger logger.Logger
}

func NewHttpClient(config config.Config, logger logger.Logger) *HttpClient {
	return &HttpClient{
		Config: config,
		Logger: logger,
	}
}

func (r *HttpClient) HTTPRequest(url string, method string, payload []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Duration(r.Config.HTTP.Timeout),
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d - Response: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log request details
	r.Logger.Info(fmt.Sprintf("new request to Sirius API: %s - Payload:\n%s", url, payload))

	return body, nil
}
