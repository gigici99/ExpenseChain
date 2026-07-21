package handler

import (
	"net/http"

	"expense-chain/src/model"
)

func NewRouter(
	auth *AuthHandler,
	mw *AuthMiddleware,
	company *CompanyHandler,
	employee *EmployeeHandler,
	card *CardHandler,
	policy *PolicyHandler,
	transaction *TransactionHandler,
	ledger *LedgerHandler,
) http.Handler {
	mux := http.NewServeMux()

	const (
		admin = model.RoleAdmin
		comp  = model.RoleCompany
		emp   = model.RoleEmployee
	)

	// --- Auth ---
	mux.HandleFunc("POST /api/auth/login", auth.Login)
	mux.HandleFunc("POST /api/auth/refresh", auth.Refresh)
	mux.HandleFunc("POST /api/auth/register", mw.Protect(auth.Register, admin))

	// --- Ledger (audit + blockchain) ---
	mux.HandleFunc("GET /api/ledger", mw.Protect(ledger.GetAll, comp))
	mux.HandleFunc("GET /api/ledger/verify", mw.Protect(ledger.Verify, comp))
	mux.HandleFunc("GET /api/ledger/entity/{entity_id}", mw.Protect(ledger.GetByEntityID, comp))

	// --- Company ---
	mux.HandleFunc("POST /api/companies", mw.Protect(company.Create, admin))
	mux.HandleFunc("GET /api/companies", mw.Protect(company.GetAll, comp))
	mux.HandleFunc("GET /api/companies/{id}", mw.Protect(company.GetByID, comp))
	mux.HandleFunc("PUT /api/companies/{id}", mw.Protect(company.Update, admin))
	mux.HandleFunc("DELETE /api/companies/{id}", mw.Protect(company.Delete, admin))

	// --- Employee ---
	mux.HandleFunc("POST /api/employees", mw.Protect(employee.Create, comp))
	mux.HandleFunc("GET /api/employees", mw.Protect(employee.GetAll, comp))
	mux.HandleFunc("GET /api/employees/{id}", mw.Protect(employee.GetByID, comp, emp))
	mux.HandleFunc("PUT /api/employees/{id}", mw.Protect(employee.Update, comp))
	mux.HandleFunc("DELETE /api/employees/{id}", mw.Protect(employee.Delete, comp))
	mux.HandleFunc("GET /api/employees/company/{company_id}", mw.Protect(employee.GetByCompanyID, comp))

	// --- Card ---
	mux.HandleFunc("POST /api/cards", mw.Protect(card.Create, comp))
	mux.HandleFunc("GET /api/cards/{id}", mw.Protect(card.GetByID, comp, emp))
	mux.HandleFunc("DELETE /api/cards/{id}", mw.Protect(card.Delete, comp))
	mux.HandleFunc("GET /api/cards/employee/{employee_id}", mw.Protect(card.GetByEmployeeID, comp, emp))

	// --- Policy ---
	mux.HandleFunc("POST /api/policies", mw.Protect(policy.Create, comp))
	mux.HandleFunc("GET /api/policies", mw.Protect(policy.GetAll, comp))
	mux.HandleFunc("GET /api/policies/{id}", mw.Protect(policy.GetByID, comp))
	mux.HandleFunc("PUT /api/policies/{id}", mw.Protect(policy.Update, comp))
	mux.HandleFunc("DELETE /api/policies/{id}", mw.Protect(policy.Delete, comp))
	mux.HandleFunc("GET /api/policies/company/{company_id}", mw.Protect(policy.GetByCompanyID, comp))

	// --- Transaction ---
	mux.HandleFunc("POST /api/transactions", mw.Protect(transaction.Submit, emp))
	mux.HandleFunc("GET /api/transactions", mw.Protect(transaction.GetAll, comp, emp))
	mux.HandleFunc("GET /api/transactions/{id}", mw.Protect(transaction.GetByID, comp, emp))
	mux.HandleFunc("GET /api/transactions/employee/{employee_id}", mw.Protect(transaction.GetByEmployeeID, comp, emp))

	// --- Payments ---
	mux.HandleFunc("POST /api/payments/incoming", mw.Protect(transaction.IncomingPayment, emp))

	return mux
}
