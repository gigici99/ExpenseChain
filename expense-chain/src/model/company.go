package model

import "time"

type Company struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	VatID     string    `json:"vat_id"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
}
