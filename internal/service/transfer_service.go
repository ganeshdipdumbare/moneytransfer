package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"math"
	"strings"
	"time"

	"moneytransfer/internal/account"
	"moneytransfer/internal/tools"
	"moneytransfer/internal/transfer"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

type BulkTransferRequest struct {
	OrganizationName string
	OrganizationBIC  string
	OrganizationIBAN string
	Transfers        []transfer.Transfer
}

//go:generate go run go.uber.org/mock/mockgen -source=transfer_service.go -destination=../../mock/service_mock.go -package=mock -mock_names=TransferService=TransferServiceMock
type TransferService interface {
	BulkTransfer(ctx context.Context, req BulkTransferRequest) error
}

// RetryConfig is a struct that contains the retry configuration for the transfer service
type RetryConfig struct {
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	MaxRetries int
}

type transferService struct {
	accountRepo  account.Repository
	transferRepo transfer.Repository
	db           *sql.DB
	logger       *slog.Logger
	retryConfig  RetryConfig
}

// NewTransferService is a function that creates a new transfer service
func NewTransferService(db *sql.DB, logger *slog.Logger, accountRepo account.Repository, transferRepo transfer.Repository, retryConfig RetryConfig) *transferService {
	return &transferService{
		accountRepo:  accountRepo,
		transferRepo: transferRepo,
		db:           db,
		logger:       logger,
		retryConfig:  retryConfig,
	}
}

// BulkTransfer is a function that processes a bulk transfer request
// It retries the transfer if the database transaction fails due to serialization conflicts
// It returns an error if the transfer fails after the maximum number of retries
func (s *transferService) BulkTransfer(ctx context.Context, req BulkTransferRequest) error {
	s.logger.Info("Processing bulk transfer request", "organization_bic", req.OrganizationBIC, "organization_iban", req.OrganizationIBAN)

	var err error

	for attempt := 0; attempt < s.retryConfig.MaxRetries; attempt++ {
		err = s.executeBulkTransfer(ctx, req)
		if err == nil {
			s.logger.Info("Bulk transfer processed successfully", "organization_bic", req.OrganizationBIC, "organization_iban", req.OrganizationIBAN, "attempt", attempt+1)
			return nil
		}

		if !isRetryableError(err) {
			s.logger.Error("Non-retryable error occurred during bulk transfer", "error", err, "attempt", attempt+1)
			return err
		}

		if attempt == s.retryConfig.MaxRetries-1 {
			break
		}

		delay := tools.CalculateBackoff(s.retryConfig.BaseDelay, s.retryConfig.MaxDelay, attempt)
		s.logger.Warn("Retryable error occurred during bulk transfer, retrying", "error", err, "attempt", attempt+1, "retry_after", delay)

		select {
		case <-time.After(delay):
			// Continue with the next iteration
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	s.logger.Error("Failed to process bulk transfer after maximum retries", "max_retries", s.retryConfig.MaxRetries)
	return err
}

func (s *transferService) executeBulkTransfer(ctx context.Context, req BulkTransferRequest) error {
	s.logger.Info("Processing bulk transfer request", "organization_bic", req.OrganizationBIC, "organization_iban", req.OrganizationIBAN)

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	account, err := s.accountRepo.GetByIBAN(req.OrganizationIBAN, tx)
	if err != nil {
		s.logger.Error("Failed to get bank account", "error", err)
		return err
	}

	totalTransfer, err := calculateTotalTransfer(req.Transfers)
	if err != nil {
		s.logger.Error("Failed to calculate total transfer", "error", err)
		return err
	}

	s.logger.Debug("Transfer details", "total_transfer", totalTransfer, "account_balance", account.BalanceCents)

	if account.BalanceCents < totalTransfer {
		s.logger.Warn("Insufficient funds", "required", totalTransfer, "available", account.BalanceCents)
		return ErrInsufficientFunds
	}

	transfersList := make([]transfer.Transfer, len(req.Transfers))
	for i, ct := range req.Transfers {
		transfersList[i] = *transfer.NewTransfer(
			ct.CounterpartyName,
			ct.CounterpartyIBAN,
			ct.CounterpartyBIC,
			ct.AmountCents,
			account.ID,
			ct.Description,
		)
	}

	err = s.transferRepo.CreateBulkTransfers(ctx, tx, transfersList)
	if err != nil {
		s.logger.Error("Failed to create transfers", "error", err)
		return err
	}

	// Update the account balance
	account.BalanceCents = account.BalanceCents - totalTransfer
	err = s.accountRepo.Update(account, tx)
	if err != nil {
		s.logger.Error("Failed to update account balance", "error", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction", "error", err)
		return err
	}

	s.logger.Info("Bulk transfer processed successfully", "organization_bic", req.OrganizationBIC, "organization_iban", req.OrganizationIBAN, "total_transfer", totalTransfer)
	return nil
}

func calculateTotalTransfer(transfers []transfer.Transfer) (int64, error) {
	var total int64
	for _, t := range transfers {
		// Check for potential overflow
		if t.AmountCents < 0 {
			return 0, errors.New("negative transfer amount not allowed")
		}
		if total > math.MaxInt64-t.AmountCents {
			return 0, errors.New("total transfer amount exceeds maximum allowed value")
		}
		total += t.AmountCents
	}
	return total, nil
}

func isRetryableError(err error) bool {
	// Check if the error is due to serializable isolation conflicts
	// This might need to be adjusted based on the specific error returned by your database
	return errors.Is(err, sql.ErrTxDone) || errors.Is(err, sql.ErrConnDone) || strings.Contains(err.Error(), "serialization failure")
}
