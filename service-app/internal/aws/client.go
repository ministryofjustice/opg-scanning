package aws

import (
	"context"
	"fmt"
	"io"
	"time"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/ministryofjustice/opg-scanning/config"
)

type AwsClientInterface interface {
	GetSecretValue(ctx context.Context, secretName string) (string, error)
}

type AwsClient struct {
	config         *config.Config
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
		o.UsePathStyle = appConfig.App.Environment == "local"
	})

	return &AwsClient{
		config:         appConfig,
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

func (a *AwsClient) PersistFormData(ctx context.Context, body io.Reader, docType string) (string, error) {
	bucketName := a.config.Aws.JobsQueueBucket
	if bucketName == "" {
		return "", fmt.Errorf("JOBSQUEUE_BUCKET is not set")
	}

	// Generate the filename using the required format
	currentTime := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("FORM_DDC_%s_%s.xml", currentTime, docType)

	// Create the S3 input
	input := &s3.PutObjectInput{
		Bucket:               &bucketName,
		Key:                  &fileName,
		Body:                 body,
		ServerSideEncryption: types.ServerSideEncryptionAes256,
	}

	// Upload the file to S3
	_, err := a.S3.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf(
			"failed to upload object to S3: %w (endpoint: %s, bucket: %s, key: %s)",
			err, a.config.Aws.Endpoint, bucketName, fileName,
		)
	}

	return fileName, nil
}
