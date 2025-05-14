package httpclient

import (
	"context"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

type httpClientInterface interface {
	HTTPRequest(ctx context.Context, url, method string, payload []byte, headers map[string]string) ([]byte, error)
	GetConfig() *config.Config
	GetLogger() *logger.Logger
}

func (r *HttpClient) GetConfig() *config.Config {
	return r.config
}

func (r *HttpClient) GetLogger() *logger.Logger {
	return r.logger
}
