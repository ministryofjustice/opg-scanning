package httpclient

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/stretchr/testify/mock"
)

func TestEnsureTokenConcurrency(t *testing.T) {

	logger := *logger.NewLogger()
	mockConfig := config.Config{
		Auth: config.Auth{
			ApiUsername:   "test",
			JWTSecretARN:  "local/jwt-key",
			JWTExpiration: 3600,
			JWTTestSecret: "mock-signing-secret",
		},
	}

	mockAwsClient := new(aws.MockAwsClient)
	mockAwsClient.On("GetSecretValue", mock.Anything, mock.AnythingOfType("string")).
		Return("mock-signing-secret", nil)

	httpClient := NewHttpClient(mockConfig, logger)

	middleware := &Middleware{
		Client:      httpClient,
		Config:      &mockConfig,
		Logger:      &logger,
		tokenExpiry: time.Now().Add(-1 * time.Minute),
		mu:          sync.RWMutex{},
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := middleware.ensureToken(context.Background())
			if err != nil {
				t.Errorf("ensureToken failed: %v", err)
			}
		}()
	}
	wg.Wait()
}
