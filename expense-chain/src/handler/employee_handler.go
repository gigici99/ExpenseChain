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

// Create makes an employee and, if credentials are supplied, a linked EMPLOYEE login user.
// COMPANY (and ADMIN) reach this route.
func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CompanyID string `json:"company_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		PolicyID  string `json:"policy_id"`
		Username  string `json:"username"` // optional: EMPLOYEE user login
		Password  string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	employee, err := h.svc.Create(body.CompanyID, body.FirstName, body.LastName, body.Email, body.PolicyID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// provision the employee's login user
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
	employees, err := h.svc.GetAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, employees)
}

func (h *EmployeeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	employee, err := h.svc.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, employee)
}

func (h *EmployeeHandler) GetByCompanyID(w http.ResponseWriter, r *http.Request) {
	companyID := r.PathValue("company_id")
	employees, err := h.svc.GetByCompanyID(companyID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, employees)
}

func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
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
	id := r.PathValue("id")
	if err := h.svc.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
