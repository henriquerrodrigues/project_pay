package models

import (
	"time"

	"github.com/google/uuid"
)

type Transfer struct {
	ID                uuid.UUID  `json:"id"`
	EndToendID        string     `json:"endtoendID"`
	SenderAccountID   uuid.UUID  `json:"senderAccountID"`
	ReceiverAccountID uuid.UUID  `json:"receiverAccountID"`
	Amount            int64      `json:"amount"`
	Status            string     `json:"status"`
	Description       string     `json:"description"`
	CreatedAt         time.Time  `json:"createdAt"`
	CompletedAt       *time.Time `json:"completedAt"`
}

type CreateTransferRequest struct {
	SenderAccountID   uuid.UUID `json:"senderAccountID"`
	ReceiverAccountID uuid.UUID `json:"receiverAccountID"`
	Amount            int64     `json:"amount"`
	Description       string    `json:"description"`
}

type CreateTransferResponse struct {
	ID uuid.UUID `json:"id"`
}
