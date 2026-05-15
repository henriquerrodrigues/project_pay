package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"project_pay/cmd/internal/models"
	usersrepo "project_pay/cmd/internal/repositories/users"
)

type UserHandler struct {
	repo *usersrepo.Repository
}

func NewUserHandler(repo *usersrepo.Repository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "body inválido", http.StatusBadRequest)
		return
	}

	id, err := h.repo.Create(r.Context(), req)
	if err != nil {
		http.Error(w, "erro ao criar usuário", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.CreateUserResponse{ID: id})
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "usuário não encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
