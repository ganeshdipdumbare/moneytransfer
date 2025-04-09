package transfer

import (
	"errors"
	"math/rand"
	"time"
)

type Transfer struct {
	// ID is the unique identifier for the transfer
	// it is generated by the database
	ID               int64
	CounterpartyName string
	CounterpartyIBAN string
	CounterpartyBIC  string
	AmountCents      int64
	BankAccountID    int64
	Description      string
}

func NewTransfer(counterpartyName, counterpartyIBAN, counterpartyBIC string, amountCents, bankAccountID int64, description string) *Transfer {
	// Generate a simple unique ID based on timestamp and random number
	// In a real application, you might want to use a more robust ID generation method
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomID := r.Int63n(1000000) + 1 // Ensure ID is not zero

	return &Transfer{
		ID:               randomID,
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
