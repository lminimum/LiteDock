package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/usecase"
	"github.com/lminimum/LiteDock/internal/usecase/remote_machine"
	"github.com/lminimum/LiteDock/pkg/logger"
)

type DashboardHandler struct {
	uc            remote_machine.UseCaseInterface
	containerUC   usecase.Container
	l             logger.Interface
}

func NewDashboardRoutes(apiV1Group fiber.Router, uc remote_machine.UseCaseInterface, container usecase.Container, l logger.Interface) {
	h := &DashboardHandler{uc: uc, containerUC: container, l: l}

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

	totalContainers, err := h.containerUC.CountAll(c.Context())
	if err != nil {
		h.l.Warn("DashboardHandler.Stats.CountAll failed: %v", err)
		totalContainers = 0
	}

	runningContainers, err := h.containerUC.CountByStatus(c.Context(), "running")
	if err != nil {
		h.l.Warn("DashboardHandler.Stats.CountByStatus failed: %v", err)
		runningContainers = 0
	}

	stoppedContainers, err := h.containerUC.CountByStatus(c.Context(), "exited")
	if err != nil {
		h.l.Warn("DashboardHandler.Stats.CountByStatus failed: %v", err)
		stoppedContainers = 0
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"machines": fiber.Map{
				"total": machineCount,
			},
			"containers": fiber.Map{
				"total":   totalContainers,
				"running": runningContainers,
				"stopped": stoppedContainers,
			},
		},
	})
}
