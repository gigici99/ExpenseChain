package repository

import (
	"errors"
	"fmt"
	"log"

	"expense-chain/src/model"

	"gorm.io/gorm"
)

type PolicyRepository struct {
	db *gorm.DB
}

func NewPolicyRepository(db *gorm.DB) *PolicyRepository {
	return &PolicyRepository{db: db}
}

func (r *PolicyRepository) Create(policy *model.Policy) error {
	if err := r.db.Create(policy).Error; err != nil {
		return fmt.Errorf("PolicyRepository.Create: %w", err)
	}
	log.Printf("[Policy] created id=%s company_id=%s name=%s", policy.ID, policy.CompanyID, policy.Name)
	return nil
}

func (r *PolicyRepository) FindByID(id string) (*model.Policy, error) {
	var policy model.Policy
	err := r.db.First(&policy, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("PolicyRepository.FindByID: not found id=%s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("PolicyRepository.FindByID: %w", err)
	}
	return &policy, nil
}

func (r *PolicyRepository) FindByCompanyID(companyID string) ([]model.Policy, error) {
	var policies []model.Policy
	if err := r.db.Where("company_id = ?", companyID).Find(&policies).Error; err != nil {
		return nil, fmt.Errorf("PolicyRepository.FindByCompanyID: %w", err)
	}
	log.Printf("[Policy] found %d records for company_id=%s", len(policies), companyID)
	return policies, nil
}

func (r *PolicyRepository) FindAll() ([]model.Policy, error) {
	var policies []model.Policy
	if err := r.db.Find(&policies).Error; err != nil {
		return nil, fmt.Errorf("PolicyRepository.FindAll: %w", err)
	}
	log.Printf("[Policy] found %d records", len(policies))
	return policies, nil
}

func (r *PolicyRepository) Update(policy *model.Policy) error {
	result := r.db.Save(policy)
	if result.Error != nil {
		return fmt.Errorf("PolicyRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("PolicyRepository.Update: no rows affected id=%s", policy.ID)
	}
	log.Printf("[Policy] updated id=%s", policy.ID)
	return nil
}

func (r *PolicyRepository) Delete(id string) error {
	result := r.db.Delete(&model.Policy{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("PolicyRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("PolicyRepository.Delete: not found id=%s", id)
	}
	log.Printf("[Policy] deleted id=%s", id)
	return nil
}
