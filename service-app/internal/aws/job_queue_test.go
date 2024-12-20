package aws

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/stretchr/testify/assert"
)

// Tests the QueueSetForProcessing function using LocalStack.
func TestAwsQueue_QueueSetForProcessing(t *testing.T) {
	cfg := config.NewConfig()

	if cfg.App.Environment != "local" {
		t.Skip("Skipping test as it requires localstack")
	}

	awsQueue, err := NewAwsQueue(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, awsQueue)

	scannedCaseResponse := &types.ScannedCaseResponse{
		UID: "test-uid-123",
	}
	fileName := "test-file.xml"

	messageID, err := awsQueue.QueueSetForProcessing(context.Background(), scannedCaseResponse, fileName)
	assert.NoError(t, err)
	assert.NotNil(t, messageID)

	validateMessageInQueue(t, awsQueue, scannedCaseResponse, fileName)
}

func validateMessageInQueue(t *testing.T, awsQueue *AwsQueue, scannedCaseResponse *types.ScannedCaseResponse, fileName string) {
	// Receive messages from the queue
	output, err := awsQueue.SqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(awsQueue.QueueURL),
		MaxNumberOfMessages: aws.Int64(1),
		WaitTimeSeconds:     aws.Int64(0),
	})
	assert.NoError(t, err)
	assert.Len(t, output.Messages, 1)

	// Parse the message body
	var receivedMessage map[string]interface{}
	err = json.Unmarshal([]byte(*output.Messages[0].Body), &receivedMessage)
	assert.NoError(t, err)

	// Validate message content
	assert.Equal(t, scannedCaseResponse.UID, receivedMessage["uid"])
	assert.Equal(t, fileName, receivedMessage["filename"])
}
