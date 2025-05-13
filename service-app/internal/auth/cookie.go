package auth

import (
	"net/http"
	"time"
)

type cookieHelper interface {
	getTokenFromCookie(r *http.Request) (string, error)
	setTokenInCookie(w http.ResponseWriter, token string, expiry time.Time) error
}

type MembraneCookieHelper struct {
	CookieName string
	Secure     bool
}

func (h MembraneCookieHelper) getTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(h.CookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (h MembraneCookieHelper) setTokenInCookie(w http.ResponseWriter, token string, expiry time.Time) error {
	http.SetCookie(w, &http.Cookie{
		Name:     h.CookieName,
		Value:    token,
		Expires:  expiry,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteStrictMode,
	})
	return nil
}
