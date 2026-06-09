package contract

import (
	"fmt"

	"expense-chain/src/model"
	"expense-chain/src/service"
)

// Validator implements service.Validator — the smart contract.
// Pure function: given a transaction, its policy, and pre-computed spending
// totals, it deterministically returns APPROVED or REJECTED. No DB, no state.
type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(tx *model.Transaction, policy *model.Policy, spentToday, spentMonth float64) service.ValidationResult {
	// Rule 1 — per-transaction limit
	if policy.MaxAmountPerTx > 0 && tx.Amount > policy.MaxAmountPerTx {
		return reject(fmt.Sprintf("amount %.2f exceeds per-transaction limit %.2f", tx.Amount, policy.MaxAmountPerTx))
	}

	// Rule 2 — blocked category
	for _, blocked := range policy.BlockedCategories {
		if tx.Category == blocked {
			return reject(fmt.Sprintf("category %s is blocked by policy", tx.Category))
		}
	}

	// Rule 3 — card type not allowed (only enforced if policy whitelists types)
	if len(policy.AllowedCardTypes) > 0 {
		allowed := false
		for _, t := range policy.AllowedCardTypes {
			if tx.Card.Type == t {
				allowed = true
				break
			}
		}
		if !allowed {
			return reject(fmt.Sprintf("card type %s is not allowed by policy", tx.Card.Type))
		}
	}

	// Rule 4 — daily spending cap
	if policy.MaxAmountPerDay > 0 && spentToday+tx.Amount > policy.MaxAmountPerDay {
		return reject(fmt.Sprintf("daily cap exceeded: %.2f spent + %.2f > %.2f",
			spentToday, tx.Amount, policy.MaxAmountPerDay))
	}

	// Rule 5 — monthly spending cap
	if policy.MaxAmountPerMonth > 0 && spentMonth+tx.Amount > policy.MaxAmountPerMonth {
		return reject(fmt.Sprintf("monthly cap exceeded: %.2f spent + %.2f > %.2f",
			spentMonth, tx.Amount, policy.MaxAmountPerMonth))
	}

	// all checks passed
	return service.ValidationResult{Status: model.StatusApproved, RejectReason: ""}
}

func reject(reason string) service.ValidationResult {
	return service.ValidationResult{Status: model.StatusRejected, RejectReason: reason}
}
