package models

import (
	"time"

	"github.com/google/uuid"
)

type LedgerEntry struct {
	ID           uuid.UUID `json:"id"`
	TransferID   uuid.UUID `json:"transfer_id"`
	AccountID    uuid.UUID `json:"account_id"`
	Amount       int64     `json:"amount"`
	EntryType    string    `json:"entry_type"`
	BalanceAfter int64     `json:"balance_after"`
	CreatedAt    time.Time `json:"created_at"`
}
