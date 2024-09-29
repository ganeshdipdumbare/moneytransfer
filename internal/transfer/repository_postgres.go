package transfer

import (
	"context"
	"database/sql"
	"fmt"
)

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateBulkTransfers(ctx context.Context, tx *sql.Tx, transfers []Transfer) error {
	query := `
		INSERT INTO transfers (counterparty_name, counterparty_iban, counterparty_bic, amount_cents, bank_account_id, description)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, transfer := range transfers {
		_, err := stmt.ExecContext(ctx,
			transfer.CounterpartyName,
			transfer.CounterpartyIBAN,
			transfer.CounterpartyBIC,
			transfer.AmountCents,
			transfer.BankAccountID,
			transfer.Description,
		)
		if err != nil {
			return fmt.Errorf("failed to insert transfer: %w", err)
		}
	}

	return nil
}
