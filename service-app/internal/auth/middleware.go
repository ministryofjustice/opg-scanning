package auth

import (
	"context"
	"net/http"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

type Middleware struct {
	Authenticator  authenticator
	TokenGenerator tokenGenerator
	CookieHelper   cookieHelper
	logger         *logger.Logger
}

func NewMiddleware(authenticator authenticator, tokenGenerator tokenGenerator, cookieHelper cookieHelper, logger *logger.Logger) *Middleware {
	return &Middleware{
		Authenticator:  authenticator,
		TokenGenerator: tokenGenerator,
		CookieHelper:   cookieHelper,
		logger:         logger,
	}
}

func (m *Middleware) CheckAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := m.CookieHelper.getTokenFromCookie(r)
		if err != nil {
			m.respondWithError(w, http.StatusUnauthorized, "Unauthorized: Missing token", err)
			return
		}

		err = m.TokenGenerator.validateToken(token)
		if err != nil {
			m.respondWithError(w, http.StatusUnauthorized, "Unauthorized: Invalid token", err)
			return
		}

		ctx := context.WithValue(r.Context(), constants.UserContextKey, token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	m.logger.Error("%s: %v", nil, message, err)
	http.Error(w, message, statusCode)
}
