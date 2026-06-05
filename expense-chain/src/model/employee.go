package model

import "time"

type Employee struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	PolicyID  string    `json:"policy_id"`
	CreatedAt time.Time `json:"created_at"`
}
