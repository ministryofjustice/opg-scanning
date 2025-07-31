package ingestion

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	statusProcessing = "PROCESSING"
	statusCompleted  = "COMPLETED"
	statusFailed     = "FAILED"
)

type AlreadyProcessedError struct {
	CaseNo string
}

func (e AlreadyProcessedError) Error() string {
	return "document already processed"
}

type DocumentTracker struct {
	dynamo    *dynamodb.Client
	tableName string
}

func NewDocumentTracker(dynamo *dynamodb.Client, tableName string) *DocumentTracker {
	return &DocumentTracker{dynamo: dynamo, tableName: tableName}
}

func (s *DocumentTracker) SetProcessing(ctx context.Context, id, caseNo string) error {
	if id == "" {
		return nil
	}

	_, err := s.dynamo.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item: map[string]types.AttributeValue{
			"PK":     &types.AttributeValueMemberS{Value: "DOCUMENT#" + id},
			"SK":     &types.AttributeValueMemberS{Value: "DOCUMENT#" + id},
			"CaseNo": &types.AttributeValueMemberS{Value: caseNo},
			"Status": &types.AttributeValueMemberS{Value: statusProcessing},
		},
		ReturnValuesOnConditionCheckFailure: types.ReturnValuesOnConditionCheckFailureAllOld,
		ConditionExpression:                 aws.String("attribute_not_exists(PK) OR #Status = :Failed"),
		ExpressionAttributeNames: map[string]string{"#Status": "Status"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":Failed": &types.AttributeValueMemberS{Value: statusFailed},
		},
	})

	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			var v struct {
				CaseNo string
				Status string
			}
			if err := attributevalue.UnmarshalMap(condErr.Item, &v); err != nil {
				return err
			}

			if v.Status == statusCompleted || v.Status == statusFailed {
				return AlreadyProcessedError{CaseNo: v.CaseNo}
			}

			return condErr
		}

		return err
	}

	return nil
}

func (s *DocumentTracker) SetCompleted(ctx context.Context, id string) error {
	return s.setStatus(ctx, id, statusCompleted)
}

func (s *DocumentTracker) SetFailed(ctx context.Context, id string) error {
	return s.setStatus(ctx, id, statusFailed)
}

func (s *DocumentTracker) setStatus(ctx context.Context, id, status string) error {
	if id == "" {
		return nil
	}

	_, err := s.dynamo.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "DOCUMENT#" + id},
			"SK": &types.AttributeValueMemberS{Value: "DOCUMENT#" + id},
		},
		ExpressionAttributeNames: map[string]string{"#Status": "Status"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":Processing": &types.AttributeValueMemberS{Value: statusProcessing},
			":NewStatus":  &types.AttributeValueMemberS{Value: status},
		},
		ConditionExpression: aws.String("#Status = :Processing"),
		UpdateExpression:    aws.String("SET #Status = :NewStatus"),
	})

	return err
}
