package v1

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/usecase"
	"github.com/lminimum/LiteDock/pkg/logger"
)

type NetworkHandler struct {
	uc usecase.Network
	l  logger.Interface
	v  *validator.Validate
}

// NewNetworkRoutes
// @Summary Register network routes
// @Tags networks
// @Accept json
// @Produce json
func NewNetworkRoutes(apiV1Group fiber.Router, networkUseCase usecase.Network, l logger.Interface) {
	h := &NetworkHandler{uc: networkUseCase, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	machineGroup := apiV1Group.Group("/machines")
	{
		machineGroup.Get("/:id/networks", h.List)
		machineGroup.Post("/:id/networks", h.Create)
		machineGroup.Get("/:id/networks/:networkID", h.Get)
		machineGroup.Delete("/:id/networks/:networkID", h.Delete)
	}
}

type CreateNetworkRequest struct {
	Name   string `json:"name" validate:"required"`
	Driver string `json:"driver"`
}

// List - handles GET /v1/machines/:id/networks
// @Summary List networks on a machine
// @Description Get all Docker networks on a remote host
// @Tags networks
// @Produce json
// @Param id path string true "Machine ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/networks [get]
func (h *NetworkHandler) List(c *fiber.Ctx) error {
	id := c.Params("id")

	networks, err := h.uc.ListNetworks(c.Context(), id)
	if err != nil {
		h.l.Error(err, "NetworkHandler - List")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, fiber.Map{
		"networks": networks,
	})
}

// Create - handles POST /v1/machines/:id/networks
// @Summary Create a new network
// @Description Create a new Docker network on a remote host
// @Tags networks
// @Accept json
// @Produce json
// @Param id path string true "Machine ID"
// @Param request body CreateNetworkRequest true "Network creation request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/networks [post]
func (h *NetworkHandler) Create(c *fiber.Ctx) error {
	id := c.Params("id")

	var req CreateNetworkRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	network, err := h.uc.CreateNetwork(c.Context(), id, req.Name, req.Driver)
	if err != nil {
		h.l.Error(err, "NetworkHandler - Create")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return createdResponse(c, network)
}

// Get - handles GET /v1/machines/:id/networks/:networkID
// @Summary Inspect a network
// @Description Get detailed information about a Docker network
// @Tags networks
// @Produce json
// @Param id path string true "Machine ID"
// @Param networkID path string true "Network ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/networks/{networkID} [get]
func (h *NetworkHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	networkID := c.Params("networkID")

	result, err := h.uc.InspectNetwork(c.Context(), id, networkID)
	if err != nil {
		h.l.Error(err, "NetworkHandler - Get")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, result)
}

// Delete - handles DELETE /v1/machines/:id/networks/:networkID
// @Summary Delete a network
// @Description Remove a Docker network from a remote host
// @Tags networks
// @Produce json
// @Param id path string true "Machine ID"
// @Param networkID path string true "Network ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/networks/{networkID} [delete]
func (h *NetworkHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	networkID := c.Params("networkID")

	err := h.uc.DeleteNetwork(c.Context(), id, networkID)
	if err != nil {
		h.l.Error(err, "NetworkHandler - Delete")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successMessage(c, "Network deleted successfully")
}
