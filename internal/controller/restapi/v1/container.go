package v1

import (
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// ContainerHandler handles container requests.
type ContainerHandler struct {
	c usecase.Container
	l logger.Interface
	v *validator.Validate
}

// List handles GET /v1/containers
func (h *ContainerHandler) List(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"containers": []interface{}{}})
}

// Get handles GET /v1/containers/:id
func (h *ContainerHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"id": id, "name": "placeholder"})
}
