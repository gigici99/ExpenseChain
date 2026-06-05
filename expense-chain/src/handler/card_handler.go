package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"expense-chain/src/model"
	"expense-chain/src/service"
)

type CardHandler struct {
	svc *service.CardService
}

func NewCardHandler(svc *service.CardService) *CardHandler {
	return &CardHandler{svc: svc}
}

func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EmployeeID string `json:"employee_id"`
		Type       string `json:"type"`
		Last4      string `json:"last4"`
		Holder     string `json:"holder"`
		Provider   string `json:"provider"`
		ExpiresAt  string `json:"expires_at"` // RFC3339
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	expiresAt, err := time.Parse(time.RFC3339, body.ExpiresAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "expires_at must be RFC3339 format")
		return
	}

	card, err := h.svc.Create(
		body.EmployeeID,
		model.CardType(body.Type),
		body.Last4,
		body.Holder,
		body.Provider,
		expiresAt,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, card)
}

func (h *CardHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	card, err := h.svc.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, card)
}

func (h *CardHandler) GetByEmployeeID(w http.ResponseWriter, r *http.Request) {
	employeeID := r.PathValue("employee_id")
	cards, err := h.svc.GetByEmployeeID(employeeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cards)
}

func (h *CardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
