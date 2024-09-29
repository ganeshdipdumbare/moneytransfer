package transfer

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresRepositoryTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer testcontainers.Container
	db          *sql.DB
	repo        Repository
}

func TestPostgresRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresRepositoryTestSuite))
}

func (s *PostgresRepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
	}

	var err error
	s.pgContainer, err = testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)

	host, err := s.pgContainer.Host(s.ctx)
	s.Require().NoError(err)

	port, err := s.pgContainer.MappedPort(s.ctx, "5432")
	s.Require().NoError(err)

	dbURL := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
	s.db, err = sql.Open("postgres", dbURL)
	s.Require().NoError(err)

	s.repo = NewPostgresRepository(s.db)

	err = s.db.Ping()
	s.Require().NoError(err)

	s.createSchema()
}

func (s *PostgresRepositoryTestSuite) TearDownSuite() {
	s.db.Close()
	s.pgContainer.Terminate(s.ctx)
}

func (s *PostgresRepositoryTestSuite) createSchema() {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS transfers (
			id SERIAL PRIMARY KEY,
			counterparty_name TEXT NOT NULL,
			counterparty_iban TEXT NOT NULL,
			counterparty_bic TEXT NOT NULL,
			amount_cents INTEGER NOT NULL,
			bank_account_id INTEGER NOT NULL,
			description TEXT
		)
	`)
	s.Require().NoError(err)
}

func (s *PostgresRepositoryTestSuite) TestCreateBulkTransfers() {
	// Test case 1: Successful bulk transfer creation
	s.Run("Successful bulk transfer creation", func() {
		transfers := []Transfer{
			{
				CounterpartyName: "John Doe",
				CounterpartyIBAN: "GB29NWBK60161331926819",
				CounterpartyBIC:  "NWBKGB2L",
				AmountCents:      10000,
				BankAccountID:    1,
				Description:      "Test transfer 1",
			},
			{
				CounterpartyName: "Jane Smith",
				CounterpartyIBAN: "DE89370400440532013000",
				CounterpartyBIC:  "DEUTDEFF",
				AmountCents:      20000,
				BankAccountID:    2,
				Description:      "Test transfer 2",
			},
		}

		tx, err := s.db.Begin()
		s.Require().NoError(err)

		err = s.repo.CreateBulkTransfers(s.ctx, tx, transfers)
		s.Require().NoError(err)

		err = tx.Commit()
		s.Require().NoError(err)

		// Verify the transfers were created
		var count int
		err = s.db.QueryRow("SELECT COUNT(*) FROM transfers").Scan(&count)
		s.Require().NoError(err)
		s.Equal(2, count)
	})
}
