package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	participantssvc "project_pay/cmd/internal/services/participants"
)

type ParticipantHandler struct {
	svc *participantssvc.Service
}

func NewParticipantHandler(svc *participantssvc.Service) *ParticipantHandler {
	return &ParticipantHandler{svc: svc}
}

func (h *ParticipantHandler) CreateParticipant(w http.ResponseWriter, r *http.Request) {
	var req models.CreateParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("body inválido"))
		return
	}

	id, err := h.svc.Create(r.Context(), req)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.CreateParticipantResponse{ID: id})
}

func (h *ParticipantHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("id inválido"))
		return
	}

	participant, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(participant)
}
