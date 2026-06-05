package service

import (
	"fmt"
	"log"
	"time"

	"expense-chain/src/model"
	"expense-chain/src/repository"

	"github.com/google/uuid"
)

type CardService struct {
	repo         *repository.CardRepository
	employeeRepo *repository.EmployeeRepository
}

func NewCardService(repo *repository.CardRepository, employeeRepo *repository.EmployeeRepository) *CardService {
	return &CardService{repo: repo, employeeRepo: employeeRepo}
}

func (s *CardService) Create(employeeID string, cardType model.CardType, last4, holder, provider string, expiresAt time.Time) (*model.Card, error) {
	if employeeID == "" {
		return nil, fmt.Errorf("CardService.Create: employee_id is required")
	}
	if last4 == "" || len(last4) != 4 {
		return nil, fmt.Errorf("CardService.Create: last4 must be exactly 4 digits")
	}
	if holder == "" {
		return nil, fmt.Errorf("CardService.Create: holder is required")
	}
	if cardType == "" {
		return nil, fmt.Errorf("CardService.Create: card type is required")
	}

	// verify employee exists
	if _, err := s.employeeRepo.FindByID(employeeID); err != nil {
		return nil, fmt.Errorf("CardService.Create: employee not found: %w", err)
	}

	card := &model.Card{
		ID:         uuid.NewString(),
		EmployeeID: employeeID,
		Type:       cardType,
		Last4:      last4,
		Holder:     holder,
		Provider:   provider,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.Create(card); err != nil {
		return nil, fmt.Errorf("CardService.Create: %w", err)
	}

	log.Printf("[CardService] created card id=%s employee_id=%s type=%s last4=%s", card.ID, card.EmployeeID, card.Type, card.Last4)
	return card, nil
}

func (s *CardService) GetByID(id string) (*model.Card, error) {
	if id == "" {
		return nil, fmt.Errorf("CardService.GetByID: id is required")
	}
	card, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("CardService.GetByID: %w", err)
	}
	return card, nil
}

func (s *CardService) GetByEmployeeID(employeeID string) ([]model.Card, error) {
	if employeeID == "" {
		return nil, fmt.Errorf("CardService.GetByEmployeeID: employee_id is required")
	}
	cards, err := s.repo.FindByEmployeeID(employeeID)
	if err != nil {
		return nil, fmt.Errorf("CardService.GetByEmployeeID: %w", err)
	}
	return cards, nil
}

func (s *CardService) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("CardService.Delete: id is required")
	}
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("CardService.Delete: %w", err)
	}
	log.Printf("[CardService] deleted card id=%s", id)
	return nil
}
