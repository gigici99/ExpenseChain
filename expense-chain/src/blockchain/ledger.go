package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"expense-chain/src/model"
	"expense-chain/src/repository"

	"github.com/google/uuid"
)

const genesisHash = "GENESIS"

// Ledger is the hash-chained, append-only audit log.
// It records every application event (CREATE/UPDATE/DELETE on entities)
// and every transaction validation outcome (VALIDATED/REJECTED).
// It also implements service.BlockchainWriter for approved transactions.
type Ledger struct {
	repo *repository.LedgerRepository
	mu   sync.Mutex // serializes appends so sequence + prev_hash stay consistent
}

func NewLedger(repo *repository.LedgerRepository) *Ledger {
	return &Ledger{repo: repo}
}

// computeHash builds the SHA-256 over all the fields that must be tamper-evident.
func computeHash(seq uint64, entityType, entityID string, action model.LedgerAction, payload, prevHash string, createdAt time.Time) string {
	data := fmt.Sprintf("%d|%s|%s|%s|%s|%s|%d",
		seq, entityType, entityID, action, payload, prevHash, createdAt.UnixNano())
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

// Append serializes the entity snapshot and writes a new chained entry.
func (l *Ledger) Append(entityType, entityID string, action model.LedgerAction, entity any) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	payloadBytes, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("Ledger.Append: payload marshal failed: %w", err)
	}
	payload := string(payloadBytes)

	last, err := l.repo.FindLast()
	if err != nil {
		return fmt.Errorf("Ledger.Append: %w", err)
	}

	var seq uint64 = 1
	prevHash := genesisHash
	if last != nil {
		seq = last.Sequence + 1
		prevHash = last.Hash
	}

	now := time.Now()
	entry := &model.LedgerEntry{
		ID:         uuid.NewString(),
		Sequence:   seq,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Payload:    payload,
		PrevHash:   prevHash,
		Hash:       computeHash(seq, entityType, entityID, action, payload, prevHash, now),
		CreatedAt:  now,
	}

	return l.repo.Append(entry)
}

// AddBlock implements service.BlockchainWriter — approved transactions enter the chain here.
func (l *Ledger) AddBlock(tx *model.Transaction) error {
	return l.Append("TRANSACTION", tx.ID, model.ActionValidated, tx)
}

// VerifyResult reports the outcome of an integrity check.
type VerifyResult struct {
	Valid      bool   `json:"valid"`
	Entries    int    `json:"entries"`
	BrokenAt   uint64 `json:"broken_at,omitempty"`   // sequence of first corrupted entry
	BrokenWhy  string `json:"broken_why,omitempty"`  // hash mismatch or chain link mismatch
}

// Verify walks the full chain recomputing every hash and checking linkage.
func (l *Ledger) Verify() (*VerifyResult, error) {
	entries, err := l.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("Ledger.Verify: %w", err)
	}

	result := &VerifyResult{Valid: true, Entries: len(entries)}
	prevHash := genesisHash

	for _, e := range entries {
		if e.PrevHash != prevHash {
			result.Valid = false
			result.BrokenAt = e.Sequence
			result.BrokenWhy = "chain link mismatch: prev_hash does not match previous entry hash"
			log.Printf("[Ledger] VERIFY FAILED at seq=%d: link mismatch", e.Sequence)
			return result, nil
		}
		recomputed := computeHash(e.Sequence, e.EntityType, e.EntityID, e.Action, e.Payload, e.PrevHash, e.CreatedAt)
		if recomputed != e.Hash {
			result.Valid = false
			result.BrokenAt = e.Sequence
			result.BrokenWhy = "hash mismatch: entry data was modified"
			log.Printf("[Ledger] VERIFY FAILED at seq=%d: hash mismatch", e.Sequence)
			return result, nil
		}
		prevHash = e.Hash
	}

	log.Printf("[Ledger] verify OK — %d entries, chain intact", len(entries))
	return result, nil
}

// GetAll returns the full ledger.
func (l *Ledger) GetAll() ([]model.LedgerEntry, error) {
	return l.repo.FindAll()
}

// GetByEntityID returns the audit history of a single entity.
func (l *Ledger) GetByEntityID(entityID string) ([]model.LedgerEntry, error) {
	return l.repo.FindByEntityID(entityID)
}
