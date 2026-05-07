package models

import (
	"time"

	"github.com/google/uuid"
)

type participants struct {
	Id         uuid.UUID
	Ispb       string
	Name       string
	Created_at time.Time
}
