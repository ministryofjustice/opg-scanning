package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/stretchr/testify/assert"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func TestCheckAuthMiddleware(t *testing.T) {
	mockConfig := config.NewConfig()
	logger := logger.GetLogger(mockConfig)
	_, middleware, _, _ := PrepareMocks(mockConfig, logger)

	w := httptest.NewRecorder()
	handler := middleware.CheckAuthMiddleware(http.HandlerFunc(testHandler))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"session-data": "test",
		"iat":          time.Now().Add(-5 * time.Second).Unix(),
		"exp":          time.Now().Add(5 * time.Second).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("mysupersecrettestkeythatis128bits"))

	handler.ServeHTTP(w, &http.Request{
		Header: map[string][]string{
			"Cookie": {
				fmt.Sprintf("membrane=%s", tokenString),
			},
		},
	})

	assert.Equal(t, "ok", w.Body.String())
}
