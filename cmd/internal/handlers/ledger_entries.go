package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"project_pay/cmd/internal/apperrors"
	ledgersvc "project_pay/cmd/internal/services/ledger_entries"
)

type LedgerEntryHandler struct {
	svc *ledgersvc.Service
}

func NewLedgerEntryHandler(svc *ledgersvc.Service) *LedgerEntryHandler {
	return &LedgerEntryHandler{svc: svc}
}

func (h *LedgerEntryHandler) GetStatement(w http.ResponseWriter, r *http.Request) {
	accountID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperrors.WriteHTTP(w, r, apperrors.InvalidInput("id inválido"))
		return
	}

	entries, err := h.svc.GetStatement(r.Context(), accountID)
	if err != nil {
		apperrors.WriteHTTP(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
