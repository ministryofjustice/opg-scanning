package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

		// Validate the token (e.g., parse and verify signature)
		tg, ok := m.TokenGenerator.(*JWTTokenGenerator)
		if !ok {
			m.respondWithError(w, http.StatusInternalServerError, "Internal server error: Invalid token generator", nil)
			return
		}

		claims := jwt.MapClaims{}
		tokenParsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(tg.signingSecret), nil
		})
		if err != nil || !tokenParsed.Valid {
			m.respondWithError(w, http.StatusUnauthorized, "Unauthorized: Invalid token", err)
			return
		}

		// Verify token expiration
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				m.respondWithError(w, http.StatusUnauthorized, "Unauthorized: Token expired", nil)
				return
			}
		} else {
			m.respondWithError(w, http.StatusUnauthorized, "Unauthorized: Invalid token", nil)
			return
		}

		// Extract user information from claims
		sessionData, ok := claims["session-data"].(string)
		if !ok {
			m.respondWithError(w, http.StatusUnauthorized, "Unauthorized: Invalid token", nil)
			return
		}

		// Set the user in the context
		ctx := context.WithValue(r.Context(), userContextKey, sessionData)

		// Proceed to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	m.logger.Error("%s: %v", nil, message, err)
	http.Error(w, message, statusCode)
}
