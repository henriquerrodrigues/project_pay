package handlers

import (
	"encoding/json"
	"net/http"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	pixkeyssvc "project_pay/cmd/internal/services/pix_keys"
)

type PixKeyHandler struct {
	svc *pixkeyssvc.Service
}

func NewPixKeyHandler(svc *pixkeyssvc.Service) *PixKeyHandler {
	return &PixKeyHandler{svc: svc}
}

func (h *PixKeyHandler) CreatePixKey(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePixKeysRequest
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
	json.NewEncoder(w).Encode(models.CreatePixKeysResponse{ID: id})
}

func (h *PixKeyHandler) GetByKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	pixKey, err := h.svc.GetByKey(r.Context(), key)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pixKey)
}
