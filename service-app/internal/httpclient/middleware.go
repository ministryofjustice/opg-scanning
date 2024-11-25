package httpclient

import (
	"fmt"
)

type Middleware struct {
	Client *HttpClient
}

func NewMiddleware(client *HttpClient) *Middleware {
	return &Middleware{
		Client: client,
	}
}

func (m *Middleware) FetchToken() error {
	m.Client.Logger.Info("Skipping token fetching; JWT implementation pending.")
	return nil
}

func (m *Middleware) EnsureToken() error {
	m.Client.Logger.Info("Skipping token validation; JWT implementation pending.")
	return nil
}

func (m *Middleware) HTTPRequest(url, method string, payload []byte, headers map[string]string) ([]byte, error) {

	// Copy headers for the request
	headersCopy := make(map[string]string, len(headers))
	for k, v := range headers {
		headersCopy[k] = v
	}

	// Perform the HTTP request
	response, err := m.Client.HTTPRequest(url, method, payload, headersCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	return response, nil
}
