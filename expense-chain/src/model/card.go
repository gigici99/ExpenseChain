package model

import "time"

// CardType represents the payment instrument type
type CardType string

const (
	CardTypePhysical  CardType = "PHYSICAL"  // physical credit/debit card
	CardTypeVirtual   CardType = "VIRTUAL"   // virtual card (e.g. Google Pay, Apple Pay)
	CardTypePrepaid   CardType = "PREPAID"
)

type Card struct {
	ID         string    `json:"id"`
	EmployeeID string    `json:"employee_id"`
	Type       CardType  `json:"type"`
	Last4      string    `json:"last4"`       // last 4 digits (never store full PAN)
	Holder     string    `json:"holder"`
	Provider   string    `json:"provider"`    // e.g. "Visa", "Mastercard", "Google Pay"
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}
