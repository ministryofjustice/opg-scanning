package auth

import (
	"context"

	"github.com/ministryofjustice/opg-scanning/internal/constants"
)

type userLogin struct {
	User User `json:"user"`
}

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Retrieves the users identity from the context
func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(constants.UserContextKey).(User)
	return user, ok
}
