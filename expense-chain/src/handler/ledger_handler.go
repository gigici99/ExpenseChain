package handler

import (
	"net/http"

	"expense-chain/src/blockchain"
)

type LedgerHandler struct {
	ledger *blockchain.Ledger
}

func NewLedgerHandler(ledger *blockchain.Ledger) *LedgerHandler {
	return &LedgerHandler{ledger: ledger}
}

// GetAll returns the full audit ledger ordered by sequence.
func (h *LedgerHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	entries, err := h.ledger.GetAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

// GetByEntityID returns the audit history of a single entity.
func (h *LedgerHandler) GetByEntityID(w http.ResponseWriter, r *http.Request) {
	entityID := r.PathValue("entity_id")
	entries, err := h.ledger.GetByEntityID(entityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

// Verify recomputes every hash and checks chain linkage — tamper detection.
func (h *LedgerHandler) Verify(w http.ResponseWriter, r *http.Request) {
	result, err := h.ledger.Verify()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}
