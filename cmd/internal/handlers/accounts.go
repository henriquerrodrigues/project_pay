package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"project_pay/cmd/internal/models"
	accountsrepo "project_pay/cmd/internal/repositories/accounts"

	"github.com/google/uuid"
)

type AccountHandler struct {
	repo *accountsrepo.Repository
}

func NewAccountHandler(repo *accountsrepo.Repository) *AccountHandler {
	return &AccountHandler{repo: repo}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "body inválido", http.StatusBadRequest)
		return
	}

	id, err := h.repo.Create(r.Context(), req)
	if err != nil {
		log.Printf("erro ao criar conta: %v", err)
		http.Error(w, "erro ao criar conta", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.CreateAccountResponse{ID: id})
}

func (h *AccountHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	account, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}
	if account == nil {
		http.Error(w, "conta não encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}
