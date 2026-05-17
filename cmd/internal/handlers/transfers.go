package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"project_pay/cmd/internal/models"
	transfersrepo "project_pay/cmd/internal/repositories/transfers"

	"github.com/google/uuid"
)

type TransferHandler struct {
	repo *transfersrepo.Repository
}

func NewTransferHandler(repo *transfersrepo.Repository) *TransferHandler {
	return &TransferHandler{repo: repo}
}

func (h *TransferHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransferRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "body inválido", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "valor deve ser maior que zero", http.StatusBadRequest)
		return
	}

	if req.ReceiverAccountID == req.SenderAccountID {
		http.Error(w, "remetente e recebedor não podem ser iguais", http.StatusBadRequest)
		return
	}

	transfer, err := h.repo.Create(r.Context(), req)
	if errors.Is(err, transfersrepo.ErrInsufficientFunds) {
		http.Error(w, "saldo insuficiente", http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transfer)
}

func (h *TransferHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))

	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	transfer, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}
	if transfer == nil {
		http.Error(w, "transferência não encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transfer)
}
