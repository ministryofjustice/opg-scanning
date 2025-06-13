package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-scanning/config"
)

// Refresh token after 10 minutes
const secretTTL = 10 * time.Minute

type secretsClient interface {
	GetSecretValue(ctx context.Context, secretName string) (string, error)
}

type tokenHelper struct {
	awsClient secretsClient
	config    *config.Config

	// TODO: should prevent these being updated at the same time
	signingSecret   string
	lastSecretFetch time.Time
}

// Generate creates a new JWT token and also returns how many seconds until the
// token expires.
func (tg *tokenHelper) Generate() (string, time.Time, error) {
	if err := tg.fetchSigningSecret(); err != nil {
		return "", time.Time{}, err
	}

	now := time.Now()
	expiry := now.Add(time.Duration(tg.config.Auth.JWTExpiration) * time.Second)

	jwtClaims := jwt.MapClaims{
		"session-data": tg.config.Auth.ApiUsername,
		"iat":          now.Unix(),
		"exp":          expiry.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	signedToken, err := token.SignedString([]byte(tg.signingSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, expiry.Truncate(time.Second), nil
}

func (tg *tokenHelper) Validate(tokenString string) error {
	if err := tg.fetchSigningSecret(); err != nil {
		return err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
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

func (tg *tokenHelper) fetchSigningSecret() error {
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
