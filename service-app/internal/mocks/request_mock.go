package mocks

import (
	"context"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/stretchr/testify/mock"
)

type MockHttpClient struct {
	mock.Mock
}

func (m *MockHttpClient) HTTPRequest(ctx context.Context, url, method string, payload []byte, headers map[string]string) ([]byte, error) {
	args := m.Called(ctx, url, method, payload, headers)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockHttpClient) GetConfig() *config.Config {
	args := m.Called()
	return args.Get(0).(*config.Config)
}

func (m *MockHttpClient) GetLogger() *logger.Logger {
	args := m.Called()
	return args.Get(0).(*logger.Logger)
}
