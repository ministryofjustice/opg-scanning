package httpclient

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type Middleware struct {
	Client           *HttpClient
	TokenMutex       sync.Mutex
	Cookies          string
	RefreshThreshold time.Duration
}

func NewMiddleware(client *HttpClient, refreshThreshold time.Duration) *Middleware {
	return &Middleware{
		Client:           client,
		RefreshThreshold: refreshThreshold,
	}
}

// Fetches the authentication token and builds the cookie
func (m *Middleware) FetchToken() error {
	m.Client.Logger.Info("Fetching authentication token...")

	authURL := fmt.Sprintf("%s/auth/sessions", m.Client.Config.App.SiriusBaseURL)
	authRequest := types.AuthRequest{
		User: types.User{
			Email:    m.Client.Config.Auth.Email,
			Password: m.Client.Config.Auth.Password,
		},
	}

	payloadBytes, err := json.Marshal(authRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	response, err := m.Client.HTTPRequest(authURL, "POST", payloadBytes, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch authentication token: %w", err)
	}

	// Parse the authentication response
	var authResponse types.AuthResponse
	if err := json.Unmarshal(response, &authResponse); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	if authResponse.AuthenticationToken == "" {
		return fmt.Errorf("authentication failed: missing token in response")
	}

	// Build the cookie manually
	m.TokenMutex.Lock()
	m.Cookies = fmt.Sprintf("XSRF-TOKEN=%s; membrane=%s", authResponse.AuthenticationToken, authResponse.AuthenticationToken)
	m.TokenMutex.Unlock()

	m.Client.Logger.Info("Successfully fetched and built authentication cookies.")
	return nil
}

func (m *Middleware) EnsureToken() error {
	m.TokenMutex.Lock()
	cookiesValid := m.Cookies != ""
	m.TokenMutex.Unlock()

	if cookiesValid {
		return nil
	}

	m.Client.Logger.Info("Authentication cookies missing or expired; fetching new cookies.")
	return m.FetchToken()
}

func (m *Middleware) HTTPRequest(url, method string, payload []byte, headers map[string]string) ([]byte, error) {
	if err := m.EnsureToken(); err != nil {
		return nil, fmt.Errorf("failed to ensure authentication cookies: %w", err)
	}

	headersCopy := make(map[string]string, len(headers))
	for k, v := range headers {
		headersCopy[k] = v
	}
	m.TokenMutex.Lock()
	headersCopy["Cookie"] = m.Cookies
	m.TokenMutex.Unlock()

	m.Client.Logger.Info("Sending authorized request with cookies %v", url)
	return m.Client.HTTPRequest(url, method, payload, headersCopy)
}
