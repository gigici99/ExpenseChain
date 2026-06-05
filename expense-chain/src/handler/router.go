package handler

import "net/http"

// NewRouter wires all handlers to routes using Go 1.22+ ServeMux with method+path patterns.
func NewRouter(
	company *CompanyHandler,
	employee *EmployeeHandler,
	card *CardHandler,
	policy *PolicyHandler,
	transaction *TransactionHandler,
) http.Handler {
	mux := http.NewServeMux()

	// --- Company ---
	mux.HandleFunc("POST /api/companies", company.Create)
	mux.HandleFunc("GET /api/companies", company.GetAll)
	mux.HandleFunc("GET /api/companies/{id}", company.GetByID)
	mux.HandleFunc("PUT /api/companies/{id}", company.Update)
	mux.HandleFunc("DELETE /api/companies/{id}", company.Delete)

	// --- Employee ---
	mux.HandleFunc("POST /api/employees", employee.Create)
	mux.HandleFunc("GET /api/employees", employee.GetAll)
	mux.HandleFunc("GET /api/employees/{id}", employee.GetByID)
	mux.HandleFunc("PUT /api/employees/{id}", employee.Update)
	mux.HandleFunc("DELETE /api/employees/{id}", employee.Delete)
	mux.HandleFunc("GET /api/employees/company/{company_id}", employee.GetByCompanyID)

	// --- Card ---
	mux.HandleFunc("POST /api/cards", card.Create)
	mux.HandleFunc("GET /api/cards/{id}", card.GetByID)
	mux.HandleFunc("DELETE /api/cards/{id}", card.Delete)
	mux.HandleFunc("GET /api/cards/employee/{employee_id}", card.GetByEmployeeID)

	// --- Policy ---
	mux.HandleFunc("POST /api/policies", policy.Create)
	mux.HandleFunc("GET /api/policies", policy.GetAll)
	mux.HandleFunc("GET /api/policies/{id}", policy.GetByID)
	mux.HandleFunc("PUT /api/policies/{id}", policy.Update)
	mux.HandleFunc("DELETE /api/policies/{id}", policy.Delete)
	mux.HandleFunc("GET /api/policies/company/{company_id}", policy.GetByCompanyID)

	// --- Transaction — flusso 1: manuale ---
	mux.HandleFunc("POST /api/transactions", transaction.Submit)
	mux.HandleFunc("GET /api/transactions", transaction.GetAll)
	mux.HandleFunc("GET /api/transactions/{id}", transaction.GetByID)
	mux.HandleFunc("GET /api/transactions/employee/{employee_id}", transaction.GetByEmployeeID)

	// --- Transaction — flusso 2: evento pagamento simulato ---
	mux.HandleFunc("POST /api/payments/incoming", transaction.IncomingPayment)

	return mux
}
