package service

import (
	"fmt"
	"log"
	"time"

	"expense-chain/src/model"
	"expense-chain/src/repository"

	"github.com/google/uuid"
)

type EmployeeService struct {
	repo        *repository.EmployeeRepository
	companyRepo *repository.CompanyRepository
	audit       AuditLogger
}

func NewEmployeeService(repo *repository.EmployeeRepository, companyRepo *repository.CompanyRepository, audit AuditLogger) *EmployeeService {
	return &EmployeeService{repo: repo, companyRepo: companyRepo, audit: audit}
}

func (s *EmployeeService) Create(companyID, firstName, lastName, email, policyID string) (*model.Employee, error) {
	if companyID == "" {
		return nil, fmt.Errorf("EmployeeService.Create: company_id is required")
	}
	if firstName == "" || lastName == "" {
		return nil, fmt.Errorf("EmployeeService.Create: first_name and last_name are required")
	}
	if email == "" {
		return nil, fmt.Errorf("EmployeeService.Create: email is required")
	}

	// verify company exists
	if _, err := s.companyRepo.FindByID(companyID); err != nil {
		return nil, fmt.Errorf("EmployeeService.Create: company not found: %w", err)
	}

	employee := &model.Employee{
		ID:        uuid.NewString(),
		CompanyID: companyID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		PolicyID:  policyID,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(employee); err != nil {
		return nil, fmt.Errorf("EmployeeService.Create: %w", err)
	}

	if err := s.audit.Append("EMPLOYEE", employee.ID, model.ActionCreate, employee); err != nil {
		log.Printf("[EmployeeService] WARNING: audit append failed: %v", err)
	}

	log.Printf("[EmployeeService] created employee id=%s email=%s company_id=%s", employee.ID, employee.Email, employee.CompanyID)
	return employee, nil
}

func (s *EmployeeService) GetByID(id string) (*model.Employee, error) {
	if id == "" {
		return nil, fmt.Errorf("EmployeeService.GetByID: id is required")
	}
	employee, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("EmployeeService.GetByID: %w", err)
	}
	return employee, nil
}

func (s *EmployeeService) GetByCompanyID(companyID string) ([]model.Employee, error) {
	if companyID == "" {
		return nil, fmt.Errorf("EmployeeService.GetByCompanyID: company_id is required")
	}
	employees, err := s.repo.FindByCompanyID(companyID)
	if err != nil {
		return nil, fmt.Errorf("EmployeeService.GetByCompanyID: %w", err)
	}
	return employees, nil
}

func (s *EmployeeService) GetAll() ([]model.Employee, error) {
	employees, err := s.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("EmployeeService.GetAll: %w", err)
	}
	return employees, nil
}

func (s *EmployeeService) Update(id, firstName, lastName, email, policyID string) (*model.Employee, error) {
	employee, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("EmployeeService.Update: %w", err)
	}

	if firstName != "" {
		employee.FirstName = firstName
	}
	if lastName != "" {
		employee.LastName = lastName
	}
	if email != "" {
		employee.Email = email
	}
	if policyID != "" {
		employee.PolicyID = policyID
	}

	if err := s.repo.Update(employee); err != nil {
		return nil, fmt.Errorf("EmployeeService.Update: %w", err)
	}

	if err := s.audit.Append("EMPLOYEE", employee.ID, model.ActionUpdate, employee); err != nil {
		log.Printf("[EmployeeService] WARNING: audit append failed: %v", err)
	}

	log.Printf("[EmployeeService] updated employee id=%s", employee.ID)
	return employee, nil
}

func (s *EmployeeService) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("EmployeeService.Delete: id is required")
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("EmployeeService.Delete: %w", err)
	}
	if err := s.audit.Append("EMPLOYEE", id, model.ActionDelete, map[string]string{"id": id}); err != nil {
		log.Printf("[EmployeeService] WARNING: audit append failed: %v", err)
	}

	log.Printf("[EmployeeService] deleted employee id=%s", id)
	return nil
}
