package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type AwsQueue struct {
	SqsClient *sqs.SQS
	QueueURL  string
}

func NewAwsQueue(cfg *config.Config) (*AwsQueue, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(cfg.Aws.Region),
		Endpoint: aws.String(cfg.Aws.Endpoint),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &AwsQueue{
		SqsClient: sqs.New(sess),
		QueueURL:  cfg.Aws.JobsQueueURL,
	}, nil
}

func (q *AwsQueue) QueueSetForProcessing(ctx context.Context, scannedCaseResponse *types.ScannedCaseResponse, fileName string) (MessageID *string, err error) {
	finalMessageJson, err := createMessageBody(scannedCaseResponse, fileName)
	if err != nil {
		return nil, err
	}

	// Send the message to the SQS queue
	input := &sqs.SendMessageInput{
		QueueUrl:       aws.String(q.QueueURL),
		MessageBody:    aws.String(finalMessageJson),
		MessageGroupId: aws.String(scannedCaseResponse.UID),
	}

	output, err := q.SqsClient.SendMessageWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to send message to SQS queue: %w", err)
	}

	return output.MessageId, nil
}

func createMessageBody(scannedCaseResponse *types.ScannedCaseResponse, fileName string) (string, error) {
	// Create a message structure
	phpSerializedContent := util.PhpSerialize(map[string]interface{}{
		"uid":      scannedCaseResponse.UID,
		"filename": fileName,
	})
	phpSerializedMetadata := util.PhpSerialize(map[string]interface{}{
		"__name__": "Ddc\\\\Job\\\\FormJob",
	})

	// Create the final message structure
	finalMessage := map[string]interface{}{
		"content":  phpSerializedContent,
		"metadata": phpSerializedMetadata,
	}

	return util.PhpSerialize(finalMessage), nil
}
