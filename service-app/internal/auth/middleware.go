package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"golang.org/x/crypto/bcrypt"
)

type Middleware struct {
	awsClient        aws.AwsClientInterface
	clientMiddleware *httpclient.Middleware
	config           *config.Config
	logger           *logger.Logger
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(awsClient aws.AwsClientInterface, clientMiddleware *httpclient.Middleware, config *config.Config, logger *logger.Logger) *Middleware {
	return &Middleware{
		awsClient:        awsClient,
		clientMiddleware: clientMiddleware,
		config:           config,
		logger:           logger,
	}
}

type UserCredentials struct {
	User User `json:"user"`
}

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (m *Middleware) FetchStoredCredentials(ctx context.Context) (string, string, error) {
	// Retrieve the secret value from AWS SSM
	secretValue, err := m.awsClient.GetSsmValue(ctx, m.config.Auth.CredentialsARN)
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve secret from AWS: %w", err)
	}

	secretValue = strings.TrimPrefix(secretValue, "kms:alias/aws/ssm:")

	var emailPasswordMap map[string]string
	if err := json.Unmarshal([]byte(secretValue), &emailPasswordMap); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	for email, password := range emailPasswordMap {
		return email, password, nil
	}

	return "", "", fmt.Errorf("no valid credentials found in secret")
}

func (m *Middleware) ValidateCredentials(email, password string) (bool, error) {
	storedEmail, storedHash, err := m.FetchStoredCredentials(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	// Debug stored email
	m.logger.Info("Stored email: "+storedEmail, nil)
	if email != storedEmail {
		return false, fmt.Errorf("invalid email")
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		return false, fmt.Errorf("invalid credentials")
	}

	return true, nil
}

func (m *Middleware) IssueTokenAndSetCookie(w http.ResponseWriter, r *http.Request, clientID string) error {
	// Ensure the token is valid or generate a new one
	err := m.clientMiddleware.EnsureToken(r.Context())
	if err != nil {
		return fmt.Errorf("failed to ensure token: %w", err)
	}
	token := m.clientMiddleware.Token

	http.SetCookie(w, &http.Cookie{
		Name:     "membrane",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(m.config.Auth.JWTExpiration) * time.Second),
		HttpOnly: true,
		Secure:   m.config.App.Environment != "local",
	})

	return nil
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var credentials UserCredentials
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&credentials); err != nil {
			m.respondWithError(w, http.StatusBadRequest, "Invalid JSON", err)
			return
		}

		email := credentials.User.Email
		password := credentials.User.Password

		if email == "" || password == "" {
			m.respondWithError(w, http.StatusUnauthorized, "Unauthorized: Missing credentials", nil)
			return
		}

		valid, err := m.ValidateCredentials(email, password)
		if err != nil || !valid {
			m.respondWithError(w, http.StatusUnauthorized, "Unauthorized: Invalid credentials", err)
			return
		}

		// Set the API user
		m.clientMiddleware.ApiUser = email

		// Issue a JWT token and set it in a cookie
		err = m.IssueTokenAndSetCookie(w, r, email)
		if err != nil {
			m.respondWithError(w, http.StatusInternalServerError, "Failed to issue token", err)
			return
		}

		// Proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) CheckAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure the token is valid
		err := m.clientMiddleware.EnsureToken(r.Context())
		if err != nil {
			m.respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	m.logger.Error("%s: %v", nil, message, err)
	http.Error(w, message, statusCode)
}
