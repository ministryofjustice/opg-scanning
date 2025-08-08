package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/config"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"golang.org/x/crypto/bcrypt"
)

const cookieName = "membrane"

type tokens interface {
	Generate() (string, time.Time, error)
	Validate(string) error
}

type credentialsClient interface {
	FetchCredentials(ctx context.Context) (map[string]string, error)
}

func New(appConfig *config.Config, logger *logger.Logger, awsClient aws.AwsClientInterface) *Auth {
	return &Auth{
		tokens: &tokenHelper{
			awsClient: awsClient,
			config:    appConfig,
		},
		logger:       logger,
		credentials:  awsClient,
		secureCookie: appConfig.App.Environment != "local",
	}
}

type Auth struct {
	tokens       tokens
	credentials  credentialsClient
	logger       *logger.Logger
	secureCookie bool
}

func (a *Auth) Authenticate(w http.ResponseWriter, r *http.Request) (AuthenticatedUser, error) {
	var creds login
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		return AuthenticatedUser{}, fmt.Errorf("invalid JSON payload: %w", err)
	}

	// Validate credentials first
	if err := a.validateCredentials(r.Context(), creds.User); err != nil {
		return AuthenticatedUser{}, err
	}

	token, expiry, err := a.tokens.Generate()
	if err != nil {
		return AuthenticatedUser{}, fmt.Errorf("failed to generate token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Expires:  expiry,
		Path:     "/",
		HttpOnly: true,
		Secure:   a.secureCookie,
		SameSite: http.SameSiteStrictMode,
	})

	return AuthenticatedUser{
		Email: creds.User.Email,
		Token: token,
	}, nil
}

func (a *Auth) Check(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			a.respondWithError(r.Context(), w, http.StatusUnauthorized, "Unauthorized: Missing token", err)
			return
		}

		token := cookie.Value

		if err := a.tokens.Validate(token); err != nil {
			a.respondWithError(r.Context(), w, http.StatusUnauthorized, "Unauthorized: Invalid token", err)
			return
		}

		ctx := context.WithValue(r.Context(), constants.TokenContextKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (a *Auth) validateCredentials(ctx context.Context, user loginUser) error {
	if user.Email == "" || user.Password == "" {
		return fmt.Errorf("missing email or password")
	}

	storedCredentials, err := a.credentials.FetchCredentials(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch credentials: %w", err)
	}

	storedHash, ok := storedCredentials[user.Email]
	if !ok {
		return fmt.Errorf("unknown email: %s", user.Email)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(user.Password)); err != nil {
		return fmt.Errorf("invalid credentials")
	}

	return nil
}

func (a *Auth) respondWithError(ctx context.Context, w http.ResponseWriter, statusCode int, message string, err error) {
	a.logger.ErrorContext(ctx, fmt.Sprintf("%s: %v", message, err))
	http.Error(w, message, statusCode)
}
