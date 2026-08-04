package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	"project_pay/cmd/internal/models"
	transferssvc "project_pay/cmd/internal/services/transfers"
)

type TransferHandler struct {
	svc *transferssvc.Service
}

func NewTransferHandler(svc *transferssvc.Service) *TransferHandler {
	return &TransferHandler{svc: svc}
}

func (h *TransferHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("body inválido"))
		return
	}

	transfer, err := h.svc.Create(r.Context(), req)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transfer)
}

func (h *TransferHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("id inválido"))
		return
	}

	transfer, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transfer)
}
