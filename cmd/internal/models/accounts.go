package models

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"userID"`
	ParticipantID uuid.UUID `json:"participantID"`
	AccountNumber string    `json:"accountNumber"`
	Branch        string    `json:"branch"`
	AccountType   string    `json:"accountType"`
	Balance       int64     `json:"balance"`
	CreatedAt     time.Time `json:"createdAt"`
}

type CreateAccountRequest struct {
	ParticipantID uuid.UUID `json:"participantID"`
	AccountNumber string    `json:"accountNumber"`
	Branch        string    `json:"branch"`
	AccountType   string    `json:"accountType"`
}

type CreateAccountResponse struct {
	ID uuid.UUID `json:"id"`
}
