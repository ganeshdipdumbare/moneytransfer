package account

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBankAccount(t *testing.T) {
	ba := NewBankAccount("Test Org", 10000, "NL91ABNA0417164300", "ABNANL2A")

	assert.NotNil(t, ba)
	assert.NotZero(t, ba.ID)
	assert.Equal(t, "Test Org", ba.OrganizationName)
	assert.Equal(t, int64(10000), ba.BalanceCents)
	assert.Equal(t, "NL91ABNA0417164300", ba.IBAN)
	assert.Equal(t, "ABNANL2A", ba.BIC)
}

func TestBankAccount_Validate(t *testing.T) {
	tests := []struct {
		name    string
		account *BankAccount
		wantErr string
	}{
		{
			name: "Valid bank account",
			account: &BankAccount{
				OrganizationName: "Test Org",
				IBAN:             "NL91ABNA0417164300",
				BIC:              "ABNANL2A",
			},
			wantErr: "",
		},
		{
			name: "Missing organization name",
			account: &BankAccount{
				IBAN: "NL91ABNA0417164300",
				BIC:  "ABNANL2A",
			},
			wantErr: "organization name is required",
		},
		{
			name: "Missing IBAN",
			account: &BankAccount{
				OrganizationName: "Test Org",
				BIC:              "ABNANL2A",
			},
			wantErr: "iban is required",
		},
		{
			name: "Missing BIC",
			account: &BankAccount{
				OrganizationName: "Test Org",
				IBAN:             "NL91ABNA0417164300",
			},
			wantErr: "bic is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.account.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
