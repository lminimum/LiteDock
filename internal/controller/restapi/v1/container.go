package v1

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/usecase"
	"github.com/lminimum/LiteDock/pkg/logger"
)

// ContainerHandler handles container requests.
type ContainerHandler struct {
	c usecase.Container
	l logger.Interface
	v *validator.Validate
}

// List handles GET /v1/containers
// @Summary List all containers
// @Description Get all Docker containers from all machines
// @Tags containers
// @Produce json
// @Success 200 {object} Response
// @Router /containers [get]
func (h *ContainerHandler) List(c *fiber.Ctx) error {
	containers, err := h.c.List(c.Context())
	if err != nil {
		h.l.Error(err, "ContainerHandler - List")
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to list containers")
	}

	return successResponse(c, fiber.Map{
		"containers": containers,
	})
}

// Get handles GET /v1/containers/:id
// @Summary Get a container by ID
// @Description Get details of a specific Docker container (local)
// @Tags containers
// @Produce json
// @Param id path string true "Container ID"
// @Success 200 {object} map[string]interface{}
// @Router /containers/{id} [get]
func (h *ContainerHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	return successResponse(c, fiber.Map{
		"id":   id,
		"name": "placeholder",
	})
}
