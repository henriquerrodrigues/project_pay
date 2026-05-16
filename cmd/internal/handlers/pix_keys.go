package handlers

import (
	"encoding/json"
	"net/http"
	"project_pay/cmd/internal/models"
	pixKeyrepo "project_pay/cmd/internal/repositories/pix_keys"
)

type PixKeyHandler struct {
	repo *pixKeyrepo.Repository
}

func NewPixKeyHandler(repo *pixKeyrepo.Repository) *PixKeyHandler {
	return &PixKeyHandler{repo: repo}
}

func (h *PixKeyHandler) CreatePixKey(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePixKeysRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "body invalid", http.StatusBadRequest)
		return
	}

	id, err := h.repo.Create(r.Context(), req)
	if err != nil {
		http.Error(w, "erro ao criar chave pix", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.CreatePixKeysResponse{ID: id})
}

func (h *PixKeyHandler) GetByKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	pixKey, err := h.repo.GetByKey(r.Context(), key)
	if err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	if pixKey == nil {
		http.Error(w, "chave pix não encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pixKey)
}
