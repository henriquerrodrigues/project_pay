package models

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"userId"`
	ParticipantID uuid.UUID `json:"participantId"`
	AccountNumber string    `json:"accountNumber"`
	Branch        string    `json:"branch"`
	AccountType   string    `json:"accountType"`
	Balance       int64     `json:"balance"`
	CreatedAt     time.Time `json:"createdAt"`
}

type CreateAccountRequest struct {
	UserID        uuid.UUID `json:"userId"`
	ParticipantID uuid.UUID `json:"participantId"`
	AccountNumber string    `json:"accountNumber"`
	Branch        string    `json:"branch"`
	AccountType   string    `json:"accountType"`
}

type CreateAccountResponse struct {
	ID uuid.UUID `json:"id"`
}
