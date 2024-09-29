package transfer

import (
	"errors"

	"github.com/google/uuid"
)

type Transfer struct {
	ID               int64
	CounterpartyName string
	CounterpartyIBAN string
	CounterpartyBIC  string
	AmountCents      int64
	BankAccountID    int64
	Description      string
}

func NewTransfer(counterpartyName, counterpartyIBAN, counterpartyBIC string, amountCents, bankAccountID int64, description string) *Transfer {
	return &Transfer{
		ID:               int64(uuid.New().ID()),
		CounterpartyName: counterpartyName,
		CounterpartyIBAN: counterpartyIBAN,
		CounterpartyBIC:  counterpartyBIC,
		AmountCents:      amountCents,
		BankAccountID:    bankAccountID,
		Description:      description,
	}
}

func (t *Transfer) Validate() error {
	if t.CounterpartyName == "" {
		return errors.New("counterparty name is required")
	}
	if t.CounterpartyIBAN == "" {
		return errors.New("counterparty IBAN is required")
	}
	if t.CounterpartyBIC == "" {
		return errors.New("counterparty BIC is required")
	}
	if t.AmountCents <= 0 {
		return errors.New("amount is required")
	}
	if t.BankAccountID == 0 {
		return errors.New("bank account ID is required")
	}
	return nil
}
