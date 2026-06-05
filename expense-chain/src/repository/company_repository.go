package repository

import (
	"errors"
	"fmt"
	"log"

	"expense-chain/src/model"

	"gorm.io/gorm"
)

type CompanyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

func (r *CompanyRepository) Create(company *model.Company) error {
	if err := r.db.Create(company).Error; err != nil {
		return fmt.Errorf("CompanyRepository.Create: %w", err)
	}
	log.Printf("[Company] created id=%s name=%s", company.ID, company.Name)
	return nil
}

func (r *CompanyRepository) FindByID(id string) (*model.Company, error) {
	var company model.Company
	err := r.db.First(&company, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("CompanyRepository.FindByID: not found id=%s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("CompanyRepository.FindByID: %w", err)
	}
	return &company, nil
}

func (r *CompanyRepository) FindAll() ([]model.Company, error) {
	var companies []model.Company
	if err := r.db.Find(&companies).Error; err != nil {
		return nil, fmt.Errorf("CompanyRepository.FindAll: %w", err)
	}
	log.Printf("[Company] found %d records", len(companies))
	return companies, nil
}

func (r *CompanyRepository) Update(company *model.Company) error {
	result := r.db.Save(company)
	if result.Error != nil {
		return fmt.Errorf("CompanyRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("CompanyRepository.Update: no rows affected id=%s", company.ID)
	}
	log.Printf("[Company] updated id=%s", company.ID)
	return nil
}

func (r *CompanyRepository) Delete(id string) error {
	result := r.db.Delete(&model.Company{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("CompanyRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("CompanyRepository.Delete: not found id=%s", id)
	}
	log.Printf("[Company] deleted id=%s", id)
	return nil
}
