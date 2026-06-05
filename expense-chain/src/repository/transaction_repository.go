package repository

import (
	"errors"
	"fmt"
	"log"

	"expense-chain/src/model"

	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(tx *model.Transaction) error {
	if err := r.db.Create(tx).Error; err != nil {
		return fmt.Errorf("TransactionRepository.Create: %w", err)
	}
	log.Printf("[Transaction] created id=%s employee_id=%s amount=%.2f %s status=%s",
		tx.ID, tx.Employee.ID, tx.Amount, tx.Currency, tx.Status)
	return nil
}

func (r *TransactionRepository) FindByID(id string) (*model.Transaction, error) {
	var tx model.Transaction
	err := r.db.First(&tx, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("TransactionRepository.FindByID: not found id=%s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("TransactionRepository.FindByID: %w", err)
	}
	return &tx, nil
}

func (r *TransactionRepository) FindByEmployeeID(employeeID string) ([]model.Transaction, error) {
	var txs []model.Transaction
	if err := r.db.Where("employee__i_d = ?", employeeID).Find(&txs).Error; err != nil {
		return nil, fmt.Errorf("TransactionRepository.FindByEmployeeID: %w", err)
	}
	log.Printf("[Transaction] found %d records for employee_id=%s", len(txs), employeeID)
	return txs, nil
}

func (r *TransactionRepository) FindByStatus(status model.TransactionStatus) ([]model.Transaction, error) {
	var txs []model.Transaction
	if err := r.db.Where("status = ?", status).Find(&txs).Error; err != nil {
		return nil, fmt.Errorf("TransactionRepository.FindByStatus: %w", err)
	}
	log.Printf("[Transaction] found %d records with status=%s", len(txs), status)
	return txs, nil
}

func (r *TransactionRepository) FindAll() ([]model.Transaction, error) {
	var txs []model.Transaction
	if err := r.db.Find(&txs).Error; err != nil {
		return nil, fmt.Errorf("TransactionRepository.FindAll: %w", err)
	}
	log.Printf("[Transaction] found %d records", len(txs))
	return txs, nil
}

// UpdateStatus aggiorna solo status e reject_reason — usato dallo smart contract dopo validazione.
func (r *TransactionRepository) UpdateStatus(id string, status model.TransactionStatus, rejectReason string) error {
	result := r.db.Model(&model.Transaction{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        status,
			"reject_reason": rejectReason,
		})
	if result.Error != nil {
		return fmt.Errorf("TransactionRepository.UpdateStatus: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("TransactionRepository.UpdateStatus: not found id=%s", id)
	}
	log.Printf("[Transaction] status updated id=%s status=%s", id, status)
	return nil
}

func (r *TransactionRepository) Delete(id string) error {
	result := r.db.Delete(&model.Transaction{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("TransactionRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("TransactionRepository.Delete: not found id=%s", id)
	}
	log.Printf("[Transaction] deleted id=%s", id)
	return nil
}
