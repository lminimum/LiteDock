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

// CreateComposeRequest is the request body for creating a Compose project.
type CreateComposeRequest struct {
	Name     string `json:"name" validate:"required"`
	Content  string `json:"content" validate:"required"`
	FilePath string `json:"file_path"`
}

// UpdateComposeRequest is the request body for updating a Compose project.
type UpdateComposeRequest struct {
	Content string `json:"content" validate:"required"`
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

// List handles GET /v1/machines/:id/compose
// @Summary List Compose projects
// @Description Get all Docker Compose projects on a remote machine
// @Tags compose
// @Produce json
// @Param id path string true "Machine ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose [get]
func (h *ComposeHandler) List(c *fiber.Ctx) error {
	machineID := c.Params("id")
	projects, err := h.uc.List(c.UserContext(), machineID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successResponse(c, fiber.Map{"projects": projects})
}

// Get handles GET /v1/machines/:id/compose/:name
// @Summary Get a Compose project
// @Description Get Docker Compose project details from a remote machine
// @Tags compose
// @Produce json
// @Param id path string true "Machine ID"
// @Param name path string true "Compose project name"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose/{name} [get]
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

// Create handles POST /v1/machines/:id/compose
// @Summary Create a Compose project
// @Description Create a Docker Compose project on a remote machine
// @Tags compose
// @Accept json
// @Produce json
// @Param id path string true "Machine ID"
// @Param request body CreateComposeRequest true "Compose project creation request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose [post]
func (h *ComposeHandler) Create(c *fiber.Ctx) error {
	var req CreateComposeRequest
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

// Update handles PUT /v1/machines/:id/compose/:name
// @Summary Update a Compose project
// @Description Update Docker Compose project content on a remote machine
// @Tags compose
// @Accept json
// @Produce json
// @Param id path string true "Machine ID"
// @Param name path string true "Compose project name"
// @Param request body UpdateComposeRequest true "Compose project update request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose/{name} [put]
func (h *ComposeHandler) Update(c *fiber.Ctx) error {
	var req UpdateComposeRequest
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

// Delete handles DELETE /v1/machines/:id/compose/:name
// @Summary Delete a Compose project
// @Description Delete a Docker Compose project from a remote machine
// @Tags compose
// @Produce json
// @Param id path string true "Machine ID"
// @Param name path string true "Compose project name"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose/{name} [delete]
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

// Up handles POST /v1/machines/:id/compose/:name/up
// @Summary Start a Compose project
// @Description Run docker compose up for a project on a remote machine
// @Tags compose
// @Produce json
// @Param id path string true "Machine ID"
// @Param name path string true "Compose project name"
// @Success 202 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose/{name}/up [post]
func (h *ComposeHandler) Up(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	if err := h.uc.Up(c.UserContext(), machineID, name); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	c.Status(fiber.StatusAccepted)
	return successMessage(c, "compose up started")
}

// Down handles POST /v1/machines/:id/compose/:name/down
// @Summary Stop and remove a Compose project
// @Description Run docker compose down for a project on a remote machine
// @Tags compose
// @Produce json
// @Param id path string true "Machine ID"
// @Param name path string true "Compose project name"
// @Param volumes query bool false "Remove named volumes" default(false)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose/{name}/down [post]
func (h *ComposeHandler) Down(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	volumes := c.Query("volumes") == "true"
	if err := h.uc.Down(c.UserContext(), machineID, name, volumes); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successMessage(c, "compose down completed")
}

// Build handles POST /v1/machines/:id/compose/:name/build
// @Summary Build Compose project images
// @Description Run docker compose build for a project on a remote machine
// @Tags compose
// @Produce json
// @Param id path string true "Machine ID"
// @Param name path string true "Compose project name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose/{name}/build [post]
func (h *ComposeHandler) Build(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	if err := h.uc.Build(c.UserContext(), machineID, name); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successMessage(c, "compose build completed")
}

// Start handles POST /v1/machines/:id/compose/:name/start
// @Summary Start Compose services
// @Description Start services for a Docker Compose project on a remote machine
// @Tags compose
// @Produce json
// @Param id path string true "Machine ID"
// @Param name path string true "Compose project name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose/{name}/start [post]
func (h *ComposeHandler) Start(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	if err := h.uc.Start(c.UserContext(), machineID, name); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successMessage(c, "compose services started")
}

// Stop handles POST /v1/machines/:id/compose/:name/stop
// @Summary Stop Compose services
// @Description Stop services for a Docker Compose project on a remote machine
// @Tags compose
// @Produce json
// @Param id path string true "Machine ID"
// @Param name path string true "Compose project name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose/{name}/stop [post]
func (h *ComposeHandler) Stop(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	if err := h.uc.Stop(c.UserContext(), machineID, name); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successMessage(c, "compose services stopped")
}

// Restart handles POST /v1/machines/:id/compose/:name/restart
// @Summary Restart Compose services
// @Description Restart services for a Docker Compose project on a remote machine
// @Tags compose
// @Produce json
// @Param id path string true "Machine ID"
// @Param name path string true "Compose project name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose/{name}/restart [post]
func (h *ComposeHandler) Restart(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	if err := h.uc.Restart(c.UserContext(), machineID, name); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successMessage(c, "compose services restarted")
}

// Logs handles GET /v1/machines/:id/compose/:name/logs
// @Summary Get Compose project logs
// @Description Get logs for a Docker Compose project on a remote machine
// @Tags compose
// @Produce json
// @Param id path string true "Machine ID"
// @Param name path string true "Compose project name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose/{name}/logs [get]
func (h *ComposeHandler) Logs(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	logs, err := h.uc.Logs(c.UserContext(), machineID, name)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successResponse(c, fiber.Map{"logs": logs})
}

// Ps handles GET /v1/machines/:id/compose/:name/ps
// @Summary List Compose project services
// @Description List services and containers for a Docker Compose project on a remote machine
// @Tags compose
// @Produce json
// @Param id path string true "Machine ID"
// @Param name path string true "Compose project name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/compose/{name}/ps [get]
func (h *ComposeHandler) Ps(c *fiber.Ctx) error {
	machineID := c.Params("id")
	name := c.Params("name")
	services, err := h.uc.Ps(c.UserContext(), machineID, name)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return successResponse(c, fiber.Map{"services": services})
}
