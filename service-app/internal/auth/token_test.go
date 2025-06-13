package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateToken(t *testing.T) {
	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		GetSecretValue(mock.Anything, "aws::my-secret-arn").
		Return("my-secret", nil)

	tg := tokenHelper{
		config: &config.Config{
			Auth: config.Auth{
				ApiUsername:   "user@host.example",
				JWTSecretARN:  "aws::my-secret-arn",
				JWTExpiration: 5,
			},
		},
		awsClient: secretsClient,
	}

	tokenString, expiry, err := tg.Generate()
	assert.Nil(t, err)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(tg.signingSecret), nil
	})
	assert.Nil(t, err)

	tokenExpiry, _ := token.Claims.GetExpirationTime()
	assert.Equal(t, tokenExpiry.Time, expiry)

	assert.Equal(t, "user@host.example", token.Claims.(jwt.MapClaims)["session-data"])
}

func TestValidateToken(t *testing.T) {
	testCases := map[string]struct {
		claims jwt.MapClaims
		ok     bool
	}{
		"ok": {
			claims: jwt.MapClaims{
				"session-data": "test",
				"iat":          time.Now().Add(-5 * time.Second).Unix(),
				"exp":          time.Now().Add(5 * time.Second).Unix(),
			},
			ok: true,
		},
		"empty": {
			claims: jwt.MapClaims{},
			ok:     false,
		},
		"expired": {
			claims: jwt.MapClaims{
				"session-data": "test",
				"iat":          time.Now().Add(-5 * time.Second).Unix(),
				"exp":          time.Now().Add(-3 * time.Second).Unix(),
			},
			ok: false,
		},
		"not-yet-issued": {
			claims: jwt.MapClaims{
				"session-data": "test",
				"iat":          time.Now().Add(3 * time.Second).Unix(),
				"exp":          time.Now().Add(5 * time.Second).Unix(),
			},
			ok: false,
		},
		"missing-session-data": {
			claims: jwt.MapClaims{
				"iat": time.Now().Add(-5 * time.Second).Unix(),
				"exp": time.Now().Add(5 * time.Second).Unix(),
			},
			ok: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			secretsClient := newMockSecretsClient(t)
			secretsClient.EXPECT().
				GetSecretValue(mock.Anything, "aws::my-secret-arn").
				Return("my-secret", nil)

			tg := tokenHelper{
				config: &config.Config{
					Auth: config.Auth{
						JWTSecretARN: "aws::my-secret-arn",
					},
				},
				awsClient: secretsClient,
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, tc.claims)
			tokenString, _ := token.SignedString([]byte("my-secret"))

			err := tg.Validate(tokenString)

			if tc.ok {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
