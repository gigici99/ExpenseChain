package service

import (
	"fmt"
	"log"
	"time"

	"expense-chain/src/model"
	"expense-chain/src/repository"

	"github.com/google/uuid"
)

type PolicyService struct {
	repo        *repository.PolicyRepository
	companyRepo *repository.CompanyRepository
	audit       AuditLogger
}

func NewPolicyService(repo *repository.PolicyRepository, companyRepo *repository.CompanyRepository, audit AuditLogger) *PolicyService {
	return &PolicyService{repo: repo, companyRepo: companyRepo, audit: audit}
}

func (s *PolicyService) Create(
	companyID string,
	name string,
	maxPerTx float64,
	maxPerDay float64,
	maxPerMonth float64,
	blockedCategories []model.TransactionCategory,
	allowedCardTypes []model.CardType,
) (*model.Policy, error) {
	if companyID == "" {
		return nil, fmt.Errorf("PolicyService.Create: company_id is required")
	}
	if name == "" {
		return nil, fmt.Errorf("PolicyService.Create: name is required")
	}
	if maxPerTx <= 0 {
		return nil, fmt.Errorf("PolicyService.Create: max_amount_per_tx must be > 0")
	}

	// verify company exists
	if _, err := s.companyRepo.FindByID(companyID); err != nil {
		return nil, fmt.Errorf("PolicyService.Create: company not found: %w", err)
	}

	policy := &model.Policy{
		ID:                uuid.NewString(),
		CompanyID:         companyID,
		Name:              name,
		MaxAmountPerTx:    maxPerTx,
		MaxAmountPerDay:   maxPerDay,
		MaxAmountPerMonth: maxPerMonth,
		BlockedCategories: blockedCategories,
		AllowedCardTypes:  allowedCardTypes,
		CreatedAt:         time.Now(),
	}

	if err := s.repo.Create(policy); err != nil {
		return nil, fmt.Errorf("PolicyService.Create: %w", err)
	}

	if err := s.audit.Append("POLICY", policy.ID, model.ActionCreate, policy); err != nil {
		log.Printf("[PolicyService] WARNING: audit append failed: %v", err)
	}

	log.Printf("[PolicyService] created policy id=%s company_id=%s name=%s", policy.ID, policy.CompanyID, policy.Name)
	return policy, nil
}

func (s *PolicyService) GetByID(id string) (*model.Policy, error) {
	if id == "" {
		return nil, fmt.Errorf("PolicyService.GetByID: id is required")
	}
	policy, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("PolicyService.GetByID: %w", err)
	}
	return policy, nil
}

func (s *PolicyService) GetByCompanyID(companyID string) ([]model.Policy, error) {
	if companyID == "" {
		return nil, fmt.Errorf("PolicyService.GetByCompanyID: company_id is required")
	}
	policies, err := s.repo.FindByCompanyID(companyID)
	if err != nil {
		return nil, fmt.Errorf("PolicyService.GetByCompanyID: %w", err)
	}
	return policies, nil
}

func (s *PolicyService) GetAll() ([]model.Policy, error) {
	policies, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("PolicyService.GetAll: %w", err)
	}
	return policies, nil
}

func (s *PolicyService) Update(
	id string,
	name string,
	maxPerTx float64,
	maxPerDay float64,
	maxPerMonth float64,
	blockedCategories []model.TransactionCategory,
	allowedCardTypes []model.CardType,
) (*model.Policy, error) {
	policy, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("PolicyService.Update: %w", err)
	}

	if name != "" {
		policy.Name = name
	}
	if maxPerTx > 0 {
		policy.MaxAmountPerTx = maxPerTx
	}
	if maxPerDay > 0 {
		policy.MaxAmountPerDay = maxPerDay
	}
	if maxPerMonth > 0 {
		policy.MaxAmountPerMonth = maxPerMonth
	}
	if blockedCategories != nil {
		policy.BlockedCategories = blockedCategories
	}
	if allowedCardTypes != nil {
		policy.AllowedCardTypes = allowedCardTypes
	}

	if err := s.repo.Update(policy); err != nil {
		return nil, fmt.Errorf("PolicyService.Update: %w", err)
	}

	if err := s.audit.Append("POLICY", policy.ID, model.ActionUpdate, policy); err != nil {
		log.Printf("[PolicyService] WARNING: audit append failed: %v", err)
	}

	log.Printf("[PolicyService] updated policy id=%s", policy.ID)
	return policy, nil
}

func (s *PolicyService) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("PolicyService.Delete: id is required")
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("PolicyService.Delete: %w", err)
	}
	if err := s.audit.Append("POLICY", id, model.ActionDelete, map[string]string{"id": id}); err != nil {
		log.Printf("[PolicyService] WARNING: audit append failed: %v", err)
	}

	log.Printf("[PolicyService] deleted policy id=%s", id)
	return nil
}
