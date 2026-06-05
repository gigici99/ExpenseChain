package repository

import (
	"errors"
	"fmt"
	"log"

	"expense-chain/src/model"

	"gorm.io/gorm"
)

type CardRepository struct {
	db *gorm.DB
}

func NewCardRepository(db *gorm.DB) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) Create(card *model.Card) error {
	if err := r.db.Create(card).Error; err != nil {
		return fmt.Errorf("CardRepository.Create: %w", err)
	}
	log.Printf("[Card] created id=%s employee_id=%s type=%s", card.ID, card.EmployeeID, card.Type)
	return nil
}

func (r *CardRepository) FindByID(id string) (*model.Card, error) {
	var card model.Card
	err := r.db.First(&card, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("CardRepository.FindByID: not found id=%s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("CardRepository.FindByID: %w", err)
	}
	return &card, nil
}

func (r *CardRepository) FindByEmployeeID(employeeID string) ([]model.Card, error) {
	var cards []model.Card
	if err := r.db.Where("employee_id = ?", employeeID).Find(&cards).Error; err != nil {
		return nil, fmt.Errorf("CardRepository.FindByEmployeeID: %w", err)
	}
	log.Printf("[Card] found %d records for employee_id=%s", len(cards), employeeID)
	return cards, nil
}

func (r *CardRepository) Update(card *model.Card) error {
	result := r.db.Save(card)
	if result.Error != nil {
		return fmt.Errorf("CardRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("CardRepository.Update: no rows affected id=%s", card.ID)
	}
	log.Printf("[Card] updated id=%s", card.ID)
	return nil
}

func (r *CardRepository) Delete(id string) error {
	result := r.db.Delete(&model.Card{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("CardRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("CardRepository.Delete: not found id=%s", id)
	}
	log.Printf("[Card] deleted id=%s", id)
	return nil
}
