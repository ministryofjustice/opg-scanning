package auth

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticateMiddleware(t *testing.T) {
	// @TODO implement test case checking token/cookie validity
}

func TestEnsureTokenConcurrency(t *testing.T) {
	cfg := config.NewConfig()
	// Set a reasonable JWTExpiration for testiing e.g. 60 seconds
	cfg.Auth.JWTExpiration = 60

	logger := logger.NewLogger(cfg)

	_, _, mockAwsClient, _ := PrepareMocks(cfg, logger)
	tokenGenerator := NewJWTTokenGenerator(mockAwsClient, cfg, logger)

	var wg sync.WaitGroup
	numGoroutines := 10
	tokensChan := make(chan string, numGoroutines)
	errorsChan := make(chan error, numGoroutines)

	// Launch multiple goroutines to call EnsureToken concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := tokenGenerator.EnsureToken(context.Background())
			if err != nil {
				errorsChan <- err
				return
			}
			token := tokenGenerator.GetToken()
			tokensChan <- token
		}()
	}
	wg.Wait()
	close(errorsChan)
	close(tokensChan)

	// Check for any errors from goroutines
	for err := range errorsChan {
		t.Errorf("EnsureToken failed: %v", err)
	}

	// Assert that GetSecretValue was called only once
	mockAwsClient.AssertNumberOfCalls(t, "GetSecretValue", 1)

	// Collect all tokens and verify they are the same
	var firstToken string
	for token := range tokensChan {
		if firstToken == "" {
			firstToken = token
			assert.NotEmpty(t, firstToken, "First token should not be empty")
		} else {
			assert.Equal(t, firstToken, token, "All tokens should be identical")
		}
	}

	// verify that the token expiry is in the future
	expiry := tokenGenerator.GetExpiry()
	assert.True(t, expiry.After(time.Now()), "Token expiry should be in the future")

	// verify that tokenExpiry is around now + JWTExpiration
	expectedExpiry := time.Now().Add(time.Duration(cfg.Auth.JWTExpiration) * time.Second)
	assert.WithinDuration(t, expectedExpiry, expiry, 2*time.Second, "Token expiry should be set correctly")

	mockAwsClient.AssertExpectations(t)
}
