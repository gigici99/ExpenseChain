package model

import "time"

// TransactionCategory represents merchant/expense categories that a policy can restrict
type TransactionCategory string

const (
	CategoryFood        TransactionCategory = "FOOD"
	CategoryTravel      TransactionCategory = "TRAVEL"
	CategoryAccommodation TransactionCategory = "ACCOMMODATION"
	CategoryEntertainment TransactionCategory = "ENTERTAINMENT"
	CategoryElectronics TransactionCategory = "ELECTRONICS"
	CategoryOther       TransactionCategory = "OTHER"
)

type Policy struct {
	ID                string                `json:"id"`
	CompanyID         string                `json:"company_id"`
	Name              string                `json:"name"`
	MaxAmountPerTx    float64               `json:"max_amount_per_tx"`    // max € per single transaction
	MaxAmountPerDay   float64               `json:"max_amount_per_day"`   // daily spending cap
	MaxAmountPerMonth float64               `json:"max_amount_per_month"` // monthly spending cap
	BlockedCategories []TransactionCategory `json:"blocked_categories" gorm:"serializer:json"`   // categories not allowed
	AllowedCardTypes  []CardType            `json:"allowed_card_types" gorm:"serializer:json"`   // e.g. no virtual cards
	CreatedAt         time.Time             `json:"created_at"`
}
