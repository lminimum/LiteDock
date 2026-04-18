// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lminimum/LiteDock/config"
	"github.com/lminimum/LiteDock/internal/controller/restapi"
	"github.com/lminimum/LiteDock/internal/repo/persistent"
	"github.com/lminimum/LiteDock/internal/usecase/auth"
	"github.com/lminimum/LiteDock/internal/usecase/container"
	"github.com/lminimum/LiteDock/pkg/database"
	"github.com/lminimum/LiteDock/pkg/httpserver"
	"github.com/lminimum/LiteDock/pkg/logger"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	if err := AutoMigrate(cfg); err != nil {
		l.Error(fmt.Errorf("app - Run - AutoMigrate: %w", err))
	}

	// Repository - use database factory for multi-database support
	db, err := database.New()
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - database.New: %w", err))
	}
	defer db.Close()

	// Repository instances
	userRepo := persistent.NewUserRepo(db)
	containerRepo := persistent.NewContainerRepo(db)

	// Auth UseCase
	authUseCase := auth.New(userRepo, l)

	// Container UseCase (placeholder for Docker management)
	containerUseCase := container.New(containerRepo, l)

	// HTTP Server
	httpServer := httpserver.New(l, httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
	restapi.NewRouter(httpServer.App, cfg, containerUseCase, authUseCase, l)

	// Start servers
	httpServer.Start()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: %s", s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
