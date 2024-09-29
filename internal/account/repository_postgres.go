package account

import (
	"database/sql"
	"errors"
	"fmt"
)

type bankAccountPostgresRepository struct {
	db *sql.DB
}

type bankAccountModel struct {
	ID               int64
	OrganizationName string
	BalanceCents     int64
	IBAN             string
	BIC              string
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &bankAccountPostgresRepository{db: db}
}

func (r *bankAccountPostgresRepository) Create(account *BankAccount, tx *sql.Tx) (*BankAccount, error) {
	query := `
		INSERT INTO bank_accounts (organization_name, balance_cents, iban, bic)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, account.OrganizationName, account.BalanceCents, account.IBAN, account.BIC)
	} else {
		row = r.db.QueryRow(query, account.OrganizationName, account.BalanceCents, account.IBAN, account.BIC)
	}

	err := row.Scan(&account.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create bank account: %w", err)
	}

	return account, nil
}

func (r *bankAccountPostgresRepository) Get(id int64, tx *sql.Tx) (*BankAccount, error) {
	query := `
		SELECT id, organization_name, balance_cents, iban, bic
		FROM bank_accounts
		WHERE id = $1
	`

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, id)
	} else {
		row = r.db.QueryRow(query, id)
	}

	var model bankAccountModel
	err := row.Scan(&model.ID, &model.OrganizationName, &model.BalanceCents, &model.IBAN, &model.BIC)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("bank account not found")
		}
		return nil, err
	}

	return toAccount(model), nil
}

func (r *bankAccountPostgresRepository) Update(account *BankAccount, tx *sql.Tx) error {
	query := `
		UPDATE bank_accounts
		SET organization_name = $1, balance_cents = $2, iban = $3, bic = $4
		WHERE id = $5
	`

	var err error
	if tx != nil {
		_, err = tx.Exec(query, account.OrganizationName, account.BalanceCents, account.IBAN, account.BIC, account.ID)
	} else {
		_, err = r.db.Exec(query, account.OrganizationName, account.BalanceCents, account.IBAN, account.BIC, account.ID)
	}

	return err
}

func (r *bankAccountPostgresRepository) Delete(id int64, tx *sql.Tx) error {
	query := `
		DELETE FROM bank_accounts
		WHERE id = $1
	`

	var err error
	if tx != nil {
		_, err = tx.Exec(query, id)
	} else {
		_, err = r.db.Exec(query, id)
	}

	return err
}

func (r *bankAccountPostgresRepository) GetByIBAN(iban string, tx *sql.Tx) (*BankAccount, error) {
	query := `
		SELECT id, organization_name, balance_cents, iban, bic
		FROM bank_accounts
		WHERE iban = $1
	`

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, iban)
	} else {
		row = r.db.QueryRow(query, iban)
	}

	var model bankAccountModel
	err := row.Scan(&model.ID, &model.OrganizationName, &model.BalanceCents, &model.IBAN, &model.BIC)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("bank account not found")
		}
		return nil, err
	}

	return toAccount(model), nil
}

func toAccount(model bankAccountModel) *BankAccount {
	return &BankAccount{
		ID:               model.ID,
		OrganizationName: model.OrganizationName,
		BalanceCents:     model.BalanceCents,
		IBAN:             model.IBAN,
		BIC:              model.BIC,
	}
}
