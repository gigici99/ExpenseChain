package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"expense-chain/src/model"
	"expense-chain/src/service"
)

type CardHandler struct {
	svc         *service.CardService
	employeeSvc *service.EmployeeService
}

func NewCardHandler(svc *service.CardService, employeeSvc *service.EmployeeService) *CardHandler {
	return &CardHandler{svc: svc, employeeSvc: employeeSvc}
}

// verifyCardOwnership checks that the employee owning this card belongs to the caller's company/identity.
func (h *CardHandler) verifyCardOwnership(claims *service.Claims, employeeID string) bool {
	if claims.Role == model.RoleAdmin {
		return true
	}
	if claims.Role == model.RoleEmployee {
		return employeeID == claims.EmployeeID
	}
	// COMPANY — check employee belongs to company
	emp, err := h.employeeSvc.GetByID(employeeID)
	if err != nil {
		return false
	}
	return emp.CompanyID == claims.CompanyID
}

func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	var body struct {
		EmployeeID string `json:"employee_id"`
		Type       string `json:"type"`
		Last4      string `json:"last4"`
		Holder     string `json:"holder"`
		Provider   string `json:"provider"`
		ExpiresAt  string `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !h.verifyCardOwnership(claims, body.EmployeeID) {
		writeError(w, http.StatusForbidden, "access denied")
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
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	id := r.PathValue("id")
	card, err := h.svc.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if !h.verifyCardOwnership(claims, card.EmployeeID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, card)
}

func (h *CardHandler) GetByEmployeeID(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	employeeID := r.PathValue("employee_id")

	if !h.verifyCardOwnership(claims, employeeID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	cards, err := h.svc.GetByEmployeeID(employeeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cards)
}

func (h *CardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	id := r.PathValue("id")

	card, err := h.svc.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if !h.verifyCardOwnership(claims, card.EmployeeID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
