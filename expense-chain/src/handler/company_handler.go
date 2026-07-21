package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"expense-chain/src/model"
	"expense-chain/src/service"
)

type CompanyHandler struct {
	svc  *service.CompanyService
	auth *service.AuthService
}

func NewCompanyHandler(svc *service.CompanyService, auth *service.AuthService) *CompanyHandler {
	return &CompanyHandler{svc: svc, auth: auth}
}

func (h *CompanyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		VatID    string `json:"vat_id"`
		Address  string `json:"address"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	company, err := h.svc.Create(body.Name, body.VatID, body.Address)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if body.Username != "" && body.Password != "" {
		if _, err := h.auth.Register(body.Username, body.Password, model.RoleCompany, company.ID, ""); err != nil {
			log.Printf("[CompanyHandler] company %s created but user provisioning failed: %v", company.ID, err)
			writeError(w, http.StatusBadRequest, "azienda creata, ma utente non creato: "+err.Error())
			return
		}
	}

	writeJSON(w, http.StatusCreated, company)
}

func (h *CompanyHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	// COMPANY user sees only own company
	if claims.Role == model.RoleCompany {
		company, err := h.svc.GetByID(claims.CompanyID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, []model.Company{*company})
		return
	}

	// ADMIN sees all
	companies, err := h.svc.GetAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, companies)
}

func (h *CompanyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFrom(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing claims")
		return
	}

	id := r.PathValue("id")

	if claims.Role == model.RoleCompany && id != claims.CompanyID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	company, err := h.svc.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, company)
}

func (h *CompanyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Name    string `json:"name"`
		VatID   string `json:"vat_id"`
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	company, err := h.svc.Update(id, body.Name, body.VatID, body.Address)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, company)
}

func (h *CompanyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
