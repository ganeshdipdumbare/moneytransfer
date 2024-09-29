package infra

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func NewDatabase(dbPath string, logger *slog.Logger) (*sql.DB, error) {
	// Parse the connection string
	u, err := url.Parse(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %v", err)
	}

	// Extract query parameters
	q := u.Query()

	// Extract dbname from query parameters
	dbName := q.Get("dbname")
	if dbName == "" {
		dbName = "qonto_accounts"
	}
	q.Del("dbname")

	// Ensure sslmode is set correctly
	if q.Get("sslmode") == "" {
		q.Set("sslmode", "disable")
	}

	// Reconstruct the connection string for the 'postgres' database
	u.Path = "/postgres"
	u.RawQuery = q.Encode()
	postgresDBPath := u.String()

	// Connect to the 'postgres' database
	postgresDB, err := sql.Open("postgres", postgresDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres database: %v", err)
	}
	defer postgresDB.Close()

	// Check if the target database exists
	var exists bool
	err = postgresDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database exists: %v", err)
	}

	// If the database doesn't exist, create it
	if !exists {
		_, err = postgresDB.Exec("CREATE DATABASE " + pq.QuoteIdentifier(dbName))
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %v", err)
		}
		logger.Info(fmt.Sprintf("Created '%s' database", dbName))
	}

	// Reconstruct the connection string for the target database
	u.Path = "/" + dbName
	u.RawQuery = q.Encode()
	targetDBPath := u.String()

	// Connect to the target database with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := sql.Open("postgres", targetDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %w", dbName, err)
	}

	// Set connection pool settings
	// TODO: Make these configurable via env variables
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err = runMigrations(db, logger); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB, logger *slog.Logger) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create driver: %v", err)
	}

	// Ensure the path to your migrations is correct
	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations", // Update this path if necessary
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %v", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			logger.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("could not run migrations: %w", err)
	}

	logger.Info("Database migrations completed successfully")
	return nil
}
