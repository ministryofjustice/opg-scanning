package auth

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateToken(t *testing.T) {
	mockAwsClient := &aws.MockAwsClient{}
	mockAwsClient.
		On("GetSecretValue", mock.Anything, "aws::my-secret-arn").
		Return("my-secret", nil)

	outBuf := bytes.NewBuffer([]byte{})
	mockLogger := &logger.Logger{SlogLogger: slog.New(slog.NewJSONHandler(outBuf, nil))}

	tg := jwtTokenGenerator{
		config: &config.Config{
			Auth: config.Auth{
				ApiUsername:   "user@host.example",
				JWTSecretARN:  "aws::my-secret-arn",
				JWTExpiration: 5,
			},
		},
		awsClient: mockAwsClient,
		logger:    mockLogger,
	}

	tokenString, expiry, err := tg.generateToken()
	assert.Nil(t, err)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(tg.signingSecret), nil
	})
	assert.Nil(t, err)

	tokenExpiry, _ := token.Claims.GetExpirationTime()
	assert.Equal(t, tokenExpiry.Time, expiry)

	assert.Equal(t, "user@host.example", token.Claims.(jwt.MapClaims)["session-data"])

	assert.Contains(t, outBuf.String(), "Generated new JWT token.")
}

func TestValidateToken(t *testing.T) {
	testCases := map[string]struct {
		claims     jwt.MapClaims
		expectedOk bool
	}{
		"ok": {
			claims: jwt.MapClaims{
				"session-data": "test",
				"iat":          time.Now().Add(-5 * time.Second).Unix(),
				"exp":          time.Now().Add(5 * time.Second).Unix(),
			},
			expectedOk: true,
		},
		"empty": {
			claims:     jwt.MapClaims{},
			expectedOk: false,
		},
		"expired": {
			claims: jwt.MapClaims{
				"session-data": "test",
				"iat":          time.Now().Add(-5 * time.Second).Unix(),
				"exp":          time.Now().Add(-3 * time.Second).Unix(),
			},
			expectedOk: false,
		},
		"not-yet-issued": {
			claims: jwt.MapClaims{
				"session-data": "test",
				"iat":          time.Now().Add(3 * time.Second).Unix(),
				"exp":          time.Now().Add(5 * time.Second).Unix(),
			},
			expectedOk: false,
		},
		"missing-session-data": {
			claims: jwt.MapClaims{
				"iat": time.Now().Add(-5 * time.Second).Unix(),
				"exp": time.Now().Add(5 * time.Second).Unix(),
			},
			expectedOk: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			mockAwsClient := &aws.MockAwsClient{}
			mockAwsClient.
				On("GetSecretValue", mock.Anything, "aws::my-secret-arn").
				Return("my-secret", nil)

			tg := jwtTokenGenerator{
				config: &config.Config{
					Auth: config.Auth{
						JWTSecretARN: "aws::my-secret-arn",
					},
				},
				awsClient: mockAwsClient,
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, tc.claims)
			tokenString, _ := token.SignedString([]byte("my-secret"))

			err := tg.validateToken(tokenString)

			if tc.expectedOk {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
