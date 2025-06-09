package auth

type userLogin struct {
	User User `json:"user"`
}

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
