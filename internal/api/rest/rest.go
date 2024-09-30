package rest

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "moneytransfer/docs" // This is important!
	"moneytransfer/internal/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	nilArgErr   = "nil %v not allowed"
	emptyArgErr = "empty %v not allowed"
)

// @title Money Transfer API
// @version 1.0
// @description This is a money transfer service API.
// @termsOfService http://swagger.io/terms/

// @contact.name Ganeshdip Dumbare
// @contact.url https://github.com/ganeshdipdumbare
// @contact.email ganeshdip.dumbare@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// RestApi defines methods to handle rest server
type RestApi interface {
	StartServer()
	GracefulStopServer()
}

type apiDetails struct {
	service service.TransferService
	server  *http.Server
	logger  *slog.Logger
}

// NewApi creates new api instance, otherwise returns error
func NewApi(logger *slog.Logger, a service.TransferService, port string) (RestApi, error) {
	if logger == nil {
		return nil, fmt.Errorf(nilArgErr, "logger")
	}

	if a == nil {
		return nil, fmt.Errorf(nilArgErr, "app")
	}

	if port == "" {
		return nil, fmt.Errorf(emptyArgErr, "port")
	}

	api := &apiDetails{
		service: a,
		logger:  logger,
	}

	router := api.setupRouter()
	api.server = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%v", port),
		Handler: router,
	}

	// Add Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return api, nil
}

// StartServer starts the rest server
// it listens for a kill signal to stop the server gracefully
func (a *apiDetails) StartServer() {
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutting down server...")
	a.GracefulStopServer()
}

// GracefulStopServer stops the rest server gracefully
func (a *apiDetails) GracefulStopServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Server forced to shutdown:", err)
	}
	a.logger.Info("Server exiting")
}
