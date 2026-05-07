package models

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	uuid            uuid.UUID
	user_id         uuid.UUID
	participants_id uuid.UUID
	account_number  string
	branch          string
	account_type    string
	balance         int64
	created_at      time.Time
}
