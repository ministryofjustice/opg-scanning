package aws

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	cfg "github.com/ministryofjustice/opg-scanning/config"
	"github.com/stretchr/testify/assert"
)

func TestPersistFormData_LocalStack(t *testing.T) {
	ctx := context.Background()

	appConfig := cfg.NewConfig()

	if appConfig.App.Environment != "local" {
		t.Skip("Skipping test as it requires localstack")
	}

	cfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(appConfig.Aws.Region),
	)
	assert.NoError(t, err, "Failed to load AWS configuration")

	awsClient, err := NewAwsClient(ctx, cfg, appConfig)
	assert.NoError(t, err, "Failed to load AWS client")

	// Test PersistFormData valid
	docType := "TestDoc"
	body := bytes.NewReader([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?><test>test</test>"))
	fileName, err := awsClient.PersistFormData(ctx, body, docType)
	assert.NoError(t, err, "PersistFormData should not return an error")
	assert.Contains(t, fileName, "FORM_DDC_", "Expected file name to start with 'FORM_DDC_'")

	// Test PersistFormData invalid
	docType = "TestDoc"
	body = bytes.NewReader([]byte("invalid xml"))
	_, err = awsClient.PersistFormData(ctx, body, docType)
	assert.Error(t, err, "PersistFormData should return an error for invalid XML")

	currentTime := time.Now().Format("20060102150405")
	expectedKey := fmt.Sprintf("FORM_DDC_%s_%s.xml", currentTime, docType)
	listObjectsOutput, err := awsClient.S3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: awsSdk.String(appConfig.Aws.JobsQueueBucket),
	})

	assert.NoError(t, err, "Failed to list objects in the bucket")

	var found bool
	for _, object := range listObjectsOutput.Contents {
		if *object.Key == expectedKey {
			found = true
			break
		}
	}
	assert.True(t, found, fmt.Sprintf("Expected object key '%s' not found in the bucket", expectedKey))
}
