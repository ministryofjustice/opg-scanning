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

var scannedCaseResponse = &types.ScannedCaseResponse{
	UID: "700000001219",
}

const fileName = "SET_DDC_20250106093401__LPA_677ba389ab101.xml"

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
	cfg := config.NewConfig()

	// Only run this test if we're in the "local" environment
	if cfg.App.Environment != "local" {
		t.Skip("Skipping test as it requires localstack")
	}

	awsQueue, err := NewAwsQueue(cfg)
	assert.NoError(t, err, "Failed to create AwsQueue instance")
	assert.NotNil(t, awsQueue)

	// Call the method to simulate queuing the message
	messageID, err := awsQueue.QueueSetForProcessing(context.Background(), scannedCaseResponse, fileName)
	assert.NoError(t, err, "Failed to queue message")
	assert.NotNil(t, messageID)

	validateMessageInQueue(t, awsQueue)
}

func validateMessageInQueue(t *testing.T, awsQueue *AwsQueue) {
	// Receive messages from the queue
	output, err := awsQueue.SqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(awsQueue.QueueURL),
		MaxNumberOfMessages: aws.Int64(1),
		WaitTimeSeconds:     aws.Int64(0),
	})
	assert.NoError(t, err)
	assert.NotNil(t, output.Messages, "Expected at least one message in the queue")
	assert.Len(t, output.Messages, 1, "Expected exactly one message in the queue")

	// Parse the message body
	var receivedMessage map[string]interface{}
	err = json.Unmarshal([]byte(*output.Messages[0].Body), &receivedMessage)
	assert.NoError(t, err)
}
