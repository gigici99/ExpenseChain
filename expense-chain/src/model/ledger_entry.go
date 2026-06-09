package model

import "time"

// LedgerAction describes what happened to an entity
type LedgerAction string

const (
	ActionCreate    LedgerAction = "CREATE"
	ActionUpdate    LedgerAction = "UPDATE"
	ActionDelete    LedgerAction = "DELETE"
	ActionValidated LedgerAction = "VALIDATED" // transaction approved by smart contract
	ActionRejected  LedgerAction = "REJECTED"  // transaction rejected by smart contract
)

// LedgerEntry is one immutable record in the hash-chained audit ledger.
// Append-only: never updated, never deleted.
type LedgerEntry struct {
	ID         string       `json:"id"`
	Sequence   uint64       `json:"sequence" gorm:"uniqueIndex"`
	EntityType string       `json:"entity_type"` // COMPANY | EMPLOYEE | CARD | POLICY | TRANSACTION
	EntityID   string       `json:"entity_id"`
	Action     LedgerAction `json:"action"`
	Payload    string       `json:"payload"`   // JSON snapshot of the entity at event time
	PrevHash   string       `json:"prev_hash"` // hash of previous entry ("GENESIS" for first)
	Hash       string       `json:"hash"`      // SHA-256 over this entry's fields + PrevHash
	CreatedAt  time.Time    `json:"created_at"`
}
