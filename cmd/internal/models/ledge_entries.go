package models

import (
	"time"

	"github.com/google/uuid"
)

type LedgeEntries struct {
	id            uuid.UUID
	transfer_id   uuid.UUID
	account_id    uuid.UUID
	amount        int64
	entry_type    string
	balance_after int64
	created_at    time.Time
}
