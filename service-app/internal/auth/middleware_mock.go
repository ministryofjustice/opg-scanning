package auth

import (
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/mocks"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func PrepareMocks(mockConfig *config.Config, logger *logger.Logger) (*mocks.MockHttpClient, *Middleware, *aws.MockAwsClient, *jwtTokenGenerator) {
	// Initialize the mock AWS client
	mockAwsClient := new(aws.MockAwsClient)
	mockAwsClient.On("GetSecretValue", mock.Anything, "local/jwt-key").Maybe().Return("mysupersecrettestkeythatis128bits", nil)
	mockAwsClient.On("FetchCredentials", mock.Anything).Maybe().Return(map[string]string{
		mockConfig.Auth.ApiUsername: hashPassword("password"),
	}, nil)

	mockAwsClient.
		On("PersistSetData", mock.Anything, mock.Anything).
		Return("path/my-set.xml", nil).
		Maybe()
	mockAwsClient.
		On("PersistFormData", mock.Anything, mock.Anything, mock.Anything).
		Return("testFileName", nil).
		Maybe()
	mockAwsClient.
		On("QueueSetForProcessing", mock.Anything, mock.Anything, mock.Anything).
		Return("123", nil).
		Maybe()

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

func hashPassword(password string) string {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic("failed to hash password in mock: " + err.Error())
	}
	return string(hashedBytes)
}
