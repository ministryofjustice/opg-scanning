package aws

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockAwsClient struct {
	mock.Mock
}

func (m *MockAwsClient) GetSecretValue(ctx context.Context, secretName string) (string, error) {
	args := m.Called(ctx, secretName)
	return args.String(0), args.Error(1)
}

func (m *MockAwsClient) GetSsmValue(ctx context.Context, secretName string) (string, error) {
	args := m.Called(ctx, secretName)
	return args.String(0), args.Error(1)
}
