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

func (h *TransactionHandler) Submit(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	var body struct {
		EmployeeID  string                    `json:"employee_id"`
		CardID      string                    `json:"card_id"`
		Amount      float64                   `json:"amount"`
		Currency    string                    `json:"currency"`
		Category    model.TransactionCategory `json:"category"`
		Description string                    `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// EMPLOYEE can only submit for themselves
	if claims.Role == model.RoleEmployee {
		body.EmployeeID = claims.EmployeeID
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

func (h *TransactionHandler) IncomingPayment(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	var body struct {
		EmployeeID  string                    `json:"employee_id"`
		CardID      string                    `json:"card_id"`
		Amount      float64                   `json:"amount"`
		Currency    string                    `json:"currency"`
		Category    model.TransactionCategory `json:"category"`
		Description string                    `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if claims.Role == model.RoleEmployee {
		body.EmployeeID = claims.EmployeeID
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
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	switch claims.Role {
	case model.RoleCompany:
		txs, err := h.svc.GetByCompanyID(claims.CompanyID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, txs)
	case model.RoleEmployee:
		txs, err := h.svc.GetByEmployeeID(claims.EmployeeID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, txs)
	default:
		// ADMIN
		txs, err := h.svc.GetAll()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, txs)
	}
}

func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	id := r.PathValue("id")
	tx, err := h.svc.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if claims.Role == model.RoleCompany && tx.CompanyID != claims.CompanyID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}
	if claims.Role == model.RoleEmployee && tx.EmployeeID != claims.EmployeeID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, tx)
}

func (h *TransactionHandler) GetByEmployeeID(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	employeeID := r.PathValue("employee_id")

	// EMPLOYEE can only query own transactions
	if claims.Role == model.RoleEmployee && employeeID != claims.EmployeeID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	txs, err := h.svc.GetByEmployeeID(employeeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, txs)
}
