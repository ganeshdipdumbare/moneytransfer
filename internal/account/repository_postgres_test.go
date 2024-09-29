package account

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RepositoryTestSuite struct {
	suite.Suite
	db          *sql.DB
	pgContainer testcontainers.Container
	repo        Repository
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (s *RepositoryTestSuite) SetupSuite() {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithStartupTimeout(time.Minute),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(s.T(), err)

	s.pgContainer = pgContainer

	mappedPort, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(s.T(), err)

	host, err := pgContainer.Host(ctx)
	require.NoError(s.T(), err)

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, mappedPort.Port())

	// Retry connection logic
	var dbErr error
	for i := 0; i < 5; i++ {
		s.db, dbErr = sql.Open("postgres", dsn)
		if dbErr == nil {
			err = s.db.Ping()
			if err == nil {
				break
			}
		}
		s.T().Logf("Failed to connect to database, retrying in 2 seconds... (attempt %d/5)", i+1)
		time.Sleep(2 * time.Second)
	}
	require.NoError(s.T(), dbErr)
	require.NoError(s.T(), err, "Failed to ping database after 5 attempts")

	// Create the bank_accounts table
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS bank_accounts (
			id SERIAL PRIMARY KEY,
			organization_name TEXT NOT NULL,
			balance_cents BIGINT NOT NULL,
			iban TEXT NOT NULL UNIQUE,
			bic TEXT NOT NULL
		)
	`)
	require.NoError(s.T(), err)

	s.repo = NewPostgresRepository(s.db)
}

func (s *RepositoryTestSuite) TearDownSuite() {
	s.db.Close()
	s.pgContainer.Terminate(context.Background())
}

func (s *RepositoryTestSuite) SetupTest() {
	// Clear the table before each test
	_, err := s.db.Exec("DELETE FROM bank_accounts")
	require.NoError(s.T(), err)
}

func (s *RepositoryTestSuite) TestNewPostgresRepository() {
	s.NotNil(s.repo)
	s.IsType(&bankAccountPostgresRepository{}, s.repo)
}

func (s *RepositoryTestSuite) TestCreate() {
	account := &BankAccount{
		OrganizationName: "Test Org",
		BalanceCents:     10000,
		IBAN:             "NL91ABNA0417164300",
		BIC:              "ABNANL2A",
	}

	accountCreated, err := s.repo.Create(account, nil)
	s.NoError(err)

	// Verify the account was created
	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM bank_accounts WHERE iban = $1", account.IBAN).Scan(&count)
	s.NoError(err)
	s.Equal(1, count)

	// Add this assertion to verify that the ID is set
	s.NotZero(accountCreated.ID)
}

func (s *RepositoryTestSuite) TestGet() {
	account := &BankAccount{
		OrganizationName: "Test Org",
		BalanceCents:     10000,
		IBAN:             "NL91ABNA0417164300",
		BIC:              "ABNANL2A",
	}
	_, err := s.repo.Create(account, nil)
	s.Require().NoError(err)

	// Debug: Check if the account was actually created
	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM bank_accounts WHERE iban = $1", account.IBAN).Scan(&count)
	s.Require().NoError(err)
	s.T().Logf("Number of accounts with IBAN %s: %d", account.IBAN, count)

	// Debug: Get the ID of the created account
	var id int64
	err = s.db.QueryRow("SELECT id FROM bank_accounts WHERE iban = $1", account.IBAN).Scan(&id)
	s.Require().NoError(err)
	s.T().Logf("Created account ID: %d", id)

	// Now try to retrieve the account
	retrievedAccount, err := s.repo.Get(id, nil)
	if err != nil {
		s.T().Logf("Error retrieving account: %v", err)
	}
	s.Require().NoError(err)
	s.Require().NotNil(retrievedAccount)

	s.Equal(account.OrganizationName, retrievedAccount.OrganizationName)
	s.Equal(account.BalanceCents, retrievedAccount.BalanceCents)
	s.Equal(account.IBAN, retrievedAccount.IBAN)
	s.Equal(account.BIC, retrievedAccount.BIC)
}

func (s *RepositoryTestSuite) TestGetByIBAN() {
	account := &BankAccount{
		OrganizationName: "Test Org",
		BalanceCents:     10000,
		IBAN:             "NL91ABNA0417164300",
		BIC:              "ABNANL2A",
	}
	_, err := s.repo.Create(account, nil)
	s.Require().NoError(err)

	retrievedAccount, err := s.repo.GetByIBAN(account.IBAN, nil)
	s.NoError(err)
	s.NotNil(retrievedAccount)
	s.Equal(account.OrganizationName, retrievedAccount.OrganizationName)
	s.Equal(account.BalanceCents, retrievedAccount.BalanceCents)
	s.Equal(account.IBAN, retrievedAccount.IBAN)
	s.Equal(account.BIC, retrievedAccount.BIC)
}

func (s *RepositoryTestSuite) TestUpdate() {
	// Create initial account
	account := &BankAccount{
		OrganizationName: "Test Org",
		BalanceCents:     10000,
		IBAN:             "NL91ABNA0417164300",
		BIC:              "ABNANL2A",
	}
	accountCreated, err := s.repo.Create(account, nil)
	s.Require().NoError(err)

	// Update the account
	accountCreated.OrganizationName = "Updated Org"
	accountCreated.BalanceCents = 20000
	err = s.repo.Update(accountCreated, nil)
	s.Require().NoError(err)

	// Retrieve the updated account
	updatedAccount, err := s.repo.GetByIBAN(accountCreated.IBAN, nil)
	s.Require().NoError(err)
	s.Require().NotNil(updatedAccount)

	// Assertions
	s.Equal("Updated Org", updatedAccount.OrganizationName, "Organization name was not updated")
	s.Equal(int64(20000), updatedAccount.BalanceCents, "Balance was not updated")
}

func (s *RepositoryTestSuite) TestDelete() {
	account := &BankAccount{
		OrganizationName: "Test Org",
		BalanceCents:     10000,
		IBAN:             "NL91ABNA0417164300",
		BIC:              "ABNANL2A",
	}
	_, err := s.repo.Create(account, nil)
	s.Require().NoError(err)

	err = s.repo.Delete(account.ID, nil)
	s.NoError(err)

	_, err = s.repo.Get(account.ID, nil)
	s.Error(err)
	s.Contains(err.Error(), "bank account not found")
}

func (s *RepositoryTestSuite) TestGet_NotFound() {
	_, err := s.repo.Get(9999, nil)
	s.Error(err)
	s.Contains(err.Error(), "bank account not found")
}

func (s *RepositoryTestSuite) TestGetByIBAN_NotFound() {
	_, err := s.repo.GetByIBAN("NON_EXISTENT_IBAN", nil)
	s.Error(err)
	s.Contains(err.Error(), "bank account not found")
}
