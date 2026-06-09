package repository

import (
	"errors"
	"fmt"
	"log"

	"expense-chain/src/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("UserRepository.Create: %w", err)
	}
	log.Printf("[User] created id=%s username=%s role=%s", user.ID, user.Username, user.Role)
	return nil
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, "username = ?", username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("UserRepository.FindByUsername: not found username=%s", username)
	}
	if err != nil {
		return nil, fmt.Errorf("UserRepository.FindByUsername: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("UserRepository.FindByID: not found id=%s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("UserRepository.FindByID: %w", err)
	}
	return &user, nil
}
