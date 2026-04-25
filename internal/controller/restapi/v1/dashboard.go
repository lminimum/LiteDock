package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/usecase/remote_machine"
	"github.com/lminimum/LiteDock/pkg/logger"
)

type DashboardHandler struct {
	uc remote_machine.UseCaseInterface
	l  logger.Interface
}

func NewDashboardRoutes(apiV1Group fiber.Router, uc remote_machine.UseCaseInterface, l logger.Interface) {
	h := &DashboardHandler{uc: uc, l: l}

	dashboard := apiV1Group.Group("/dashboard")
	{
		dashboard.Get("/stats", h.Stats)
	}
}

// Stats handles GET /v1/dashboard/stats
// @Summary Get dashboard statistics
// @Description Get aggregated stats for the dashboard (machine count, container counts)
// @Tags dashboard
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /dashboard/stats [get]
func (h *DashboardHandler) Stats(c *fiber.Ctx) error {
	machineCount, err := h.uc.Count(c.Context())
	if err != nil {
		h.l.Error(err, "DashboardHandler.Stats.Count failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"machines": fiber.Map{
				"total": machineCount,
			},
		},
	})
}
