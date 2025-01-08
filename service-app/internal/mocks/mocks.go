package mocks

import (
	"fmt"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/util"

	"github.com/stretchr/testify/mock"
)

// Initializes the middleware and mock dependencies
func PrepareMocks(mockConfig *config.Config, logger *logger.Logger) (*httpclient.MockHttpClient, *httpclient.Middleware, *aws.MockAwsClient) {
	// Hash the password as required
	passwordHash := util.HashPassword("test")

	// Initialize the mock AWS client
	mockAwsClient := new(aws.MockAwsClient)
	mockAwsClient.On("GetSsmValue", mock.Anything, "/local/local-credentials").Return(fmt.Sprintf("opg_document_and_d@publicguardian.gsi.gov.uk:%s", passwordHash), nil)
	mockAwsClient.On("GetSecretValue", mock.Anything, "local/jwt-key").Return("mysupersecrettestkeythatis128bits", nil)

	// Create the HTTP client and middleware
	httpClient := new(httpclient.MockHttpClient)
	httpClient.On("GetConfig").Return(mockConfig)
	httpClient.On("GetLogger").Return(logger)

	httpClientMiddleware, err := httpclient.NewMiddleware(httpClient, mockAwsClient)
	if err != nil {
		panic(err)
	}

	return httpClient, httpClientMiddleware, mockAwsClient
}
