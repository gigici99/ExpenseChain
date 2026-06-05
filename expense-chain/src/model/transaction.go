package model

import "time"

// TransactionStatus tracks validation result from smart contract
type TransactionStatus string

const (
	StatusPending  TransactionStatus = "PENDING"   // submitted, not yet validated
	StatusApproved TransactionStatus = "APPROVED"  // passed policy checks
	StatusRejected TransactionStatus = "REJECTED"  // blocked by policy
)

type Transaction struct {
	ID          string              `json:"id"`
	Company     Company             `json:"company" gorm:"serializer:json"`
	Employee    Employee            `json:"employee" gorm:"serializer:json"`
	Card        Card                `json:"card" gorm:"serializer:json"`
	Policy      Policy              `json:"policy" gorm:"serializer:json"`
	Amount      float64             `json:"amount"`
	Currency    string              `json:"currency"`     // ISO 4217, e.g. "EUR"
	Category    TransactionCategory `json:"category"`
	Description string              `json:"description"`
	Status      TransactionStatus   `json:"status"`
	RejectReason string             `json:"reject_reason,omitempty"` // set by smart contract on rejection
	CreatedAt   time.Time           `json:"created_at"`
	ValidatedAt *time.Time          `json:"validated_at,omitempty"`  // nil until smart contract runs
}
