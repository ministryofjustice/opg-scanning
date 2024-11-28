package httpclient

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEnsureTokenWithAuthorization(t *testing.T) {
	ctx := context.Background()

	// Mock configuration
	mockConfig := &config.Config{
		Auth: config.Auth{
			JWTSecretARN:  "jwt-secret",
			JWTExpiration: 300,
		},
	}

	// Properly initialize the logger
	mockLogger := logger.NewLogger()

	// Mock HTTP Client
	mockHttpClient := new(MockHttpClient)
	mockHttpClient.On("GetConfig").Return(mockConfig)
	mockHttpClient.On("GetLogger").Return(mockLogger)

	// Mock AWS Client
	mockAwsClient := new(aws.MockAwsClient)
	mockAwsClient.On("GetSecretValue", mock.Anything, "jwt-secret").
		Return("mock-signing-secret", nil)

	// Initialize Middleware
	middleware := NewMiddleware(mockHttpClient, mockAwsClient)

	// Test generating a new token and verifying Authorization header
	t.Run("EnsureToken generates Authorization header", func(t *testing.T) {
		headers := map[string]string{"Custom-Header": "value"}

		// Simulate HTTPRequest with the middleware
		mockHttpClient.On("HTTPRequest",
			mock.Anything,
			"http://example.com",
			"POST",
			[]byte{},
			mock.MatchedBy(func(h map[string]string) bool {
				auth, ok := h["Authorization"]
				return ok && len(auth) > 7 && auth[:7] == "Bearer "
			}),
		).Return([]byte("mock-response"), nil)

		// Call HTTPRequest
		response, err := middleware.HTTPRequest(ctx, "http://example.com", "POST", []byte{}, headers)
		require.NoError(t, err)
		require.Equal(t, []byte("mock-response"), response)

		// Assert that EnsureToken was called, and a token was generated
		require.NotEmpty(t, middleware.Token)
		require.WithinDuration(t, time.Now().Add(1*time.Hour), middleware.TokenExpiry, 1*time.Second)

		// Assert mock HTTP client calls
		mockHttpClient.AssertCalled(t, "HTTPRequest",
			mock.Anything,
			"http://example.com",
			"POST",
			[]byte{},
			mock.MatchedBy(func(h map[string]string) bool { // Ensure headers matched
				auth, ok := h["Authorization"]
				return ok && len(auth) > 7 && auth[:7] == "Bearer "
			}),
		)
	})

	// Assert AWS Client calls
	mockAwsClient.AssertCalled(t, "GetSecretValue", mock.Anything, "jwt-secret")
	mockHttpClient.AssertCalled(t, "GetConfig")
	mockHttpClient.AssertCalled(t, "GetLogger")
}
