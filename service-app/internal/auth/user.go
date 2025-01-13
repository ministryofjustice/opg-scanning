package auth

import "context"

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type contextKey string

const userContextKey = contextKey("auth-user")

// Retrieves the users identity from the context
func UserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(userContextKey).(string)
	return user, ok
}
