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

// CreateAccountRequest é o body do POST /accounts.
// O dono da conta vem do token (context), nunca do body.
type CreateAccountRequest struct {
	ParticipantID uuid.UUID `json:"participantID"`
	AccountNumber string    `json:"accountNumber"`
	Branch        string    `json:"branch"`
	AccountType   string    `json:"accountType"`
}

type CreateAccountResponse struct {
	ID uuid.UUID `json:"id"`
}
