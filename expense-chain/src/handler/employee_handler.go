package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"expense-chain/src/model"
	"expense-chain/src/service"
)

type EmployeeHandler struct {
	svc  *service.EmployeeService
	auth *service.AuthService
}

func NewEmployeeHandler(svc *service.EmployeeService, auth *service.AuthService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc, auth: auth}
}

func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	var body struct {
		CompanyID string `json:"company_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		PolicyID  string `json:"policy_id"`
		Username  string `json:"username"`
		Password  string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// COMPANY users can only create employees for their own company
	if claims.Role == model.RoleCompany {
		body.CompanyID = claims.CompanyID
	}

	employee, err := h.svc.Create(body.CompanyID, body.FirstName, body.LastName, body.Email, body.PolicyID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if body.Username != "" && body.Password != "" {
		if _, err := h.auth.Register(body.Username, body.Password, model.RoleEmployee, employee.CompanyID, employee.ID); err != nil {
			log.Printf("[EmployeeHandler] employee %s created but user provisioning failed: %v", employee.ID, err)
			writeError(w, http.StatusBadRequest, "dipendente creato, ma utente non creato: "+err.Error())
			return
		}
	}

	writeJSON(w, http.StatusCreated, employee)
}

func (h *EmployeeHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	if claims.Role == model.RoleCompany {
		employees, err := h.svc.GetByCompanyID(claims.CompanyID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, employees)
		return
	}

	// ADMIN sees all
	employees, err := h.svc.GetAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, employees)
}

func (h *EmployeeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	id := r.PathValue("id")
	employee, err := h.svc.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if claims.Role == model.RoleCompany && employee.CompanyID != claims.CompanyID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, employee)
}

func (h *EmployeeHandler) GetByCompanyID(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	companyID := r.PathValue("company_id")

	// COMPANY can only query own employees
	if claims.Role == model.RoleCompany && companyID != claims.CompanyID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	employees, err := h.svc.GetByCompanyID(companyID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, employees)
}

func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	id := r.PathValue("id")

	// ownership check for COMPANY
	if claims.Role == model.RoleCompany {
		existing, err := h.svc.GetByID(id)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if existing.CompanyID != claims.CompanyID {
			writeError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	var body struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		PolicyID  string `json:"policy_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	employee, err := h.svc.Update(id, body.FirstName, body.LastName, body.Email, body.PolicyID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, employee)
}

func (h *EmployeeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	id := r.PathValue("id")

	if claims.Role == model.RoleCompany {
		existing, err := h.svc.GetByID(id)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if existing.CompanyID != claims.CompanyID {
			writeError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	if err := h.svc.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
