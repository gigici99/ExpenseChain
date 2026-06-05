package service

import (
	"fmt"
	"log"
	"time"

	"expense-chain/src/model"
	"expense-chain/src/repository"

	"github.com/google/uuid"
)

type CompanyService struct {
	repo *repository.CompanyRepository
}

func NewCompanyService(repo *repository.CompanyRepository) *CompanyService {
	return &CompanyService{repo: repo}
}

func (s *CompanyService) Create(name, vatID, address string) (*model.Company, error) {
	if name == "" {
		return nil, fmt.Errorf("CompanyService.Create: name is required")
	}
	if vatID == "" {
		return nil, fmt.Errorf("CompanyService.Create: vat_id is required")
	}

	company := &model.Company{
		ID:        uuid.NewString(),
		Name:      name,
		VatID:     vatID,
		Address:   address,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(company); err != nil {
		return nil, fmt.Errorf("CompanyService.Create: %w", err)
	}

	log.Printf("[CompanyService] created company id=%s name=%s", company.ID, company.Name)
	return company, nil
}

func (s *CompanyService) GetByID(id string) (*model.Company, error) {
	if id == "" {
		return nil, fmt.Errorf("CompanyService.GetByID: id is required")
	}
	company, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("CompanyService.GetByID: %w", err)
	}
	return company, nil
}

func (s *CompanyService) GetAll() ([]model.Company, error) {
	companies, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("CompanyService.GetAll: %w", err)
	}
	return companies, nil
}

func (s *CompanyService) Update(id, name, vatID, address string) (*model.Company, error) {
	company, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("CompanyService.Update: %w", err)
	}

	if name != "" {
		company.Name = name
	}
	if vatID != "" {
		company.VatID = vatID
	}
	if address != "" {
		company.Address = address
	}

	if err := s.repo.Update(company); err != nil {
		return nil, fmt.Errorf("CompanyService.Update: %w", err)
	}

	log.Printf("[CompanyService] updated company id=%s", company.ID)
	return company, nil
}

func (s *CompanyService) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("CompanyService.Delete: id is required")
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("CompanyService.Delete: %w", err)
	}
	log.Printf("[CompanyService] deleted company id=%s", id)
	return nil
}
