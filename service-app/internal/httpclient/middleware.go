package httpclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

type Middleware struct {
	Client      HttpClientInterface
	Config      *config.Config
	Logger      *logger.Logger
	awsClient   aws.AwsClientInterface
	token       string
	tokenExpiry time.Time
	mu          sync.RWMutex
}

type Claims struct {
	SessionData string
	Iat         int64
	Exp         int64
}

func NewMiddleware(client HttpClientInterface, awsClient aws.AwsClientInterface) (*Middleware, error) {
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
		Client:    client,
		awsClient: awsClient,
		Config:    config,
		Logger:    logger,
	}, nil
}

func NewClaims(cfg config.Config) (Claims, error) {
	if cfg.Auth.ApiUsername == "" {
		return Claims{}, fmt.Errorf("middleware configuration is missing or invalid")
	}

	return Claims{
		SessionData: cfg.Auth.ApiUsername,
		Iat:         time.Now().Unix(),
		Exp:         time.Now().Add(time.Duration(cfg.Auth.JWTExpiration) * time.Second).Unix(),
	}, nil
}

func (m *Middleware) fetchSigningSecret(ctx context.Context) (string, error) {
	// Check hardcoded signing secret
	if m.Config.Auth.JWTTestSecret != "" {
		return m.Config.Auth.JWTTestSecret, nil
	}

	secret, err := m.awsClient.GetSecretValue(ctx, m.Config.Auth.JWTSecretARN)
	if err != nil {
		return "", fmt.Errorf("failed to fetch signing secret: %w", err)
	}
	return secret, nil
}

func (m *Middleware) generateToken(ctx context.Context) (string, error) {
	signingSecret, err := m.fetchSigningSecret(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch signing secret: %w", err)
	}

	claims, err := NewClaims(*m.Config)
	if err != nil {
		return "", fmt.Errorf("failed to create claims: %w", err)
	}

	m.Logger.Info("Generating JWT token...")

	jwtClaims := jwt.MapClaims{
		"session-data": claims.SessionData,
		"iat":          claims.Iat,
		"exp":          claims.Exp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	signedToken, err := token.SignedString([]byte(signingSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	expiry := time.Unix(claims.Exp, 0)

	m.mu.Lock()
	m.token = signedToken
	m.tokenExpiry = expiry
	m.mu.Unlock()

	return signedToken, nil
}

func (m *Middleware) ensureToken(ctx context.Context) error {
	m.mu.RLock()
	tokenValid := m.token != "" && time.Now().Before(m.tokenExpiry)
	m.mu.RUnlock()

	if tokenValid {
		m.Logger.Info("Using cached JWT token.")
		return nil
	}

	// Recheck token validity after acquiring the write lock
	if m.token != "" && time.Now().Before(m.tokenExpiry) {
		m.Logger.Info("Another goroutine refreshed the token.")
		return nil
	}

	// Generate a new token
	m.Logger.Info("Token invalid or expired. Attempting to generate a new one...")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := m.generateToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	return nil
}

// Acts as an HTTP wrapper for existing client with Authorization header set.
func (m *Middleware) HTTPRequest(ctx context.Context, url, method string, payload []byte, headers map[string]string) ([]byte, error) {
	// Ensure a valid token is available
	if err := m.ensureToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure token: %w", err)
	}

	// Safely initialize headers if nil
	if headers == nil {
		headers = make(map[string]string)
	}

	// Add Authorization header
	headers["Authorization"] = "Bearer " + m.token

	// Perform the HTTP request
	response, err := m.Client.HTTPRequest(ctx, url, method, payload, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	return response, nil
}
