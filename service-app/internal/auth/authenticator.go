package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"golang.org/x/crypto/bcrypt"
)

type Authenticator interface {
	Authenticate(w http.ResponseWriter, r *http.Request) (context.Context, string, error)
	ValidateCredentials(ctx context.Context, creds UserLogin) (context.Context, error)
}

type BasicAuthAuthenticator struct {
	awsClient      aws.AwsClientInterface
	CookieHelper   CookieHelper
	TokenGenerator TokenGenerator
}

func NewBasicAuthAuthenticator(awsClient aws.AwsClientInterface, cookieHelper CookieHelper, tokenGenerator TokenGenerator) *BasicAuthAuthenticator {
	return &BasicAuthAuthenticator{
		awsClient:      awsClient,
		CookieHelper:   cookieHelper,
		TokenGenerator: tokenGenerator,
	}
}

func (a *BasicAuthAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) (context.Context, string, error) {
	var creds UserLogin

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&creds); err != nil {
		return nil, "", fmt.Errorf("invalid JSON payload: %w", err)
	}

	// Validate credentials first
	ctx, err := a.ValidateCredentials(r.Context(), creds)
	if err != nil {
		return nil, "", err
	}

	token, expiry, err := a.TokenGenerator.GenerateToken()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	if err := a.CookieHelper.SetTokenInCookie(w, token, expiry); err != nil {
		return nil, "", fmt.Errorf("failed to set cookie: %w", err)
	}

	return ctx, token, nil
}

func (a *BasicAuthAuthenticator) ValidateCredentials(ctx context.Context, user UserLogin) (context.Context, error) {
	var creds = user.User
	if creds.Email == "" || creds.Password == "" {
		return nil, fmt.Errorf("missing email or password")
	}

	// Fetch credentials from AWS
	storedCredentials, err := a.awsClient.FetchCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credentials: %w", err)
	}

	storedHash, ok := storedCredentials[creds.Email]
	if !ok {
		return nil, fmt.Errorf("unknown email: %s", creds.Email)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(creds.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	ctx = context.WithValue(ctx, constants.UserContextKey, creds)
	return ctx, nil
}
