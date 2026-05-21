// Package app configures and runs application.
package app

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/lminimum/LiteDock/config"
	"github.com/lminimum/LiteDock/internal/action"
	"github.com/lminimum/LiteDock/internal/controller/restapi"
	v1 "github.com/lminimum/LiteDock/internal/controller/restapi/v1"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo/persistent"
	"github.com/lminimum/LiteDock/internal/usecase/auth"
	composeUseCase "github.com/lminimum/LiteDock/internal/usecase/compose"
	"github.com/lminimum/LiteDock/internal/usecase/container"
	"github.com/lminimum/LiteDock/internal/usecase/image"
	"github.com/lminimum/LiteDock/internal/usecase/network"
	"github.com/lminimum/LiteDock/internal/usecase/remote_machine"
	"github.com/lminimum/LiteDock/internal/usecase/task"
	"github.com/lminimum/LiteDock/internal/usecase/volume"
	"github.com/lminimum/LiteDock/internal/usecase/assistant"
	assistant_engine "github.com/lminimum/LiteDock/pkg/assistant/engine"
	assistant_rules "github.com/lminimum/LiteDock/pkg/assistant/rules"
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
	taskRepo := persistent.NewTaskRepo(db)
	systemMetricsRepo := persistent.NewSystemMetricsRepo(db)

	// Auto-create local machine if not exists
	initLocalMachine(l, remoteMachineRepo)

	// Task UseCase
	taskUseCase := task.New(taskRepo, l)

	authUseCase := auth.New(userRepo, l, cfg.Auth.JWTSecret)

	// Container UseCase (placeholder for Docker management)
	containerUseCase := container.New(containerRepo, remoteMachineRepo, l)

	// Network UseCase
	networkRepo := persistent.NewNetworkRepo(db)
	networkUseCase := network.New(networkRepo, remoteMachineRepo, cfg.Cache.ContainerTTL, l)

	// Volume UseCase
	volumeRepo := persistent.NewVolumeRepo(db)
	volumeUseCase := volume.New(volumeRepo, remoteMachineRepo, cfg.Cache.ContainerTTL, l)

	// Image UseCase
	imageRepo := persistent.NewImageRepo(db)
	imageUseCase := image.NewImageUseCase(imageRepo, remoteMachineRepo, cfg.Cache.ContainerTTL, l)

	// Compose UseCase
	composeRepo := persistent.NewComposeRepo(db)
	composeDir := cfg.App.ComposeDir
	if strings.HasPrefix(composeDir, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			composeDir = filepath.Join(home, composeDir[2:])
		}
	}
	composeUseCase := composeUseCase.NewComposeUseCase(composeRepo, remoteMachineRepo, cfg.Cache.ContainerTTL, composeDir, l)

	// RemoteMachine UseCase
	remoteMachineUseCase := remote_machine.New(remoteMachineRepo, containerRepo, taskUseCase, cfg.Cache.ContainerTTL, l)

	metricsCollector := collector.NewMetricsCollector(systemMetricsRepo, l, 2*time.Second)
	go metricsCollector.Start()

	// Assistant UseCase - NL parsing, fault diagnosis, config recommendations
	tokenizer, err := assistant.NewNLParserTokenizer()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - assistant.NewNLParserTokenizer: %w", err))
	} else {
		defer tokenizer.Close()
	}

	rulesLoader := assistant_rules.NewLoader()
	nlRules, err := rulesLoader.LoadNLRules("config/rules/nl_parsing.yaml")
	if err != nil {
		l.Error(fmt.Errorf("app - Run - load NL rules: %w", err))
	}

	engineRules := make([]assistant_engine.Rule, 0, len(nlRules))
	for _, r := range nlRules {
		engineRules = append(engineRules, assistant_engine.Rule{
			Name:        r.Name,
			Patterns:    r.Keywords,
			Intent:      r.Intent,
			Action:      r.Name,
			Description: r.Description,
		})
	}

	nlEngine := assistant_engine.NewEngine(engineRules, tokenizer)
	nlParserUseCase := assistant.NewNLParserUseCase(nlEngine, tokenizer, l)
	defer nlParserUseCase.Close()
	faultDiagnosisUseCase := assistant.NewFaultDiagnosisUseCase(l)
	configRecommendUseCase := assistant.NewConfigRecommendUseCase(l)

	settingsStore := v1.NewAISettingsStore(cfg.AI.APIEndpoint, cfg.AI.APIKey, cfg.AI.ModelName)

	// Action registry - register AI-callable operations
	actionRegistry := action.NewActionRegistry()
	if err := actionRegistry.Register(action.NewContainerAction(containerUseCase)); err != nil {
		l.Error(fmt.Errorf("app - Run - register container action: %w", err))
	}
	if err := actionRegistry.Register(action.NewImageAction(imageUseCase)); err != nil {
		l.Error(fmt.Errorf("app - Run - register image action: %w", err))
	}
	// Wire the action registry into the NL parser for LLM tool-calling support
	nlParserUseCase.SetActionRegistry(actionRegistry)

	assistantRateLimiter := assistant.NewRateLimiter()
	defer assistantRateLimiter.Close()

	// HTTP Server
	httpServer := httpserver.New(l, httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
	dashboardHandler := restapi.NewRouter(httpServer.App, cfg, containerUseCase, authUseCase, remoteMachineUseCase, taskUseCase, systemMetricsRepo, networkUseCase, volumeUseCase, imageUseCase, composeUseCase, nlParserUseCase, faultDiagnosisUseCase, configRecommendUseCase, actionRegistry, settingsStore, assistantRateLimiter, l)

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
