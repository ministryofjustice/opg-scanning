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
	cfg := config.NewConfig()
	logger := *logger.NewLogger(cfg)

	mockAwsClient := new(aws.MockAwsClient)
	mockAwsClient.On("GetSecretValue", mock.Anything, "local/jwt-key").Return("mysupersecrettestkeythatis128bits", nil)

	httpClient := NewHttpClient(*cfg, logger)

	middleware := &Middleware{
		Client:      httpClient,
		Config:      cfg,
		Logger:      &logger,
		awsClient:   mockAwsClient,
		tokenExpiry: time.Now().Add(time.Hour),
		ApiUser:     "test",
		mu:          sync.RWMutex{},
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := middleware.EnsureToken(context.Background())
			if err != nil {
				t.Errorf("ensureToken failed: %v", err)
			}
		}()
	}
	wg.Wait()
}
