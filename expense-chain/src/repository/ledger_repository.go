package repository

import (
	"errors"
	"fmt"
	"log"

	"expense-chain/src/model"

	"gorm.io/gorm"
)

// LedgerRepository is append-only: no Update, no Delete by design.
type LedgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) Append(entry *model.LedgerEntry) error {
	if err := r.db.Create(entry).Error; err != nil {
		return fmt.Errorf("LedgerRepository.Append: %w", err)
	}
	log.Printf("[Ledger] appended seq=%d entity=%s action=%s entity_id=%s",
		entry.Sequence, entry.EntityType, entry.Action, entry.EntityID)
	return nil
}

// FindLast returns the most recent entry, or nil if the ledger is empty.
func (r *LedgerRepository) FindLast() (*model.LedgerEntry, error) {
	var entry model.LedgerEntry
	err := r.db.Order("sequence DESC").First(&entry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("LedgerRepository.FindLast: %w", err)
	}
	return &entry, nil
}

// FindAll returns the full ledger ordered by sequence ascending.
func (r *LedgerRepository) FindAll() ([]model.LedgerEntry, error) {
	var entries []model.LedgerEntry
	if err := r.db.Order("sequence ASC").Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("LedgerRepository.FindAll: %w", err)
	}
	return entries, nil
}

func (r *LedgerRepository) FindByEntityID(entityID string) ([]model.LedgerEntry, error) {
	var entries []model.LedgerEntry
	if err := r.db.Where("entity_id = ?", entityID).Order("sequence ASC").Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("LedgerRepository.FindByEntityID: %w", err)
	}
	return entries, nil
}
