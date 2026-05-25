package models

type AuthJWT struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
