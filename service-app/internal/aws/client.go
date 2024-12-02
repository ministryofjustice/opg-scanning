package aws

import (
	"context"
	"fmt"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/ministryofjustice/opg-scanning/config"
)

type AwsClientInterface interface {
	GetSecretValue(ctx context.Context, secretName string) (string, error)
}

type AwsClient struct {
	SecretsManager *secretsmanager.Client
	S3             *s3.Client
}

// Initializes all required AWS service clients.
func NewAwsClient(ctx context.Context, cfg awsSdk.Config, appConfig *config.Config) (*AwsClient, error) {
	// Use the same endpoint for all services
	customEndpoint := appConfig.Aws.Endpoint
	if customEndpoint == "" {
		return nil, fmt.Errorf("AWS_ENDPOINT is not set")
	}

	smClient := secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = &customEndpoint
	})

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = &customEndpoint
	})

	return &AwsClient{
		SecretsManager: smClient,
		S3:             s3Client,
	}, nil
}

// Fetch secret value from Secrets Manager
func (a *AwsClient) GetSecretValue(ctx context.Context, secretName string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}
	output, err := a.SecretsManager.GetSecretValue(ctx, input)
	if err != nil {
		return "", err
	}
	return *output.SecretString, nil
}
