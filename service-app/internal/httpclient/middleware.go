// httpclient/middleware.go
package httpclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

// Middleware handles HTTP requests with authorization.
type Middleware struct {
	client httpClientInterface
	Config *config.Config
	logger *logger.Logger
}

func NewMiddleware(client httpClientInterface) (*Middleware, error) {
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
		client: client,
		Config: config,
		logger: logger,
	}, nil
}

// Acts as an HTTP wrapper for existing client with Authorization header set.
func (m *Middleware) HTTPRequest(ctx context.Context, url, method string, payload []byte, headers map[string]string) ([]byte, error) {
	// Retrieve the token
	token, ok := ctx.Value(constants.UserContextKey).(string)
	if !ok {
		return nil, errors.New("could not fetch user token from context")
	}

	// Add the Authorization header
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Authorization"] = "Bearer " + token

	// Perform the target HTTP request
	response, err := m.client.HTTPRequest(ctx, url, method, payload, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	return response, nil
}
