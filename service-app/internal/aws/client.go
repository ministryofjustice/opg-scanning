package aws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/ministryofjustice/opg-scanning/config"
	appTypes "github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type AwsClientInterface interface {
	GetSecretValue(ctx context.Context, secretName string) (string, error)
	FetchCredentials(ctx context.Context) (map[string]string, error)
	PersistFormData(ctx context.Context, body io.Reader, docType string) (string, error)
	PersistSetData(ctx context.Context, body []byte) (string, error)
	QueueSetForProcessing(ctx context.Context, scannedCaseResponse *appTypes.ScannedCaseResponse, fileName string) (MessageID *string, err error)
}

type AwsClient struct {
	config         *config.Config
	siriusQueueURL string
	SecretsManager *secretsmanager.Client
	SSM            *ssm.Client
	S3             *s3.Client
	SQS            *sqs.Client
}

// Initializes all required AWS service clients.
func NewAwsClient(ctx context.Context, cfg aws.Config, appConfig *config.Config) (*AwsClient, error) {
	// Use the same endpoint for all services
	var customEndpoint *string
	if appConfig.Aws.Endpoint != "" {
		customEndpoint = aws.String(appConfig.Aws.Endpoint)
	}

	smClient := secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = customEndpoint
	})

	SsmClient := ssm.NewFromConfig(cfg, func(o *ssm.Options) {
		o.BaseEndpoint = customEndpoint
	})

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = customEndpoint
		o.UsePathStyle = appConfig.App.Environment == "local"
	})

	sqsClient := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = customEndpoint
	})

	return &AwsClient{
		config:         appConfig,
		siriusQueueURL: appConfig.Aws.JobsQueueURL,
		SecretsManager: smClient,
		SSM:            SsmClient,
		S3:             s3Client,
		SQS:            sqsClient,
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

// Fetch secret value from SSM Parameter Store
func (a *AwsClient) GetSsmValue(ctx context.Context, secretName string) (string, error) {
	input := &ssm.GetParameterInput{
		Name:           &secretName,
		WithDecryption: aws.Bool(true),
	}

	output, err := a.SSM.GetParameter(ctx, input)
	if err != nil {
		return "", err
	}

	return *output.Parameter.Value, nil
}

func (a *AwsClient) PersistFormData(ctx context.Context, body io.Reader, docType string) (string, error) {
	bucketName := a.config.Aws.JobsQueueBucket
	if bucketName == "" {
		return "", fmt.Errorf("JOBSQUEUE_BUCKET is not set")
	}

	// Generate the filename using the required format
	currentTime := time.Now().Format("20060102150405.000000")
	currentTime = strings.Replace(currentTime, ".", "_", 1)
	fileName := fmt.Sprintf("FORM_DDC_%s_%s.xml", currentTime, docType)

	// Check body is valid XML before S3 input
	bodyBytes, bodyErr := io.ReadAll(body)
	if bodyErr != nil {
		return "", fmt.Errorf("failed to read body: %w", bodyErr)
	}
	if bodyErr = util.IsValidXML(bodyBytes); bodyErr != nil {
		return "", fmt.Errorf("invalid XML: %w", bodyErr)
	}

	// Since we consume the reader for validation, create a new reader from the buffered data
	readerForS3 := bytes.NewReader(bodyBytes)

	// Create the S3 input
	input := &s3.PutObjectInput{
		Bucket:               &bucketName,
		Key:                  &fileName,
		Body:                 readerForS3,
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		SSEKMSKeyId:          &a.config.Aws.JobsQueueBucketKmsKey,
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

func (a *AwsClient) PersistSetData(ctx context.Context, body []byte) (string, error) {
	bucketName := a.config.Aws.JobsQueueBucket
	if bucketName == "" {
		return "", fmt.Errorf("JOBSQUEUE_BUCKET is not set")
	}

	currentTime := time.Now().Format("20060102150405.000000")
	currentTime = strings.Replace(currentTime, ".", "_", 1)
	fileName := fmt.Sprintf("SET_%s.xml", currentTime)

	// Create a new reader from the buffered data
	readerForS3 := bytes.NewReader(body)

	input := &s3.PutObjectInput{
		Bucket:               &bucketName,
		Key:                  &fileName,
		Body:                 readerForS3,
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		SSEKMSKeyId:          &a.config.Aws.JobsQueueBucketKmsKey,
	}

	_, err := a.S3.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf(
			"failed to upload object to S3: %w (endpoint: %s, bucket: %s, key: %s)",
			err, a.config.Aws.Endpoint, bucketName, fileName,
		)
	}

	return fileName, nil
}

func (a *AwsClient) FetchCredentials(ctx context.Context) (map[string]string, error) {
	secretValue, err := a.GetSsmValue(ctx, a.config.Auth.CredentialsARN)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve secret from AWS: %w", err)
	}

	secretValue = strings.TrimPrefix(secretValue, "kms:alias/aws/ssm:")

	var credentials map[string]string
	if err := json.Unmarshal([]byte(secretValue), &credentials); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	if len(credentials) == 0 {
		return nil, fmt.Errorf("no credentials found in secret")
	}

	return credentials, nil
}

func (a *AwsClient) QueueSetForProcessing(ctx context.Context, scannedCaseResponse *appTypes.ScannedCaseResponse, fileName string) (MessageID *string, err error) {
	message := createMessageBody(scannedCaseResponse, fileName)
	messageJson, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message to JSON: %w", err)
	}

	// Send the message to the SQS queue
	input := &sqs.SendMessageInput{
		QueueUrl:       aws.String(a.siriusQueueURL),
		MessageBody:    aws.String(string(messageJson)),
		MessageGroupId: aws.String(scannedCaseResponse.UID),
	}

	output, err := a.SQS.SendMessage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to send message to SQS queue: %w", err)
	}

	return output.MessageId, nil
}

func createMessageBody(scannedCaseResponse *appTypes.ScannedCaseResponse, fileName string) map[string]interface{} {
	// Create a message structure
	content := map[string]interface{}{
		"uid":      scannedCaseResponse.UID,
		"filename": fileName,
	}

	// Create the final message structure
	message := map[string]interface{}{
		"content": util.PhpSerialize(content),
		"metadata": map[string]interface{}{
			"__name__": "Ddc\\Job\\FormJob",
		},
	}

	return message
}
