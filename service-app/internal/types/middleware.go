package types

type User struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type AuthRequest struct {
	User User `json:"user"`
}

type AuthResponse struct {
	Email               string `json:"email"`
	UserId              int    `json:"user_id"`
	AuthenticationToken string `json:"authentication_token"`
}
