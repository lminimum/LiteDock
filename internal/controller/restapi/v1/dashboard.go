package v1

import (
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/internal/usecase"
	"github.com/lminimum/LiteDock/internal/usecase/remote_machine"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type DashboardHandler struct {
	uc          remote_machine.UseCaseInterface
	containerUC usecase.Container
	metricsRepo repo.SystemMetricsRepo
	l           logger.Interface
}

func NewDashboardRoutes(apiV1Group fiber.Router, uc remote_machine.UseCaseInterface, container usecase.Container, metricsRepo repo.SystemMetricsRepo, l logger.Interface) {
	h := &DashboardHandler{uc: uc, containerUC: container, metricsRepo: metricsRepo, l: l}

	dashboard := apiV1Group.Group("/dashboard")
	{
		dashboard.Get("/stats", h.Stats)
		dashboard.Get("/resources", h.Resources)
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

func (h *DashboardHandler) Resources(c *fiber.Ctx) error {
	cpuVal := getCPUUsage()
	memoryVal := getMemoryUsage()
	diskVal := getDiskUsage()

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"cpu":    math.Round(cpuVal*10) / 10,
			"memory": math.Round(memoryVal*10) / 10,
			"disk":   math.Round(diskVal*10) / 10,
		},
	})
}

func getCPUUsage() float64 {
	percent, err := cpu.Percent(100*time.Millisecond, false)
	if err != nil || len(percent) == 0 {
		return 0
	}
	return percent[0]
}

func getMemoryUsage() float64 {
	m, err := mem.VirtualMemory()
	if err != nil {
		return 0
	}
	return m.UsedPercent
}

func getDiskUsage() float64 {
	parts, err := disk.Partitions(false)
	if err != nil {
		return 0
	}
	if len(parts) == 0 {
		return 0
	}
	usage, err := disk.Usage(parts[0].Mountpoint)
	if err != nil {
		return 0
	}
	return usage.UsedPercent
}
