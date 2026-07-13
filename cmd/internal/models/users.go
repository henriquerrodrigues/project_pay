package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	TaxID     string    `json:"tax_id"`
	CreatedAt time.Time `json:"created_at"`
}

// request body do POST /users
type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	TaxID    string `json:"tax_id"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	ID uuid.UUID `json:"id"`
}
