package repository

import (
	"errors"
	"fmt"
	"log"

	"expense-chain/src/model"

	"gorm.io/gorm"
)

type EmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) Create(employee *model.Employee) error {
	if err := r.db.Create(employee).Error; err != nil {
		return fmt.Errorf("EmployeeRepository.Create: %w", err)
	}
	log.Printf("[Employee] created id=%s email=%s", employee.ID, employee.Email)
	return nil
}

func (r *EmployeeRepository) FindByID(id string) (*model.Employee, error) {
	var employee model.Employee
	err := r.db.First(&employee, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("EmployeeRepository.FindByID: not found id=%s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("EmployeeRepository.FindByID: %w", err)
	}
	return &employee, nil
}

func (r *EmployeeRepository) FindByCompanyID(companyID string) ([]model.Employee, error) {
	var employees []model.Employee
	if err := r.db.Where("company_id = ?", companyID).Find(&employees).Error; err != nil {
		return nil, fmt.Errorf("EmployeeRepository.FindByCompanyID: %w", err)
	}
	log.Printf("[Employee] found %d records for company_id=%s", len(employees), companyID)
	return employees, nil
}

func (r *EmployeeRepository) FindAll() ([]model.Employee, error) {
	var employees []model.Employee
	if err := r.db.Find(&employees).Error; err != nil {
		return nil, fmt.Errorf("EmployeeRepository.FindAll: %w", err)
	}
	log.Printf("[Employee] found %d records", len(employees))
	return employees, nil
}

func (r *EmployeeRepository) Update(employee *model.Employee) error {
	result := r.db.Save(employee)
	if result.Error != nil {
		return fmt.Errorf("EmployeeRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("EmployeeRepository.Update: no rows affected id=%s", employee.ID)
	}
	log.Printf("[Employee] updated id=%s", employee.ID)
	return nil
}

func (r *EmployeeRepository) Delete(id string) error {
	result := r.db.Delete(&model.Employee{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("EmployeeRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("EmployeeRepository.Delete: not found id=%s", id)
	}
	log.Printf("[Employee] deleted id=%s", id)
	return nil
}
