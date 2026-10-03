package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"project_pay/cmd/internal/models"
	riskrepo "project_pay/cmd/internal/repositories/risk"
	transfersrepo "project_pay/cmd/internal/repositories/transfers"
	"project_pay/cmd/internal/risk"

	"github.com/google/uuid"
)

type TransferHandler struct {
	repo *transfersrepo.Repository
	risk RiskChecker
}

// RiskChecker carrega os sinais de risco e grava a trilha de auditoria.
type RiskChecker interface {
	LoadSignals(ctx context.Context, req models.CreateTransferRequest, now time.Time) (risk.Signals, error)
	Save(ctx context.Context, req models.CreateTransferRequest, transferID *uuid.UUID, a risk.Assessment) error
}

func NewTransferHandler(repo *transfersrepo.Repository) *TransferHandler {
	return &TransferHandler{repo: repo}
}

// WithRiskCheck liga a análise de risco antes de cada transferência.
// Sem ela o handler se comporta como antes.
func (h *TransferHandler) WithRiskCheck(rc RiskChecker) *TransferHandler {
	h.risk = rc
	return h
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

	var assessment *risk.Assessment
	if h.risk != nil {
		signals, err := h.risk.LoadSignals(r.Context(), req, time.Now())
		if errors.Is(err, riskrepo.ErrAccountNotFound) {
			http.Error(w, "conta não encontrada", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "erro interno", http.StatusInternalServerError)
			return
		}

		a := risk.Evaluate(risk.DefaultConfig(), signals)
		assessment = &a
		if a.Decision != risk.DecisionApprove {
			h.saveAssessment(r.Context(), req, nil, a)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(a)
			return
		}
	}

	transfer, err := h.repo.Create(r.Context(), req)
	if assessment != nil {
		var transferID *uuid.UUID
		if err == nil {
			transferID = &transfer.ID
		}
		h.saveAssessment(r.Context(), req, transferID, *assessment)
	}
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

// saveAssessment grava a decisão sem derrubar a requisição: a transferência
// já foi decidida (ou executada) quando chegamos aqui.
func (h *TransferHandler) saveAssessment(ctx context.Context, req models.CreateTransferRequest, transferID *uuid.UUID, a risk.Assessment) {
	if err := h.risk.Save(ctx, req, transferID, a); err != nil {
		log.Printf("erro ao gravar decisão de risco: %v", err)
	}
}
