// Package v1 implements routing paths. Each services in own file.
package restapi

import (
	"net/http"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/lminimum/LiteDock/config"
	_ "github.com/lminimum/LiteDock/docs" // Swagger docs.
	"github.com/lminimum/LiteDock/internal/controller/restapi/middleware"
	v1 "github.com/lminimum/LiteDock/internal/controller/restapi/v1"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/internal/usecase"
	"github.com/lminimum/LiteDock/internal/usecase/remote_machine"
	"github.com/lminimum/LiteDock/pkg/logger"
)

// NewRouter
// Swagger spec:
// @title       LiteDock API
// @description Docker container management platform
// @version     1.0
// @host        localhost:8080
// @BasePath    /v1
func NewRouter(app *fiber.App, cfg *config.Config, container usecase.Container, auth usecase.Auth, remoteMachine *remote_machine.UseCase, metricsRepo repo.SystemMetricsRepo, networkUseCase usecase.Network, volumeUseCase usecase.Volume, imageUseCase usecase.Image, l logger.Interface) *v1.DashboardHandler {
	// Options
	app.Use(middleware.Logger(l))
	app.Use(middleware.Recovery(l))

	// Prometheus metrics
	if cfg.Metrics.Enabled {
		prometheus := fiberprometheus.New("my-service-name")
		prometheus.RegisterAt(app, "/metrics")
		app.Use(prometheus.Middleware)
	}

	// Swagger
	if cfg.Swagger.Enabled {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// K8s probe
	app.Get("/healthz", func(ctx *fiber.Ctx) error { return ctx.SendStatus(http.StatusOK) })

	// Auth middleware for protected routes
	authMiddleware := middleware.AuthRequired(cfg)

	// Base v1 group
	apiV1Group := app.Group("/v1")

	var dashboardHandler *v1.DashboardHandler

	// Public routes FIRST — prevents protected group middleware from intercepting auth paths
	{
		v1.NewAuthRoutes(apiV1Group, auth, l, cfg)
	}

	// Protected routes (require authentication)
	protected := apiV1Group.Group("", authMiddleware)
	{
		v1.NewContainerRoutes(protected, container, l)
		v1.NewNetworkRoutes(protected, networkUseCase, l)
		v1.NewVolumeRoutes(protected, volumeUseCase, l)
		v1.NewImageRoutes(protected, imageUseCase, l)
		v1.NewRemoteMachineRoutes(protected, remoteMachine, l)
		dashboardHandler = v1.NewDashboardRoutes(protected, remoteMachine, container, metricsRepo, l)
	}

	return dashboardHandler
}
