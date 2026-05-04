// Package app configures and runs application.
package app

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lminimum/LiteDock/config"
	"github.com/lminimum/LiteDock/internal/controller/restapi"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo/persistent"
	"github.com/lminimum/LiteDock/internal/usecase/auth"
	"github.com/lminimum/LiteDock/internal/usecase/container"
	"github.com/lminimum/LiteDock/internal/usecase/image"
	"github.com/lminimum/LiteDock/internal/usecase/network"
	"github.com/lminimum/LiteDock/internal/usecase/remote_machine"
	"github.com/lminimum/LiteDock/internal/usecase/volume"
	"github.com/lminimum/LiteDock/pkg/collector"
	"github.com/lminimum/LiteDock/pkg/database"
	apperrors "github.com/lminimum/LiteDock/pkg/errors"
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
	remoteMachineRepo := persistent.NewRemoteMachineRepo(db)
	systemMetricsRepo := persistent.NewSystemMetricsRepo(db)

	// Auto-create local machine if not exists
	initLocalMachine(l, remoteMachineRepo)

	// Auth UseCase
	authUseCase := auth.New(userRepo, l, cfg.Auth.JWTSecret)

	// Container UseCase (placeholder for Docker management)
	containerUseCase := container.New(containerRepo, l)

	// Network UseCase
	networkRepo := persistent.NewNetworkRepo(db)
	networkUseCase := network.New(networkRepo, remoteMachineRepo, cfg.Cache.ContainerTTL, l)

	// Volume UseCase
	volumeRepo := persistent.NewVolumeRepo(db)
	volumeUseCase := volume.New(volumeRepo, remoteMachineRepo, cfg.Cache.ContainerTTL, l)

	// Image UseCase
	imageRepo := persistent.NewImageRepo(db)
	imageUseCase := image.NewImageUseCase(imageRepo, remoteMachineRepo, cfg.Cache.ContainerTTL, l)

	// RemoteMachine UseCase
	remoteMachineUseCase := remote_machine.New(remoteMachineRepo, containerRepo, cfg.Cache.ContainerTTL, l)

	metricsCollector := collector.NewMetricsCollector(systemMetricsRepo, l, 2*time.Second)
	go metricsCollector.Start()

	// HTTP Server
	httpServer := httpserver.New(l, httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
	dashboardHandler := restapi.NewRouter(httpServer.App, cfg, containerUseCase, authUseCase, remoteMachineUseCase, systemMetricsRepo, networkUseCase, volumeUseCase, imageUseCase, l)

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
	metricsCollector.Stop()
	dashboardHandler.CloseAllConnections()
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}

func initLocalMachine(l logger.Interface, machineRepo *persistent.RemoteMachineRepo) {
	ctx := context.Background()

	_, err := machineRepo.GetByID(ctx, remote_machine.LocalMachineID)
	if err == nil {
		l.Debug("app - initLocalMachine: local machine already exists")
		return
	}

	if !stderrors.Is(err, apperrors.ErrRemoteMachineNotFound) {
		l.Warn("app - initLocalMachine: failed to check local machine: %v", err)
		return
	}

	localMachine := &entity.RemoteMachine{
		ID:         remote_machine.LocalMachineID,
		Name:       "本机",
		Host:       remote_machine.LocalMachineHost,
		Port:       0,
		Username:   "local",
		AuthMethod: entity.AuthMethodPassword,
		DockerHost: "/var/run/docker.sock",
		Status:     "unknown",
	}

	err = machineRepo.Create(ctx, localMachine)
	if err != nil {
		l.Warn("app - initLocalMachine: failed to create local machine: %v", err)
		return
	}

	l.Info("app - initLocalMachine: created local machine (ID: %s)", remote_machine.LocalMachineID)
}
