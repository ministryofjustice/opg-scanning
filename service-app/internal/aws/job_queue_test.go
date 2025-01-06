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

func TestPHPSerialization(t *testing.T) {
	scannedCaseResponse := &types.ScannedCaseResponse{
		UID: "700000001219",
	}
	fileName := "SET_DDC_20250106093401__LPA_677ba389ab101.xml"

	message := createMessageBody(scannedCaseResponse, fileName)

	// Expected output format
	expectedOutput := `{"content":"a:2:{s:3:\"uid\";s:12:\"700000001219\";s:8:\"filename\";s:45:\"SET_DDC_20250106093401__LPA_677ba389ab101.xml\";}","metadata":{"__name__":"Ddc\\Job\\FormJob"}}`

	// Convert the message to JSON
	messageJson, err := json.Marshal(message)
	if err != nil {
		t.Errorf("Failed to marshal message to JSON: %v", err)
	}

	// Check if the generated output matches the expected output
	if string(messageJson) != expectedOutput {
		t.Errorf("Expected: %s\nGot: %s", expectedOutput, messageJson)
	}
}

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
