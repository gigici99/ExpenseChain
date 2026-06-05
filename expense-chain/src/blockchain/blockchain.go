package blockchain

import (
	"log"

	"expense-chain/src/model"
)

// Chain implements service.BlockchainWriter interface.
// Blockchain logic goes here — for now just logs.
type Chain struct{}

func NewChain() *Chain {
	return &Chain{}
}

func (c *Chain) AddBlock(tx *model.Transaction) error {
	// TODO: implement block creation, hashing, chain append
	log.Printf("[Blockchain] stub — would add block for tx id=%s amount=%.2f %s", tx.ID, tx.Amount, tx.Currency)
	return nil
}
