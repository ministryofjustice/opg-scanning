package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
)

type Middleware struct {
	Authenticator  Authenticator
	TokenGenerator TokenGenerator
	CookieHelper   CookieHelper
	logger         *logger.Logger
}

func NewMiddleware(authenticator Authenticator, tokenGenerator TokenGenerator, cookieHelper CookieHelper, logger *logger.Logger) *Middleware {
	return &Middleware{
		Authenticator:  authenticator,
		TokenGenerator: tokenGenerator,
		CookieHelper:   cookieHelper,
		logger:         logger,
	}
}

func (m *Middleware) AuthenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, err := m.Authenticator.Authenticate(w, r)
		if err != nil {
			m.respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
			return
		}

		// Pass the new context with user info to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) CheckAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := m.CookieHelper.GetTokenFromCookie(r)
		if err != nil {
			m.respondWithError(w, http.StatusUnauthorized, "Unauthorized: Missing token", err)
			return
		}

		ctx := context.WithValue(r.Context(), constants.UserContextKey, token)

		// Get existing  X-Amzn-Trace-Id header
		traceID := r.Header.Get(constants.XAmznTraceIDHeader)
		if traceID == "" {
			m.logger.Info("Missing X-Amzn-Trace-Id header", nil)
		}

		ctx = context.WithValue(ctx, constants.TraceIDKey, traceID)
		m.logger.Info(fmt.Sprintf("Incoming request: %s %s (TraceID: %s)", r.Method, r.URL.Path, traceID), nil)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	m.logger.Error("%s: %v", nil, message, err)
	http.Error(w, message, statusCode)
}
