package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	ledgerrepo "project_pay/cmd/internal/repositories/ledger_entries"
)

type LedgerEntryHandler struct {
	repo *ledgerrepo.Repository
}

func NewLedgerEntryHandler(repo *ledgerrepo.Repository) *LedgerEntryHandler {
	return &LedgerEntryHandler{repo: repo}
}

func (h *LedgerEntryHandler) GetStatement(w http.ResponseWriter, r *http.Request) {
	accountID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	entries, err := h.repo.GetByAccountID(r.Context(), accountID)
	if err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
