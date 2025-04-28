package aws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/stretchr/testify/assert"
)

var scannedCaseResponse = &types.ScannedCaseResponse{
	UID: "700000001219",
}

const fileName = "SET_DDC_20250106093401__LPA_677ba389ab101.xml"

func TestPersistFormData_LocalStack(t *testing.T) {
	ctx := context.Background()

	appConfig := config.NewConfig()

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
	assert.Regexp(t, regexp.MustCompile(`^FORM_DDC_\d{14}_\d{6}_TestDoc.xml$`), fileName)

	// Test PersistFormData invalid
	docType = "TestDoc"
	body = bytes.NewReader([]byte("invalid xml"))
	_, err = awsClient.PersistFormData(ctx, body, docType)
	assert.Error(t, err, "PersistFormData should return an error for invalid XML")

	listObjectsOutput, err := awsClient.S3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(appConfig.Aws.JobsQueueBucket),
	})

	assert.NoError(t, err, "Failed to list objects in the bucket")

	var found bool
	for _, object := range listObjectsOutput.Contents {
		if *object.Key == fileName {
			found = true
			break
		}
	}
	assert.True(t, found, fmt.Sprintf("Expected object key '%s' not found in the bucket", fileName))
}

func TestPersistSetData(t *testing.T) {
	ctx := context.Background()

	appConfig := config.NewConfig()

	cfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(appConfig.Aws.Region),
	)
	assert.NoError(t, err, "Failed to load AWS configuration")

	awsClient, err := NewAwsClient(ctx, cfg, appConfig)
	assert.NoError(t, err, "Failed to load AWS client")

	body := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?><Set>test</Set>")
	fileName, err := awsClient.PersistSetData(ctx, body)
	assert.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`^SET_\d{14}_\d{6}.xml$`), fileName)

	listObjectsOutput, err := awsClient.S3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(appConfig.Aws.JobsQueueBucket),
	})

	assert.NoError(t, err, "Failed to list objects in the bucket")

	var found bool
	for _, object := range listObjectsOutput.Contents {
		if *object.Key == fileName {
			found = true
			break
		}
	}
	assert.True(t, found, fmt.Sprintf("Expected object key '%s' not found in the bucket", fileName))
}

func TestAwsQueue_PHPSerialization(t *testing.T) {
	message := createMessageBody(scannedCaseResponse, fileName)

	messageJson, err := json.Marshal(message)
	assert.NoError(t, err, "Failed to marshal message to JSON")

	metadataJson := `{"metadata":{"__name__":"Ddc\\Job\\FormJob"}}`

	var actual map[string]interface{}
	var expected map[string]interface{}

	err = json.Unmarshal(messageJson, &actual)
	assert.NoError(t, err)

	err = json.Unmarshal([]byte(metadataJson), &expected)
	assert.NoError(t, err)

	// Compare metadata
	assert.Equal(t, expected["metadata"], actual["metadata"])

	assert.Contains(t, actual["content"], fileName)
	assert.Contains(t, actual["content"], scannedCaseResponse.UID)
}

func TestAwsQueue_QueueSetForProcessing(t *testing.T) {
	appConfig := config.NewConfig()
	ctx := context.Background()

	// Only run this test if we're in the "local" environment
	if appConfig.App.Environment != "local" {
		t.Skip("Skipping test as it requires localstack")
	}

	cfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(appConfig.Aws.Region),
	)
	assert.NoError(t, err)

	awsClient, err := NewAwsClient(ctx, cfg, appConfig)
	assert.NoError(t, err, "Failed to create AwsQueue instance")
	assert.NotNil(t, awsClient.SQS)

	// Call the method to simulate queuing the message
	messageID, err := awsClient.QueueSetForProcessing(context.Background(), scannedCaseResponse, fileName)
	assert.NoError(t, err, "Failed to queue message")
	assert.NotNil(t, messageID)

	validateMessageInQueue(t, ctx, awsClient.SQS, awsClient.siriusQueueURL)
}

func validateMessageInQueue(t *testing.T, ctx context.Context, sqsClient *sqs.Client, queueUrl string) {
	var output *sqs.ReceiveMessageOutput
	var err error
	// Poll for up to 5 seconds
	timeout := time.After(5 * time.Second)
	tick := time.Tick(500 * time.Millisecond)
	for {
		select {
		case <-timeout:
			t.Fatalf("Timeout waiting for message in queue")
		case <-tick:
			output, err = sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(queueUrl),
				MaxNumberOfMessages: 1,
				WaitTimeSeconds:     1,
			})
			if err != nil {
				t.Fatalf("Failed to receive messages from queue: %v", err)
			}
			if len(output.Messages) > 0 {
				assert.NotNil(t, output.Messages, "Expected at least one message in the queue")
				assert.Len(t, output.Messages, 1, "Expected exactly one message in the queue")
				// Optionally parse and assert message content.
				var receivedMessage map[string]any
				err = json.Unmarshal([]byte(*output.Messages[0].Body), &receivedMessage)
				assert.NoError(t, err)
				return
			}
		}
	}
}
