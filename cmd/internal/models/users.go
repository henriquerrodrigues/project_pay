package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	Name          string
	Email         string
	Password_hash string
	Tax_id        string
	Created_at    time.Time
}

type CreateUserRequest struct {
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Password_hash string    `json:"passwordHash"`
	Tax_id        string    `json:"taxId"`
	Created_at    time.Time `json:"created_at"`
}

type CreateUserResponse struct {
	NewUserID uuid.UUID `json:"newUserId`
}
