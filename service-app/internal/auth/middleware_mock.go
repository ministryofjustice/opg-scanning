package auth

import (
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/mocks"
	"github.com/stretchr/testify/mock"
)

func PrepareMocks(mockConfig *config.Config, logger *logger.Logger) (*mocks.MockHttpClient, *Middleware, *aws.MockAwsClient, *JWTTokenGenerator) {
	// Initialize the mock AWS client
	mockAwsClient := new(aws.MockAwsClient)
	mockAwsClient.On("GetSecretValue", mock.Anything, "local/jwt-key").Maybe().Return("mysupersecrettestkeythatis128bits", nil)
	mockAwsClient.On("FetchCredentials", mock.Anything).Maybe().Return(map[string]string{
		mockConfig.Auth.ApiUsername: "test",
	}, nil)
	// Create the HTTP client and middleware
	mockHttpClient := new(mocks.MockHttpClient)
	mockHttpClient.On("GetConfig").Return(mockConfig)
	mockHttpClient.On("GetLogger").Return(logger)

	tokenGenerator := NewJWTTokenGenerator(mockAwsClient, mockConfig, logger)
	cookieHelper := MembraneCookieHelper{
		CookieName: "membrane",
		Secure:     false,
	}
	authenticator := NewBasicAuthAuthenticator(mockAwsClient, cookieHelper, tokenGenerator)
	authMiddleware := NewMiddleware(authenticator, tokenGenerator, cookieHelper, logger)

	return mockHttpClient, authMiddleware, mockAwsClient, tokenGenerator
}
