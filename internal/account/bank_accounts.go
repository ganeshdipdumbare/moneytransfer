package account

import (
	"errors"

	"github.com/google/uuid"
)

type BankAccount struct {
	ID               int64
	OrganizationName string
	BalanceCents     int64
	IBAN             string
	BIC              string
}

func NewBankAccount(organizationName string, balanceCents int64, iban string, bic string) *BankAccount {
	return &BankAccount{
		ID:               int64(uuid.New().ID()),
		OrganizationName: organizationName,
		BalanceCents:     balanceCents,
		IBAN:             iban,
		BIC:              bic,
	}
}

func (b *BankAccount) Validate() error {
	if b.OrganizationName == "" {
		return errors.New("organization name is required")
	}
	if b.IBAN == "" {
		return errors.New("iban is required")
	}
	if b.BIC == "" {
		return errors.New("bic is required")
	}
	return nil
}
