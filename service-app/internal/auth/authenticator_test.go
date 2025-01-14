package auth

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticatorCredentials(t *testing.T) {
	cfg := config.NewConfig()
	logger := logger.NewLogger(cfg)

	_, authMiddleware, _, _ := PrepareMocks(cfg, logger)

	tests := []struct {
		name    string
		creds   User
		isValid bool
	}{
		{
			"Valid creds",
			User{
				Email:    cfg.Auth.ApiUsername,
				Password: "password",
			},
			true,
		},
		{
			"Invalid creds",
			User{
				Email:    cfg.Auth.ApiUsername,
				Password: "",
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
