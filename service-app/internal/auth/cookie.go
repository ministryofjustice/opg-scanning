package auth

import (
	"net/http"
	"time"
)

type CookieHelper interface {
	GetTokenFromCookie(r *http.Request) (string, error)
	SetTokenInCookie(w http.ResponseWriter, token string, expiry time.Time) error
}

type MembraneCookieHelper struct {
	CookieName string
	Secure     bool
}

func (h MembraneCookieHelper) GetTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(h.CookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (h MembraneCookieHelper) SetTokenInCookie(w http.ResponseWriter, token string, expiry time.Time) error {
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
