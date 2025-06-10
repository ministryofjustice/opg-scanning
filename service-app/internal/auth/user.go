package auth

type login struct {
	User loginUser `json:"user"`
}

type loginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthenticatedUser struct {
	Email string `json:"email"`
	Token string `json:"authentication_token"`
}
