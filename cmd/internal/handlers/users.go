package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	userssvc "project_pay/cmd/internal/services/users"
)

type UserHandler struct {
	svc *userssvc.Service
}

func NewUserHandler(svc *userssvc.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
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
	json.NewEncoder(w).Encode(models.CreateUserResponse{ID: id})
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("id inválido"))
		return
	}

	user, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
