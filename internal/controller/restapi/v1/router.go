package v1

import (
	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// NewContainerRoutes
func NewContainerRoutes(apiV1Group fiber.Router, c usecase.Container, l logger.Interface) {
	r := &ContainerHandler{c: c, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	containerGroup := apiV1Group.Group("/containers")

	{
		containerGroup.Get("/", r.List)
		containerGroup.Get("/:id", r.Get)
	}
}

// NewAuthRoutes -.
func NewAuthRoutes(apiV1Group fiber.Router, auth usecase.Auth, l logger.Interface, cfg *config.Config) {
	authHandler := &AuthHandler{auth: auth, l: l, cfg: cfg}

	authGroup := apiV1Group.Group("/auth")

	{
		authGroup.Post("/login", authHandler.Login)
		authGroup.Post("/register", authHandler.Register)
		authGroup.Get("/me", authHandler.GetMe)
	}
}
