package service

import "expense-chain/src/model"

// AuditLogger records every application event into the hash-chained ledger.
// Implemented by blockchain.Ledger.
type AuditLogger interface {
	Append(entityType, entityID string, action model.LedgerAction, entity any) error
}
