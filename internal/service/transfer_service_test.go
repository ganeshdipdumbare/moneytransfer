package service_test

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"
	"time"

	"moneytransfer/internal/account"
	"moneytransfer/internal/service"
	"moneytransfer/internal/transfer"
	"moneytransfer/mock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// MockDB is a mock of sql.DB
type MockDB struct {
	BeginTxFunc func(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

func (m *MockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.BeginTxFunc(ctx, opts)
}

// Implement other necessary methods of sql.DB interface with empty bodies

func TestTransferService_BulkTransfer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAccountRepo := mock.NewAccountRepositoryMock(ctrl)
	mockTransferRepo := mock.NewTransferRepositoryMock(ctrl)

	// Create a new sqlmock database connection
	mockDB, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	logger := slog.Default()

	retryConfig := service.RetryConfig{
		BaseDelay:  time.Millisecond,
		MaxDelay:   time.Millisecond * 10,
		MaxRetries: 3,
	}

	svc := service.NewTransferService(mockDB, logger, mockAccountRepo, mockTransferRepo, retryConfig)

	t.Run("Successful bulk transfer", func(t *testing.T) {
		ctx := context.Background()
		req := service.BulkTransferRequest{
			OrganizationName: "Test Org",
			OrganizationBIC:  "TESTBIC",
			OrganizationIBAN: "TEST123456789",
			Transfers: []transfer.Transfer{
				{AmountCents: 1000},
				{AmountCents: 2000},
			},
		}

		// Expect a transaction to be started
		sqlMock.ExpectBegin()

		mockAccountRepo.EXPECT().GetByIBAN(req.OrganizationIBAN, gomock.Any()).Return(&account.BankAccount{
			ID:               1,
			BalanceCents:     5000,
			IBAN:             req.OrganizationIBAN,
			BIC:              req.OrganizationBIC,
			OrganizationName: req.OrganizationName,
		}, nil)

		mockTransferRepo.EXPECT().CreateBulkTransfers(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockAccountRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		// Expect the transaction to be committed
		sqlMock.ExpectCommit()

		err := svc.BulkTransfer(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Insufficient funds", func(t *testing.T) {
		ctx := context.Background()
		req := service.BulkTransferRequest{
			OrganizationName: "Test Org",
			OrganizationBIC:  "TESTBIC",
			OrganizationIBAN: "TEST123456789",
			Transfers: []transfer.Transfer{
				{AmountCents: 3000},
				{AmountCents: 3000},
			},
		}

		// Expect a transaction to be started
		sqlMock.ExpectBegin()

		mockAccountRepo.EXPECT().GetByIBAN(req.OrganizationIBAN, gomock.Any()).Return(&account.BankAccount{
			ID:               1,
			BalanceCents:     5000,
			IBAN:             req.OrganizationIBAN,
			BIC:              req.OrganizationBIC,
			OrganizationName: req.OrganizationName,
		}, nil)

		// Expect the transaction to be rolled back
		sqlMock.ExpectRollback()

		err := svc.BulkTransfer(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")
	})

	t.Run("Retryable error", func(t *testing.T) {
		ctx := context.Background()
		req := service.BulkTransferRequest{
			OrganizationName: "Test Org",
			OrganizationBIC:  "TESTBIC",
			OrganizationIBAN: "TEST123456789",
			Transfers: []transfer.Transfer{
				{AmountCents: 1000},
			},
		}

		retryableErr := errors.New("serialization failure")

		// Expect 3 transaction begins (initial + 2 retries)
		for i := 0; i < 3; i++ {
			sqlMock.ExpectBegin()
			mockAccountRepo.EXPECT().GetByIBAN(req.OrganizationIBAN, gomock.Any()).Return(&account.BankAccount{
				ID:               1,
				BalanceCents:     5000,
				IBAN:             req.OrganizationIBAN,
				BIC:              req.OrganizationBIC,
				OrganizationName: req.OrganizationName,
			}, nil)
			mockTransferRepo.EXPECT().CreateBulkTransfers(ctx, gomock.Any(), gomock.Any()).Return(retryableErr)
			sqlMock.ExpectRollback()
		}

		err := svc.BulkTransfer(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, retryableErr, err)
	})

	t.Run("Non-retryable error", func(t *testing.T) {
		ctx := context.Background()
		req := service.BulkTransferRequest{
			OrganizationName: "Test Org",
			OrganizationBIC:  "TESTBIC",
			OrganizationIBAN: "TEST123456789",
			Transfers: []transfer.Transfer{
				{AmountCents: 1000},
			},
		}

		nonRetryableErr := errors.New("non-retryable error")

		sqlMock.ExpectBegin()
		mockAccountRepo.EXPECT().GetByIBAN(req.OrganizationIBAN, gomock.Any()).Return(nil, nonRetryableErr)
		sqlMock.ExpectRollback()

		err := svc.BulkTransfer(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, nonRetryableErr, err)
	})

	// Ensure all expectations were met
	if err := sqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
