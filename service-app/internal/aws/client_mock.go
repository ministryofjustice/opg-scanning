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

func (m *MockAwsClient) FetchCredentials(ctx context.Context) (map[string]string, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]string), args.Error(1)
}
