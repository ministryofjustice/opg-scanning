package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticatorCredentials(t *testing.T) {
	cfg := config.NewConfig()
	logger := logger.GetLogger(cfg)

	_, authMiddleware, _, _ := PrepareMocks(cfg, logger)

	tests := []struct {
		name    string
		creds   userLogin
		isValid bool
	}{
		{
			"Valid creds",
			userLogin{
				User: User{
					Email:    cfg.Auth.ApiUsername,
					Password: "password",
				},
			},
			true,
		},
		{
			"Invalid creds",
			userLogin{
				User: User{
					Email:    cfg.Auth.ApiUsername,
					Password: "",
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authenticator := authMiddleware.Authenticator
			_, err := authenticator.ValidateCredentials(context.Background(), tt.creds)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	cfg := config.NewConfig()
	logger := logger.GetLogger(cfg)

	_, authMiddleware, _, _ := PrepareMocks(cfg, logger)

	w := httptest.NewRecorder()
	request := &http.Request{
		Body: io.NopCloser(strings.NewReader(fmt.Sprintf(`{"user":{"email":"%s","password":"password"}}`, cfg.Auth.ApiUsername))),
	}

	ctx, token, err := authMiddleware.Authenticator.Authenticate(w, request)
	assert.Nil(t, err)
	assert.NotNil(t, ctx)

	assert.Contains(t, w.Header().Get("Set-Cookie"), fmt.Sprintf("membrane=%s;", token))
}
