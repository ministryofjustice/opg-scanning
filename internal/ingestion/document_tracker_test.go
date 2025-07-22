package ingestion

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func withDocumentTracker(t *testing.T, fn func(tracker *DocumentTracker)) {
	if testing.Short() {
		t.Skip()
		return
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-west-1"),
		config.WithBaseEndpoint("http://localhost:4566"),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "fakeKeyId",
				SecretAccessKey: "fakeAccessKey",
			}, nil
		})),
	)
	if err != nil {
		t.Logf("unable to load SDK config: %s", err)
		t.Fail()
		return
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)

	tracker := NewDocumentTracker(dynamoClient, "test")
	_, err = dynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("test"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       types.KeyTypeHash,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
	})
	if !assert.Nil(t, err) {
		t.Log(err.Error())
		return
	}

	defer dynamoClient.DeleteTable(ctx, &dynamodb.DeleteTableInput{TableName: aws.String("test")}) //nolint:errcheck
	fn(tracker)
}

func TestIntegrationDocumentTracker_ProvidesCaseNoWhenCompleted(t *testing.T) {
	withDocumentTracker(t, func(tracker *DocumentTracker) {
		err := tracker.SetProcessing(ctx, "my-id", "my-caseno")
		assert.Nil(t, err)

		err = tracker.SetCompleted(ctx, "my-id")
		assert.Nil(t, err)

		err = tracker.SetProcessing(ctx, "my-id", "my-other-caseno")
		var v AlreadyProcessedError
		if assert.ErrorAs(t, err, &v) {
			assert.Equal(t, "my-caseno", v.CaseNo)
		}
	})
}

func TestIntegrationDocumentTracker_ProvidesCaseNoWhenFailed(t *testing.T) {
	withDocumentTracker(t, func(tracker *DocumentTracker) {
		err := tracker.SetProcessing(ctx, "my-id", "my-caseno")
		assert.Nil(t, err)

		err = tracker.SetFailed(ctx, "my-id")
		assert.Nil(t, err)

		err = tracker.SetProcessing(ctx, "my-id", "my-other-caseno")
		var v AlreadyProcessedError
		if assert.ErrorAs(t, err, &v) {
			assert.Equal(t, "my-caseno", v.CaseNo)
		}
	})
}

func TestIntegrationDocumentTracker_ErrorsWhenCurrentlyProcessing(t *testing.T) {
	withDocumentTracker(t, func(tracker *DocumentTracker) {
		err := tracker.SetProcessing(ctx, "my-id", "my-caseno")
		assert.Nil(t, err)

		err = tracker.SetProcessing(ctx, "my-id", "my-other-caseno")
		var v *types.ConditionalCheckFailedException
		assert.ErrorAs(t, err, &v)
	})
}
