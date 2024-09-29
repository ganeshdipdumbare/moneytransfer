package account

import (
	"database/sql"
)

//go:generate go run go.uber.org/mock/mockgen -source=repository.go -destination=../../mock/account_repository_mock.go -package=mock -mock_names=Repository=AccountRepositoryMock
type Repository interface {
	Create(acc *BankAccount, tx *sql.Tx) (*BankAccount, error)
	Get(id int64, tx *sql.Tx) (*BankAccount, error)
	GetByIBAN(iban string, tx *sql.Tx) (*BankAccount, error)
	Update(acc *BankAccount, tx *sql.Tx) error
	Delete(id int64, tx *sql.Tx) error
}
