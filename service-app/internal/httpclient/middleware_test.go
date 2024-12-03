package httpclient

import (
	"context"
	"sync"
	"testing"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

func TestEnsureTokenConcurrency(t *testing.T) {
	ctx := context.Background()
	logger := *logger.NewLogger()
	cfg := config.NewConfig()
	mockConfig := config.Config{
		HTTP: cfg.HTTP,
		App:  cfg.App,
		Aws:  cfg.Aws,
		Auth: config.Auth{
			ApiUsername:   "test",
			JWTSecretARN:  "local/jwt-key",
			JWTExpiration: 3600,
		},
	}

	// TODO: Check if we can integration AWS client during git actions workflow
	// For now skip the test
	if cfg.App.Environment != "local" {
		t.Skip("Skipping test as it requires localstack and RUN_LOCAL_TESTS is not set to true")
	}

	// Log mockConfig
	// t.Logf("mockConfig: %+v", mockConfig)

	// Load AWS configuration
	awsCfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(cfg.Aws.Region),
	)
	if err != nil {
		t.Errorf("Failed to load AWS config %v", err)
		return
	}
	// Initialize AwsClient
	awsClient, err := aws.NewAwsClient(ctx, awsCfg, &mockConfig)
	if err != nil {
		t.Errorf("failed to initialize AWS clients: %v", err)
	}
	httpClient := NewHttpClient(mockConfig, logger)

	middleware := &Middleware{
		Client:      httpClient,
		Config:      &mockConfig,
		Logger:      &logger,
		awsClient:   awsClient,
		tokenExpiry: time.Now().Add(time.Hour),
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
