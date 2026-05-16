package handlers

import (
	"encoding/json"
	"net/http"
	"project_pay/cmd/internal/models"
	participantsrepo "project_pay/cmd/internal/repositories/participants"

	"github.com/google/uuid"
)

type ParticipantHandler struct {
	repo *participantsrepo.Repository
}

func NewParticipantHandler(repo *participantsrepo.Repository) *ParticipantHandler {
	return &ParticipantHandler{repo: repo}
}

func (h *ParticipantHandler) CreateParticipant(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePaticipantRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "body invalid", http.StatusBadRequest)
		return
	}

	id, err := h.repo.Create(r.Context(), req)
	if err != nil {
		http.Error(w, "error ao criar banco", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.CreateParticipantResponse{ID: id})
}

func (h *ParticipantHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))

	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	participant, err := h.repo.GetByID(r.Context(), id)
	if err == nil {
		http.Error(w, "conta não encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(participant)
}
