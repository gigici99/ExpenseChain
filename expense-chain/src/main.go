package main

import (
	"log"
	"net/http"
	"os"

	"expense-chain/src/blockchain"
	"expense-chain/src/contract"
	"expense-chain/src/db"
	"expense-chain/src/handler"
	"expense-chain/src/model"
	"expense-chain/src/repository"
	"expense-chain/src/service"
)

func main() {
	// --- DB ---
	database := db.MustConnect("expense.db")
	defer db.Close(database)

	// AutoMigrate — creates tables from structs if not exist
	if err := database.AutoMigrate(
		&model.Company{},
		&model.Employee{},
		&model.Card{},
		&model.Policy{},
		&model.Transaction{},
		&model.LedgerEntry{},
		&model.User{},
	); err != nil {
		log.Fatalf("[main] AutoMigrate failed: %v", err)
	}

	// --- JWT secret ---
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-insecure-secret-change-me" // dev fallback only
		log.Println("[main] WARNING: JWT_SECRET not set, using insecure dev fallback")
	}

	// --- Repositories ---
	companyRepo := repository.NewCompanyRepository(database)
	employeeRepo := repository.NewEmployeeRepository(database)
	cardRepo := repository.NewCardRepository(database)
	policyRepo := repository.NewPolicyRepository(database)
	txRepo := repository.NewTransactionRepository(database)
	ledgerRepo := repository.NewLedgerRepository(database)
	userRepo := repository.NewUserRepository(database)

	// --- Ledger (audit trail + blockchain for approved transactions) ---
	ledger := blockchain.NewLedger(ledgerRepo)

	// --- Services ---
	companySvc := service.NewCompanyService(companyRepo, ledger)
	employeeSvc := service.NewEmployeeService(employeeRepo, companyRepo, ledger)
	cardSvc := service.NewCardService(cardRepo, employeeRepo, ledger)
	policySvc := service.NewPolicyService(policyRepo, companyRepo, ledger)

	validator := contract.NewValidator()

	txSvc := service.NewTransactionService(txRepo, policyRepo, employeeRepo, cardRepo, companyRepo, validator, ledger, ledger)
	authSvc := service.NewAuthService(userRepo, jwtSecret)

	// --- Seed admin ---
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123" // dev default
		log.Println("[main] WARNING: ADMIN_PASSWORD not set, seed admin uses default 'admin123'")
	}
	if err := authSvc.EnsureAdmin("admin", adminPassword); err != nil {
		log.Fatalf("[main] seed admin failed: %v", err)
	}

	// --- Handlers ---
	companyH := handler.NewCompanyHandler(companySvc, authSvc)
	employeeH := handler.NewEmployeeHandler(employeeSvc, authSvc)
	cardH := handler.NewCardHandler(cardSvc)
	policyH := handler.NewPolicyHandler(policySvc)
	txH := handler.NewTransactionHandler(txSvc)
	ledgerH := handler.NewLedgerHandler(ledger)
	authH := handler.NewAuthHandler(authSvc)

	// --- Middleware ---
	authMW := handler.NewAuthMiddleware(authSvc)

	// --- Router ---
	router := handler.NewRouter(authH, authMW, companyH, employeeH, cardH, policyH, txH, ledgerH)

	// --- Server ---
	addr := ":8080"
	log.Printf("[main] server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("[main] server error: %v", err)
	}
}
