package transfer

import (
	"context"
	"database/sql"
)

//go:generate go run go.uber.org/mock/mockgen -source=repository.go -destination=../../mock/transfer_repository_mock.go -package=mock -mock_names=Repository=TransferRepositoryMock
type Repository interface {
	CreateBulkTransfers(ctx context.Context, tx *sql.Tx, transfers []Transfer) error
}
