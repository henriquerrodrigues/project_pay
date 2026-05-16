package models

import (
	"time"

	"github.com/google/uuid"
)

// representa a entidade banco
type Participant struct {
	ID        uuid.UUID `json:"id"`
	Ispb      string    `json:"ispb"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateParticipantRequest struct {
	Ispb string `json:"ispb"`
	Name string `json:"name"`
}
type CreateParticipantResponse struct {
	ID uuid.UUID `json:"id"`
}
