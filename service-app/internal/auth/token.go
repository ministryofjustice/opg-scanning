package auth

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

// Refresh token after 10 minutes
const secretTTL = 10 * time.Minute

type TokenGenerator interface {
	EnsureToken(ctx context.Context) error
	GetToken() string
	GetExpiry() time.Time
}

type JWTTokenGenerator struct {
	awsClient       aws.AwsClientInterface
	config          *config.Config
	logger          *logger.Logger
	signingSecret   string
	ApiUser         string
	mu              sync.RWMutex
	token           string
	tokenExpiry     time.Time
	lastSecretFetch time.Time
}

type Claims struct {
	SessionData string `json:"session-data"`
	Iat         int64  `json:"iat"`
	Exp         int64  `json:"exp"`
}

func NewJWTTokenGenerator(awsClient aws.AwsClientInterface, config *config.Config, logger *logger.Logger) *JWTTokenGenerator {
	return &JWTTokenGenerator{
		awsClient: awsClient,
		config:    config,
		logger:    logger,
	}
}

func (tg *JWTTokenGenerator) NewClaims() (Claims, error) {
	if tg.ApiUser == "" {
		tg.ApiUser = tg.config.Auth.ApiUsername
	}

	return Claims{
		SessionData: tg.ApiUser,
		Iat:         time.Now().Unix(),
		Exp:         time.Now().Add(time.Duration(tg.config.Auth.JWTExpiration) * time.Second).Unix(),
	}, nil
}

func (tg *JWTTokenGenerator) fetchSigningSecret(ctx context.Context) error {
	shouldFetch := time.Since(tg.lastSecretFetch) >= secretTTL || tg.signingSecret == ""
	if !shouldFetch {
		return nil
	}

	secret, err := tg.awsClient.GetSecretValue(ctx, tg.config.Auth.JWTSecretARN)
	if err != nil {
		return fmt.Errorf("failed to fetch signing secret: %w", err)
	}
	tg.signingSecret = secret
	tg.lastSecretFetch = time.Now()
	return nil
}

// Creates a new JWT token.
func (tg *JWTTokenGenerator) generateToken(ctx context.Context) (string, error) {
	if err := tg.fetchSigningSecret(ctx); err != nil {
		return "", err
	}

	claims, err := tg.NewClaims()
	if err != nil {
		return "", fmt.Errorf("failed to create claims: %w", err)
	}

	jwtClaims := jwt.MapClaims{
		"session-data": claims.SessionData,
		"iat":          claims.Iat,
		"exp":          claims.Exp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	signedToken, err := token.SignedString([]byte(tg.signingSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	expiry := time.Unix(claims.Exp, 0)

	tg.token = signedToken
	tg.tokenExpiry = expiry

	tg.logger.Info("Generated new JWT token.", nil)

	return signedToken, nil
}

// Ensures that a valid token exists, generating a new one if necessary.
func (tg *JWTTokenGenerator) EnsureToken(ctx context.Context) error {
	// First, acquire a read lock to check if the token is valid
	tg.mu.RLock()
	tokenValid := tg.token != "" && time.Now().Before(tg.tokenExpiry)
	tg.mu.RUnlock()

	if tokenValid {
		return nil
	}

	tg.mu.Lock()
	defer tg.mu.Unlock()

	// Recheck the token validity under the write lock
	if tg.token != "" && time.Now().Before(tg.tokenExpiry) {
		tg.logger.Info("Another goroutine refreshed the token.", nil)
		return nil
	}

	// Generate a new token
	ctx, cancel := context.WithTimeout(ctx, time.Duration(tg.config.HTTP.Timeout)*time.Second)
	defer cancel()

	_, err := tg.generateToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	return nil
}

func (tg *JWTTokenGenerator) GetToken() string {
	tg.mu.RLock()
	defer tg.mu.RUnlock()
	return tg.token
}

func (tg *JWTTokenGenerator) GetExpiry() time.Time {
	tg.mu.RLock()
	defer tg.mu.RUnlock()
	return tg.tokenExpiry
}
