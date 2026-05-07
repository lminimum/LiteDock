package v1

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/usecase"
	pkgErrors "github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/logger"
)

type ComposeHandler struct {
	uc usecase.Compose
	l  logger.Interface
}

func NewComposeRoutes(g fiber.Router, composeUseCase usecase.Compose, l logger.Interface) {
	h := &ComposeHandler{uc: composeUseCase, l: l}

	// Group under /machines/:id/compose
	g.Get("/machines/:id/compose", h.List)
	g.Get("/machines/:id/compose/:name", h.Get)
	g.Post("/machines/:id/compose", h.Create)
	g.Put("/machines/:id/compose/:name", h.Update)
	g.Delete("/machines/:id/compose/:name", h.Delete)
	g.Post("/machines/:id/compose/:name/up", h.Up)
	g.Post("/machines/:id/compose/:name/down", h.Down)
	g.Post("/machines/:id/compose/:name/build", h.Build)
	g.Post("/machines/:id/compose/:name/start", h.Start)
	g.Post("/machines/:id/compose/:name/stop", h.Stop)
	g.Post("/machines/:id/compose/:name/restart", h.Restart)
	g.Get("/machines/:id/compose/:name/logs", h.Logs)
	g.Get("/machines/:id/compose/:name/ps", h.Ps)
}

func (h *ComposeHandler) List(c *fiber.Ctx) error {
	machineID := c.Params("id")
	projects, err := h.uc.List(c.UserContext(), machineID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successResponse(c, fiber.Map{"projects": projects})
}

func (h *ComposeHandler) Get(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	project, err := h.uc.Get(c.UserContext(), machineID, name)
	if err != nil {
		if errors.Is(err, pkgErrors.ErrNotFound) {
			return errorResponse(c, fiber.StatusNotFound, "project not found")
		}
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successResponse(c, fiber.Map{"project": project})
}

func (h *ComposeHandler) Create(c *fiber.Ctx) error {
	var req struct {
		Name     string `json:"name" validate:"required"`
		Content  string `json:"content" validate:"required"`
		FilePath string `json:"file_path"`
	}
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}
	machineID := c.Params("id")
	project, err := h.uc.Create(c.UserContext(), machineID, req.Name, req.Content, req.FilePath)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return createdResponse(c, fiber.Map{"project": project})
}

func (h *ComposeHandler) Update(c *fiber.Ctx) error {
	var req struct {
		Content string `json:"content" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}
	machineID := c.Params("id")
	name := c.Params("name")
	if err := h.uc.Update(c.UserContext(), machineID, name, req.Content); err != nil {
		if errors.Is(err, pkgErrors.ErrNotFound) {
			return errorResponse(c, fiber.StatusNotFound, "project not found")
		}
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successMessage(c, "project updated")
}

func (h *ComposeHandler) Delete(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	if err := h.uc.Delete(c.UserContext(), machineID, name); err != nil {
		if errors.Is(err, pkgErrors.ErrNotFound) {
			return errorResponse(c, fiber.StatusNotFound, "project not found")
		}
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successMessage(c, "project deleted")
}

func (h *ComposeHandler) Up(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	if err := h.uc.Up(c.UserContext(), machineID, name); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	c.Status(fiber.StatusAccepted)
	return successMessage(c, "compose up started")
}

func (h *ComposeHandler) Down(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	volumes := c.Query("volumes") == "true"
	if err := h.uc.Down(c.UserContext(), machineID, name, volumes); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successMessage(c, "compose down completed")
}

func (h *ComposeHandler) Build(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	if err := h.uc.Build(c.UserContext(), machineID, name); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successMessage(c, "compose build completed")
}

func (h *ComposeHandler) Start(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	if err := h.uc.Start(c.UserContext(), machineID, name); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successMessage(c, "compose services started")
}

func (h *ComposeHandler) Stop(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	if err := h.uc.Stop(c.UserContext(), machineID, name); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successMessage(c, "compose services stopped")
}

func (h *ComposeHandler) Restart(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	if err := h.uc.Restart(c.UserContext(), machineID, name); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successMessage(c, "compose services restarted")
}

func (h *ComposeHandler) Logs(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	logs, err := h.uc.Logs(c.UserContext(), machineID, name)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successResponse(c, fiber.Map{"logs": logs})
}

func (h *ComposeHandler) Ps(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	services, err := h.uc.Ps(c.UserContext(), machineID, name)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successResponse(c, fiber.Map{"services": services})
}
