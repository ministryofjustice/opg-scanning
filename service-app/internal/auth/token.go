package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

// Refresh token after 10 minutes
const secretTTL = 10 * time.Minute

type TokenGenerator interface {
	GenerateToken() (string, time.Time, error)
	ValidateToken(string) error
}

type JWTTokenGenerator struct {
	awsClient       aws.AwsClientInterface
	config          *config.Config
	logger          *logger.Logger
	signingSecret   string
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
	return Claims{
		SessionData: tg.config.Auth.ApiUsername,
		Iat:         time.Now().Unix(),
		Exp:         time.Now().Add(time.Duration(tg.config.Auth.JWTExpiration) * time.Second).Unix(),
	}, nil
}

func (tg *JWTTokenGenerator) fetchSigningSecret() error {
	shouldFetch := time.Since(tg.lastSecretFetch) >= secretTTL || tg.signingSecret == ""
	if !shouldFetch {
		return nil
	}

	secret, err := tg.awsClient.GetSecretValue(context.Background(), tg.config.Auth.JWTSecretARN)
	if err != nil {
		return fmt.Errorf("failed to fetch signing secret: %w", err)
	}
	tg.signingSecret = secret
	tg.lastSecretFetch = time.Now()
	return nil
}

// Creates a new JWT token.
func (tg *JWTTokenGenerator) GenerateToken() (string, time.Time, error) {
	if err := tg.fetchSigningSecret(); err != nil {
		return "", time.Time{}, err
	}

	claims, err := tg.NewClaims()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create claims: %w", err)
	}

	jwtClaims := jwt.MapClaims{
		"session-data": claims.SessionData,
		"iat":          claims.Iat,
		"exp":          claims.Exp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	signedToken, err := token.SignedString([]byte(tg.signingSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	expiry := time.Unix(claims.Exp, 0)

	tg.logger.Info("Generated new JWT token.", nil)

	return signedToken, expiry, nil
}

func (tg *JWTTokenGenerator) ValidateToken(tokenString string) error {
	if err := tg.fetchSigningSecret(); err != nil {
		return err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(tg.signingSecret), nil
	}, jwt.WithIssuedAt(), jwt.WithExpirationRequired())

	if err != nil {
		return err
	}

	sessionData := token.Claims.(jwt.MapClaims)["session-data"]
	if sessionData == nil || sessionData == "" {
		return errors.New("session-data claim is required")
	}

	return nil
}
