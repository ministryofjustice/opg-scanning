package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type AwsClientInterface interface {
	GetSecretValue(ctx context.Context, secretName string) (string, error)
}

type AwsClient struct {
	SecretsManager *secretsmanager.Client
	S3             *s3.Client
}

// Initializes all required AWS service clients.
func NewAwsClient(ctx context.Context, cfg aws.Config) (*AwsClient, error) {
	smClient := secretsmanager.NewFromConfig(cfg)

	s3Client := s3.NewFromConfig(cfg)

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
