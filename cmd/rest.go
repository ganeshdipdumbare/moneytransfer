/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"
	"os"

	"moneytransfer/config"
	"moneytransfer/internal/account"
	"moneytransfer/internal/api/rest"
	"moneytransfer/internal/infra"
	"moneytransfer/internal/service"
	"moneytransfer/internal/transfer"

	"github.com/spf13/cobra"
)

// @title Money Transfer API
// @version 1.0
// @description This is a money transfer service API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// restCmd represents the rest command
var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "Start the REST API server",
	Long: `This command initializes and starts the REST API server.
It sets up the necessary routes and listens for incoming HTTP requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load the configuration
		config, err := config.LoadConfig()
		if err != nil {
			os.Exit(1)
		}

		// Create logger instance
		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: config.LogLevel}))

		// Create DB instance
		db, err := infra.NewDatabase(config.DatabaseURL, logger)
		if err != nil {
			logger.Error("failed to create new database", slog.Any("error", err))
			os.Exit(1)
		}

		// Create account and transaction repositories
		accountRepo := account.NewPostgresRepository(db)
		transferRepo := transfer.NewPostgresRepository(db)

		// Create account and transaction services
		retryConfig := service.RetryConfig{
			BaseDelay:  config.RetryConfig.BaseDelay,
			MaxDelay:   config.RetryConfig.MaxDelay,
			MaxRetries: config.RetryConfig.MaxRetries,
		}
		transferService := service.NewTransferService(db, logger, accountRepo, transferRepo, retryConfig)

		// create a new rest api instance
		api, err := rest.NewApi(transferService, config.ServerPort)
		if err != nil {
			logger.Error("failed to create new rest api", slog.Any("error", err))
			os.Exit(1)
		}

		// start the rest server
		api.StartServer()
	},
}

func init() {
	rootCmd.AddCommand(restCmd)
}
