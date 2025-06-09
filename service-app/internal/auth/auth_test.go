package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

var expectedError = errors.New("problem")

func TestAuthAuthenticate(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", strings.NewReader(`{"user":{"email":"john.doe@example.com","password":"not-a-password"}}`))

	hash, _ := bcrypt.GenerateFromPassword([]byte("not-a-password"), 0)
	now := time.Now().UTC()

	credentials := newMockCredentialsClient(t)
	credentials.EXPECT().
		FetchCredentials(r.Context()).
		Return(map[string]string{"john.doe@example.com": string(hash)}, nil)

	tokens := newMockTokens(t)
	tokens.EXPECT().
		Generate().
		Return("a-token", now, nil)

	auth := &Auth{
		credentials: credentials,
		tokens:      tokens,
	}

	user, err := auth.Authenticate(w, r)
	assert.NoError(t, err)
	assert.Equal(t, AuthenticatedUser{Email: "john.doe@example.com", Token: "a-token"}, user)

	resp := w.Result()
	assert.Equal(t, fmt.Sprintf("membrane=a-token; Path=/; Expires=%s; HttpOnly; SameSite=Strict", now.Format(http.TimeFormat)), resp.Header.Get("Set-Cookie"))
}

func TestAuthAuthenticate_TokenCannotBeGenerated(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", strings.NewReader(`{"user":{"email":"john.doe@example.com","password":"not-a-password"}}`))

	hash, _ := bcrypt.GenerateFromPassword([]byte("not-a-password"), 0)

	credentials := newMockCredentialsClient(t)
	credentials.EXPECT().
		FetchCredentials(mock.Anything).
		Return(map[string]string{"john.doe@example.com": string(hash)}, nil)

	tokens := newMockTokens(t)
	tokens.EXPECT().
		Generate().
		Return("", time.Time{}, expectedError)

	auth := &Auth{
		credentials: credentials,
		tokens:      tokens,
	}

	_, err := auth.Authenticate(w, r)
	assert.ErrorIs(t, err, expectedError)
}

func TestAuthAuthenticate_CredentialsCannotBeFetched(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", strings.NewReader(`{"user":{"email":"john.doe@example.com","password":"not-a-password"}}`))

	credentials := newMockCredentialsClient(t)
	credentials.EXPECT().
		FetchCredentials(mock.Anything).
		Return(nil, expectedError)

	auth := &Auth{
		credentials: credentials,
	}

	_, err := auth.Authenticate(w, r)
	assert.ErrorIs(t, err, expectedError)
}

func TestAuthAuthenticate_InvalidCredentials(t *testing.T) {
	testcases := map[string]struct {
		body  string
		error string
	}{
		"missing email": {
			body:  `{"user":{"password":"not-a-password"}}`,
			error: "missing email or password",
		},
		"missing password": {
			body:  `{"user":{"email":"john.doe@example.com"}}`,
			error: "missing email or password",
		},
		"incorrect email": {
			body:  `{"user":{"email":"who@example.com","password":"not-a-password"}}`,
			error: "unknown email: who@example.com",
		},
		"incorrect password": {
			body:  `{"user":{"email":"john.doe@example.com","password":"wrong"}}`,
			error: "invalid credentials",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "", strings.NewReader(tc.body))

			hash, _ := bcrypt.GenerateFromPassword([]byte("not-a-password"), 0)

			credentials := newMockCredentialsClient(t)
			credentials.EXPECT().
				FetchCredentials(mock.Anything).
				Return(map[string]string{"john.doe@example.com": string(hash)}, nil).
				Maybe()

			auth := &Auth{
				credentials: credentials,
			}

			_, err := auth.Authenticate(w, r)
			assert.ErrorContains(t, err, tc.error)
		})
	}
}

func TestAuthUse(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", strings.NewReader(`{"user":{"email":"john.doe@example.com","password":"not-a-password"}}`))
	r.AddCookie(&http.Cookie{Name: cookieName, Value: "a-token"})

	tokens := newMockTokens(t)
	tokens.EXPECT().
		Validate("a-token").
		Return(nil)

	auth := &Auth{
		tokens: tokens,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, context.WithValue(context.Background(), constants.TokenContextKey, "a-token"), r.Context())

		w.WriteHeader(http.StatusTeapot)
	})
	auth.Use(handler)(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestAuthUse_InvalidToken(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", strings.NewReader(`{"user":{"email":"john.doe@example.com","password":"not-a-password"}}`))
	r.AddCookie(&http.Cookie{Name: cookieName, Value: "a-token"})

	tokens := newMockTokens(t)
	tokens.EXPECT().
		Validate("a-token").
		Return(expectedError)

	var logBuffer bytes.Buffer
	logger := &logger.Logger{
		SlogLogger: slog.New(slog.NewTextHandler(&logBuffer, nil)),
	}

	auth := &Auth{
		tokens: tokens,
		logger: logger,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, context.WithValue(context.Background(), constants.TokenContextKey, "a-token"), r.Context())

		w.WriteHeader(http.StatusTeapot)
	})
	auth.Use(handler)(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	assert.Regexp(t, `^time=[0-9TZ\-:.+]+ level=ERROR msg="Unauthorized: Invalid token: problem"
$`, logBuffer.String())
}
