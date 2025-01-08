package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

type Middleware struct {
	Client        HttpClientInterface
	Config        *config.Config
	Logger        *logger.Logger
	awsClient     aws.AwsClientInterface
	Token         string
	tokenExpiry   time.Time
	signingSecret string
	ApiUser       string
	mu            sync.RWMutex
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

func (m *Middleware) NewClaims() (Claims, error) {
	if m.ApiUser == "" {
		m.ApiUser = m.Config.Auth.ApiUsername
	}

	return Claims{
		SessionData: m.ApiUser,
		Iat:         time.Now().Unix(),
		Exp:         time.Now().Add(time.Duration(m.Config.Auth.JWTExpiration) * time.Second).Unix(),
	}, nil
}

func (m *Middleware) fetchSigningSecret(ctx context.Context) error {
	if m.signingSecret != "" {
		return nil
	}

	secret, err := m.awsClient.GetSecretValue(ctx, m.Config.Auth.JWTSecretARN)
	if err != nil {
		return fmt.Errorf("failed to fetch signing secret: %w", err)
	}

	m.signingSecret = secret

	return nil
}

func (m *Middleware) generateToken(ctx context.Context) (string, error) {
	err := m.fetchSigningSecret(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch signing secret: %w", err)
	}

	claims, err := m.NewClaims()
	if err != nil {
		return "", fmt.Errorf("failed to create claims: %w", err)
	}

	jwtClaims := jwt.MapClaims{
		"session-data": claims.SessionData,
		"iat":          claims.Iat,
		"exp":          claims.Exp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	signedToken, err := token.SignedString([]byte(m.signingSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	expiry := time.Unix(claims.Exp, 0)

	m.mu.Lock()
	m.Token = signedToken
	m.tokenExpiry = expiry
	m.mu.Unlock()

	return signedToken, nil
}

func (m *Middleware) EnsureToken(ctx context.Context) error {
	m.mu.RLock()
	tokenValid := m.Token != "" && time.Now().Before(m.tokenExpiry)
	m.mu.RUnlock()

	if tokenValid {
		m.Logger.Info("Using cached JWT token.", nil)
		return nil
	}

	// Recheck token validity after acquiring the write lock
	if m.Token != "" && time.Now().Before(m.tokenExpiry) {
		m.Logger.Info("Another goroutine refreshed the token.", nil)
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(m.Config.HTTP.Timeout)*time.Second)
	defer cancel()

	_, err := m.generateToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	return nil
}

// Acts as an HTTP wrapper for existing client with Authorization header set.
func (m *Middleware) HTTPRequest(ctx context.Context, url, method string, payload []byte, headers map[string]string, r *http.Request) ([]byte, error) {
	// Check for JWT token in the cookies if not already in the headers
	// if m.Token == "" {
	// 	// Check if cookie is present
	// 	cookie, err := r.Cookie("membane")
	// 	if err == nil {
	// 		m.mu.RLock()
	// 		m.Token = cookie.Value
	// 		m.mu.RUnlock()
	// 	}
	// }

	// Ensure token is valid
	if err := m.EnsureToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure token: %w", err)
	}

	// Add the Authorization header
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Authorization"] = "Bearer " + m.Token

	// Perform the target HTTP request
	response, err := m.Client.HTTPRequest(ctx, url, method, payload, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	return response, nil
}
