// httpclient/middleware.go
package httpclient

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

// Middleware handles HTTP requests with authorization.
type Middleware struct {
	Client         HttpClientInterface
	Config         *config.Config
	Logger         *logger.Logger
	TokenGenerator auth.TokenGenerator
}

func NewMiddleware(client HttpClientInterface, tokenGenerator auth.TokenGenerator) (*Middleware, error) {
	config := client.GetConfig()
	logger := client.GetLogger()

	// Validate required config fields
	if config.Auth.JWTSecretARN == "" {
		return nil, fmt.Errorf("JWTSecretARN is missing in config")
	}
	if config.Auth.JWTExpiration <= 0 {
		return nil, fmt.Errorf("invalid JWTExpiration in config")
	}

	return &Middleware{
		Client:         client,
		Config:         config,
		Logger:         logger,
		TokenGenerator: tokenGenerator,
	}, nil
}

// Acts as an HTTP wrapper for existing client with Authorization header set.
func (m *Middleware) HTTPRequest(ctx context.Context, url, method string, payload []byte, headers map[string]string) ([]byte, error) {
	// Ensure token is valid using the TokenGenerator from auth package
	if err := m.TokenGenerator.EnsureToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure token: %w", err)
	}

	// Retrieve the token
	token := m.TokenGenerator.GetToken()

	// Add the Authorization header
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Authorization"] = "Bearer " + token

	// Perform the target HTTP request
	response, err := m.Client.HTTPRequest(ctx, url, method, payload, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	return response, nil
}
