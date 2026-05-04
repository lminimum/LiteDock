package v1

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/usecase"
	"github.com/lminimum/LiteDock/pkg/logger"
)

type ImageHandler struct {
	uc usecase.Image
	l  logger.Interface
	v  *validator.Validate
}

// NewImageRoutes registers image management routes under /v1/machines/:id/images.
func NewImageRoutes(apiV1Group fiber.Router, imageUseCase usecase.Image, l logger.Interface) {
	h := &ImageHandler{uc: imageUseCase, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	machineGroup := apiV1Group.Group("/machines")
	{
		machineGroup.Get("/:id/images", h.List)
		machineGroup.Get("/:id/images/:imageID", h.Inspect)
		machineGroup.Post("/:id/images/pull", h.Pull)
		machineGroup.Delete("/:id/images/:imageID", h.Delete)
		machineGroup.Post("/:id/images/prune", h.Prune)
	}
}

// PullImageRequest is the request body for pulling a Docker image.
type PullImageRequest struct {
	Repository string `json:"repository" validate:"required"`
	Tag        string `json:"tag"`
}

// List handles GET /v1/machines/:id/images
func (h *ImageHandler) List(c *fiber.Ctx) error {
	machineID := c.Params("id")

	images, err := h.uc.List(c.Context(), machineID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, fiber.Map{
		"images": images,
	})
}

// Inspect handles GET /v1/machines/:id/images/:imageID
func (h *ImageHandler) Inspect(c *fiber.Ctx) error {
	machineID := c.Params("id")
	imageID := c.Params("imageID")

	image, err := h.uc.Inspect(c.Context(), machineID, imageID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, image)
}

// Pull handles POST /v1/machines/:id/images/pull
func (h *ImageHandler) Pull(c *fiber.Ctx) error {
	machineID := c.Params("id")

	var req PullImageRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	image, err := h.uc.Pull(c.Context(), machineID, req.Repository, req.Tag)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return createdResponse(c, image)
}

// Delete handles DELETE /v1/machines/:id/images/:imageID
func (h *ImageHandler) Delete(c *fiber.Ctx) error {
	machineID := c.Params("id")
	imageID := c.Params("imageID")

	_, err := h.uc.Delete(c.Context(), machineID, imageID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successMessage(c, "Image deleted successfully")
}

// Prune handles POST /v1/machines/:id/images/prune
func (h *ImageHandler) Prune(c *fiber.Ctx) error {
	machineID := c.Params("id")

	report, err := h.uc.Prune(c.Context(), machineID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, fiber.Map{
		"spaceReclaimed": report.SpaceReclaimed,
		"imagesDeleted":  len(report.ImagesDeleted),
	})
}
