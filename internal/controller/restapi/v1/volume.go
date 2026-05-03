package v1

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/usecase"
	"github.com/lminimum/LiteDock/pkg/logger"
)

type VolumeHandler struct {
	uc usecase.Volume
	l  logger.Interface
	v  *validator.Validate
}

// NewVolumeRoutes
// @Summary Register volume routes
// @Tags volumes
// @Accept json
// @Produce json
func NewVolumeRoutes(apiV1Group fiber.Router, volumeUseCase usecase.Volume, l logger.Interface) {
	h := &VolumeHandler{uc: volumeUseCase, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	machineGroup := apiV1Group.Group("/machines")
	{
		machineGroup.Get("/:id/volumes", h.List)
		machineGroup.Post("/:id/volumes", h.Create)
		machineGroup.Get("/:id/volumes/:volumeName", h.Get)
		machineGroup.Delete("/:id/volumes/:volumeName", h.Delete)
	}
}

type CreateVolumeRequest struct {
	Name   string `json:"name" validate:"required"`
	Driver string `json:"driver"`
}

// List - handles GET /v1/machines/:id/volumes
// @Summary List volumes on a machine
// @Description Get all Docker volumes on a remote host
// @Tags volumes
// @Produce json
// @Param id path string true "Machine ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/volumes [get]
func (h *VolumeHandler) List(c *fiber.Ctx) error {
	id := c.Params("id")

	volumes, err := h.uc.ListVolumes(c.Context(), id)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, fiber.Map{
		"volumes": volumes,
	})
}

// Create - handles POST /v1/machines/:id/volumes
// @Summary Create a new volume
// @Description Create a new Docker volume on a remote host
// @Tags volumes
// @Accept json
// @Produce json
// @Param id path string true "Machine ID"
// @Param request body CreateVolumeRequest true "Volume creation request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/volumes [post]
func (h *VolumeHandler) Create(c *fiber.Ctx) error {
	id := c.Params("id")

	var req CreateVolumeRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	volume, err := h.uc.CreateVolume(c.Context(), id, req.Name, req.Driver)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return createdResponse(c, volume)
}

// Get - handles GET /v1/machines/:id/volumes/:volumeName
// @Summary Inspect a volume
// @Description Get detailed information about a Docker volume
// @Tags volumes
// @Produce json
// @Param id path string true "Machine ID"
// @Param volumeName path string true "Volume Name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/volumes/{volumeName} [get]
func (h *VolumeHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	volumeName := c.Params("volumeName")

	result, err := h.uc.InspectVolume(c.Context(), id, volumeName)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, result)
}

// Delete - handles DELETE /v1/machines/:id/volumes/:volumeName
// @Summary Delete a volume
// @Description Remove a Docker volume from a remote host
// @Tags volumes
// @Produce json
// @Param id path string true "Machine ID"
// @Param volumeName path string true "Volume Name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/volumes/{volumeName} [delete]
func (h *VolumeHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	volumeName := c.Params("volumeName")

	err := h.uc.DeleteVolume(c.Context(), id, volumeName)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successMessage(c, "Volume deleted successfully")
}
