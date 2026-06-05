package main

import (
	"log"
	"net/http"

	"expense-chain/src/blockchain"
	"expense-chain/src/contract"
	"expense-chain/src/db"
	"expense-chain/src/handler"
	"expense-chain/src/repository"
	"expense-chain/src/service"

	"expense-chain/src/model"
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
	); err != nil {
		log.Fatalf("[main] AutoMigrate failed: %v", err)
	}

	// --- Repositories ---
	companyRepo := repository.NewCompanyRepository(database)
	employeeRepo := repository.NewEmployeeRepository(database)
	cardRepo := repository.NewCardRepository(database)
	policyRepo := repository.NewPolicyRepository(database)
	txRepo := repository.NewTransactionRepository(database)

	// --- Services ---
	companySvc := service.NewCompanyService(companyRepo)
	employeeSvc := service.NewEmployeeService(employeeRepo, companyRepo)
	cardSvc := service.NewCardService(cardRepo, employeeRepo)
	policySvc := service.NewPolicyService(policyRepo, companyRepo)

	validator := contract.NewValidator()
	chain := blockchain.NewChain()

	txSvc := service.NewTransactionService(txRepo, policyRepo, employeeRepo, cardRepo, companyRepo, validator, chain)

	// --- Handlers ---
	companyH := handler.NewCompanyHandler(companySvc)
	employeeH := handler.NewEmployeeHandler(employeeSvc)
	cardH := handler.NewCardHandler(cardSvc)
	policyH := handler.NewPolicyHandler(policySvc)
	txH := handler.NewTransactionHandler(txSvc)

	// --- Router ---
	router := handler.NewRouter(companyH, employeeH, cardH, policyH, txH)

	// --- Server ---
	addr := ":8080"
	log.Printf("[main] server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("[main] server error: %v", err)
	}
}
