package handler

import (
	"encoding/json"
	"net/http"

	"expense-chain/src/model"
	"expense-chain/src/service"
)

type TransactionHandler struct {
	svc *service.TransactionService
}

func NewTransactionHandler(svc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

// Submit handles flusso 1 — dipendente inserisce manualmente la spesa
func (h *TransactionHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EmployeeID  string                      `json:"employee_id"`
		CardID      string                      `json:"card_id"`
		Amount      float64                     `json:"amount"`
		Currency    string                      `json:"currency"`
		Category    model.TransactionCategory   `json:"category"`
		Description string                      `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.svc.Submit(
		body.EmployeeID,
		body.CardID,
		body.Amount,
		body.Currency,
		body.Category,
		body.Description,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, tx)
}

// IncomingPayment handles flusso 2 — evento simulato pagamento carta/Google Pay
func (h *TransactionHandler) IncomingPayment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EmployeeID  string                      `json:"employee_id"`
		CardID      string                      `json:"card_id"`
		Amount      float64                     `json:"amount"`
		Currency    string                      `json:"currency"`
		Category    model.TransactionCategory   `json:"category"`
		Description string                      `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.svc.Submit(
		body.EmployeeID,
		body.CardID,
		body.Amount,
		body.Currency,
		body.Category,
		body.Description,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	txs, err := h.svc.GetAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, txs)
}

func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tx, err := h.svc.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

func (h *TransactionHandler) GetByEmployeeID(w http.ResponseWriter, r *http.Request) {
	employeeID := r.PathValue("employee_id")
	txs, err := h.svc.GetByEmployeeID(employeeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, txs)
}
