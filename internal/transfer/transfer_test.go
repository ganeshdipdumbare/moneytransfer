package transfer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTransfer(t *testing.T) {
	transfer := NewTransfer("John Doe", "DE89370400440532013000", "DEUTDEFF", 10000, 1, "Test transfer")

	assert.NotZero(t, transfer.ID)
	assert.Equal(t, "John Doe", transfer.CounterpartyName)
	assert.Equal(t, "DE89370400440532013000", transfer.CounterpartyIBAN)
	assert.Equal(t, "DEUTDEFF", transfer.CounterpartyBIC)
	assert.Equal(t, int64(10000), transfer.AmountCents)
	assert.Equal(t, int64(1), transfer.BankAccountID)
	assert.Equal(t, "Test transfer", transfer.Description)
}

func TestTransfer_Validate(t *testing.T) {
	tests := []struct {
		name     string
		transfer Transfer
		wantErr  string
	}{
		{
			name: "Valid transfer",
			transfer: Transfer{
				CounterpartyName: "John Doe",
				CounterpartyIBAN: "DE89370400440532013000",
				CounterpartyBIC:  "DEUTDEFF",
				AmountCents:      10000,
				BankAccountID:    1,
				Description:      "Test transfer",
			},
			wantErr: "",
		},
		{
			name: "Missing counterparty name",
			transfer: Transfer{
				CounterpartyIBAN: "DE89370400440532013000",
				CounterpartyBIC:  "DEUTDEFF",
				AmountCents:      10000,
				BankAccountID:    1,
			},
			wantErr: "counterparty name is required",
		},
		{
			name: "Missing counterparty IBAN",
			transfer: Transfer{
				CounterpartyName: "John Doe",
				CounterpartyBIC:  "DEUTDEFF",
				AmountCents:      10000,
				BankAccountID:    1,
			},
			wantErr: "counterparty IBAN is required",
		},
		{
			name: "Missing counterparty BIC",
			transfer: Transfer{
				CounterpartyName: "John Doe",
				CounterpartyIBAN: "DE89370400440532013000",
				AmountCents:      10000,
				BankAccountID:    1,
			},
			wantErr: "counterparty BIC is required",
		},
		{
			name: "Invalid amount",
			transfer: Transfer{
				CounterpartyName: "John Doe",
				CounterpartyIBAN: "DE89370400440532013000",
				CounterpartyBIC:  "DEUTDEFF",
				AmountCents:      0,
				BankAccountID:    1,
			},
			wantErr: "amount is required",
		},
		{
			name: "Missing bank account ID",
			transfer: Transfer{
				CounterpartyName: "John Doe",
				CounterpartyIBAN: "DE89370400440532013000",
				CounterpartyBIC:  "DEUTDEFF",
				AmountCents:      10000,
			},
			wantErr: "bank account ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.transfer.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
