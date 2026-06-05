package handler

import (
	"encoding/json"
	"net/http"

	"expense-chain/src/model"
	"expense-chain/src/service"
)

type PolicyHandler struct {
	svc *service.PolicyService
}

func NewPolicyHandler(svc *service.PolicyService) *PolicyHandler {
	return &PolicyHandler{svc: svc}
}

func (h *PolicyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CompanyID         string                        `json:"company_id"`
		Name              string                        `json:"name"`
		MaxAmountPerTx    float64                       `json:"max_amount_per_tx"`
		MaxAmountPerDay   float64                       `json:"max_amount_per_day"`
		MaxAmountPerMonth float64                       `json:"max_amount_per_month"`
		BlockedCategories []model.TransactionCategory   `json:"blocked_categories"`
		AllowedCardTypes  []model.CardType              `json:"allowed_card_types"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	policy, err := h.svc.Create(
		body.CompanyID,
		body.Name,
		body.MaxAmountPerTx,
		body.MaxAmountPerDay,
		body.MaxAmountPerMonth,
		body.BlockedCategories,
		body.AllowedCardTypes,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, policy)
}

func (h *PolicyHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	policies, err := h.svc.GetAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, policies)
}

func (h *PolicyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	policy, err := h.svc.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, policy)
}

func (h *PolicyHandler) GetByCompanyID(w http.ResponseWriter, r *http.Request) {
	companyID := r.PathValue("company_id")
	policies, err := h.svc.GetByCompanyID(companyID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, policies)
}

func (h *PolicyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Name              string                        `json:"name"`
		MaxAmountPerTx    float64                       `json:"max_amount_per_tx"`
		MaxAmountPerDay   float64                       `json:"max_amount_per_day"`
		MaxAmountPerMonth float64                       `json:"max_amount_per_month"`
		BlockedCategories []model.TransactionCategory   `json:"blocked_categories"`
		AllowedCardTypes  []model.CardType              `json:"allowed_card_types"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	policy, err := h.svc.Update(
		id,
		body.Name,
		body.MaxAmountPerTx,
		body.MaxAmountPerDay,
		body.MaxAmountPerMonth,
		body.BlockedCategories,
		body.AllowedCardTypes,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, policy)
}

func (h *PolicyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
