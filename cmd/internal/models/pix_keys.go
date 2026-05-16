package models

import (
	"time"

	"github.com/google/uuid"
)

// PixKeys representa a chave pix
type PixKey struct {
	ID        uuid.UUID `json:"id"`
	KeyType   string    `json:"keyType"`
	KeyValue  string    `json:"keyValue"`
	AccountID uuid.UUID `json:"accountID"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreatePixKeysRequest struct {
	KeyType   string    `json:"keyType"`
	KeyValue  string    `json:"keyValue"`
	AccountID uuid.UUID `json:"accountID"`
}

type CreatePixKeysResponse struct {
	ID uuid.UUID `json:"id"`
}
