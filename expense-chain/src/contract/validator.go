package contract

import (
	"expense-chain/src/model"
	"expense-chain/src/service"
)

// Validator implements service.Validator interface.
// Smart contract logic goes here — for now approves everything.
type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(tx *model.Transaction, policy *model.Policy) service.ValidationResult {
	// TODO: implement policy validation logic
	return service.ValidationResult{
		Status:       model.StatusApproved,
		RejectReason: "",
	}
}
