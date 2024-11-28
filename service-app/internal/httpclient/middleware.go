package httpclient

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

type Middleware struct {
	Client      HttpClientInterface
	awsClient   aws.AwsClientInterface
	Token       string
	TokenExpiry time.Time
	Config      *config.Config
	Logger      *logger.Logger
}

func NewMiddleware(client HttpClientInterface, awsClient aws.AwsClientInterface) *Middleware {
	config := client.GetConfig()
	logger := client.GetLogger()

	// Validate required config fields
	if config.Auth.JWTSecretARN == "" {
		logger.Error("JWTSecretARN is missing in config")
	}
	if config.Auth.JWTExpiration <= 0 {
		logger.Error("Invalid JWTExpiration in config")
	}

	return &Middleware{
		Client:    client,
		awsClient: awsClient,
		Config:    config,
		Logger:    logger,
	}
}

func (m *Middleware) FetchSigningSecret(ctx context.Context) (string, error) {
	secret, err := m.awsClient.GetSecretValue(ctx, m.Config.Auth.JWTSecretARN)
	if err != nil {
		return "", fmt.Errorf("failed to fetch signing secret: %w", err)
	}
	return secret, nil
}

func (m *Middleware) GenerateToken(ctx context.Context) (string, error) {
	signingSecret, err := m.FetchSigningSecret(ctx)
	if err != nil {
		m.Logger.Error(fmt.Sprintf("Failed to fetch signing secret: %v", err))
		return "", err
	}

	claims := jwt.MapClaims{
		// TODO: identify username
		"session-data": "username",
		"exp":          time.Now().Add(time.Duration(m.Config.Auth.JWTExpiration) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(signingSecret))
	if err != nil {
		m.Logger.Error(fmt.Sprintf("Failed to sign token: %v", err))
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	m.Token = signedToken
	m.TokenExpiry = time.Now().Add(1 * time.Hour)
	return signedToken, nil
}

func (m *Middleware) EnsureToken(ctx context.Context) error {
	if m.Token == "" || time.Now().After(m.TokenExpiry) {
		m.Logger.Info("Generating new JWT token...")
		_, err := m.GenerateToken(ctx)
		if err != nil {
			m.Logger.Error(fmt.Sprintf("Failed to generate token: %v", err))
			return fmt.Errorf("failed to generate token: %w", err)
		}
	} else {
		m.Logger.Info("Using cached JWT token.")
	}
	return nil
}

// Acts as an HTTP wrapper for existing client with Authorization header set.
func (m *Middleware) HTTPRequest(ctx context.Context, url, method string, payload []byte, headers map[string]string) ([]byte, error) {
	// Ensure a valid token is available
	if err := m.EnsureToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure token: %w", err)
	}

	// Add Authorization header
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Authorization"] = "Bearer " + m.Token

	// Copy existing headers for the request
	headersCopy := make(map[string]string, len(headers))
	for k, v := range headers {
		headersCopy[k] = v
	}

	// Perform the HTTP request
	response, err := m.Client.HTTPRequest(ctx, url, method, payload, headersCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	return response, nil
}
