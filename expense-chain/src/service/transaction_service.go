package service

import (
	"fmt"
	"log"
	"time"

	"expense-chain/src/model"
	"expense-chain/src/repository"

	"github.com/google/uuid"
)

// ValidationResult is returned by the smart contract validator.
// Defined here as interface to keep contract package decoupled.
type ValidationResult struct {
	Status       model.TransactionStatus
	RejectReason string
}

// Validator is the smart contract interface — contract package implements this.
// It is a pure function: receives all data it needs (including pre-computed spending
// totals) and returns a decision. No DB access, no side effects.
type Validator interface {
	Validate(tx *model.Transaction, policy *model.Policy, spentToday, spentMonth float64) ValidationResult
}

// BlockchainWriter is the blockchain interface — blockchain package implements this.
type BlockchainWriter interface {
	AddBlock(tx *model.Transaction) error
}

type TransactionService struct {
	txRepo      *repository.TransactionRepository
	policyRepo  *repository.PolicyRepository
	employeeRepo *repository.EmployeeRepository
	cardRepo    *repository.CardRepository
	companyRepo *repository.CompanyRepository
	validator   Validator
	blockchain  BlockchainWriter
	audit       AuditLogger
}

func NewTransactionService(
	txRepo *repository.TransactionRepository,
	policyRepo *repository.PolicyRepository,
	employeeRepo *repository.EmployeeRepository,
	cardRepo *repository.CardRepository,
	companyRepo *repository.CompanyRepository,
	validator Validator,
	blockchain BlockchainWriter,
	audit AuditLogger,
) *TransactionService {
	return &TransactionService{
		txRepo:       txRepo,
		policyRepo:   policyRepo,
		employeeRepo: employeeRepo,
		cardRepo:     cardRepo,
		companyRepo:  companyRepo,
		validator:    validator,
		blockchain:   blockchain,
		audit:        audit,
	}
}

// Submit is the main entry point: hydrates all entities, runs smart contract, persists, writes to chain.
func (s *TransactionService) Submit(
	employeeID string,
	cardID string,
	amount float64,
	currency string,
	category model.TransactionCategory,
	description string,
) (*model.Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("TransactionService.Submit: amount must be > 0")
	}
	if currency == "" {
		return nil, fmt.Errorf("TransactionService.Submit: currency is required")
	}

	// hydrate employee
	employee, err := s.employeeRepo.FindByID(employeeID)
	if err != nil {
		return nil, fmt.Errorf("TransactionService.Submit: employee not found: %w", err)
	}

	// hydrate company
	company, err := s.companyRepo.FindByID(employee.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("TransactionService.Submit: company not found: %w", err)
	}

	// hydrate card
	card, err := s.cardRepo.FindByID(cardID)
	if err != nil {
		return nil, fmt.Errorf("TransactionService.Submit: card not found: %w", err)
	}
	if card.EmployeeID != employeeID {
		return nil, fmt.Errorf("TransactionService.Submit: card does not belong to employee")
	}

	// hydrate policy
	policy, err := s.policyRepo.FindByID(employee.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("TransactionService.Submit: policy not found: %w", err)
	}

	now := time.Now()
	tx := &model.Transaction{
		ID:          uuid.NewString(),
		EmployeeID:  employee.ID,
		CompanyID:   company.ID,
		Company:     *company,
		Employee:    *employee,
		Card:        *card,
		Policy:      *policy,
		Amount:      amount,
		Currency:    currency,
		Category:    category,
		Description: description,
		Status:      model.StatusPending,
		CreatedAt:   now,
	}

	// pre-compute spending totals for daily/monthly caps (contract stays pure)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	spentToday, err := s.txRepo.SumApprovedBetween(employee.ID, startOfDay, now)
	if err != nil {
		return nil, fmt.Errorf("TransactionService.Submit: sum today failed: %w", err)
	}
	spentMonth, err := s.txRepo.SumApprovedBetween(employee.ID, startOfMonth, now)
	if err != nil {
		return nil, fmt.Errorf("TransactionService.Submit: sum month failed: %w", err)
	}

	// smart contract validation — pure function, all inputs provided
	result := s.validator.Validate(tx, policy, spentToday, spentMonth)
	tx.Status = result.Status
	tx.RejectReason = result.RejectReason
	validatedAt := time.Now()
	tx.ValidatedAt = &validatedAt

	log.Printf("[TransactionService] validated id=%s status=%s", tx.ID, tx.Status)

	// persist transaction (approved and rejected — both go to DB for audit)
	if err := s.txRepo.Create(tx); err != nil {
		return nil, fmt.Errorf("TransactionService.Submit: persist failed: %w", err)
	}

	// only approved transactions enter the blockchain
	if tx.Status == model.StatusApproved {
		if err := s.blockchain.AddBlock(tx); err != nil {
			// log but don't fail — tx already saved to DB
			log.Printf("[TransactionService] WARNING: blockchain write failed id=%s err=%v", tx.ID, err)
		}
	} else {
		// rejected transactions also enter the ledger — full audit trail
		if err := s.audit.Append("TRANSACTION", tx.ID, model.ActionRejected, tx); err != nil {
			log.Printf("[TransactionService] WARNING: audit append failed: %v", err)
		}
		log.Printf("[TransactionService] rejected id=%s reason=%s", tx.ID, tx.RejectReason)
	}

	return tx, nil
}

func (s *TransactionService) GetByID(id string) (*model.Transaction, error) {
	if id == "" {
		return nil, fmt.Errorf("TransactionService.GetByID: id is required")
	}
	tx, err := s.txRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("TransactionService.GetByID: %w", err)
	}
	return tx, nil
}

func (s *TransactionService) GetByEmployeeID(employeeID string) ([]model.Transaction, error) {
	if employeeID == "" {
		return nil, fmt.Errorf("TransactionService.GetByEmployeeID: employee_id is required")
	}
	txs, err := s.txRepo.FindByEmployeeID(employeeID)
	if err != nil {
		return nil, fmt.Errorf("TransactionService.GetByEmployeeID: %w", err)
	}
	return txs, nil
}

func (s *TransactionService) GetByStatus(status model.TransactionStatus) ([]model.Transaction, error) {
	txs, err := s.txRepo.FindByStatus(status)
	if err != nil {
		return nil, fmt.Errorf("TransactionService.GetByStatus: %w", err)
	}
	return txs, nil
}

func (s *TransactionService) GetAll() ([]model.Transaction, error) {
	txs, err := s.txRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("TransactionService.GetAll: %w", err)
	}
	return txs, nil
}
