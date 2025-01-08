package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthenticateMiddleware(t *testing.T) {
	cfg := config.NewConfig()
	logger := logger.NewLogger(cfg)

	// Pre gen bcrypt hash for the pass
	passwordHash := util.HashPassword("test")

	// Mock AWS Client
	mockAwsClient := new(aws.MockAwsClient)

	// Mock GetSecretValue
	mockAwsClient.On("GetSsmValue", mock.Anything, "/local/local-credentials").Return(fmt.Sprintf("opg_document_and_d@publicguardian.gsi.gov.uk:%s", passwordHash), nil)
	mockAwsClient.On("GetSecretValue", mock.Anything, "local/jwt-key").Return("mysupersecrettestkeythatis128bits", nil)

	httpClient := httpclient.NewHttpClient(*cfg, *logger)
	httpClientMiddleware, _ := httpclient.NewMiddleware(httpClient, mockAwsClient)

	authMiddleware := NewMiddleware(mockAwsClient, httpClientMiddleware, cfg, logger)

	// Define the handler to test after authentication
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	})

	// Create a test server to test the middleware
	ts := httptest.NewServer(authMiddleware.Authenticate(handler))
	defer ts.Close()

	// Test valid credentials
	credentials := UserCredentials{
		User: User{
			Email:    "opg_document_and_d@publicguardian.gsi.gov.uk",
			Password: "test",
		},
	}
	body, err := json.Marshal(credentials)
	if err != nil {
		t.Fatalf("Could not marshal credentials: %v", err)
	}

	req, err := http.NewRequest("POST", ts.URL, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Send request to the test server
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("Could not send request: %v", err)
	}

	// Check response status
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check that the "membrane" cookie is set
	cookies := resp.Cookies()
	var tokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "membrane" {
			tokenCookie = cookie
			break
		}
	}

	// Assert that the token cookie exists
	assert.NotNil(t, tokenCookie)
	assert.NotEmpty(t, tokenCookie.Value)
}
