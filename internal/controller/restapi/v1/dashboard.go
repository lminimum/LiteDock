package v1

import (
	"math"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/internal/usecase"
	"github.com/lminimum/LiteDock/internal/usecase/remote_machine"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/lminimum/LiteDock/pkg/systemmetrics"
)

// MetricsWriter abstracts a WebSocket connection for streaming metrics.
// *websocket.Conn satisfies this interface via its WriteJSON method.
type MetricsWriter interface {
	WriteJSON(v any) error
}

type DashboardHandler struct {
	uc          remote_machine.UseCaseInterface
	containerUC usecase.Container
	metricsRepo repo.SystemMetricsRepo
	l           logger.Interface
	activeConns sync.Map
}

func NewDashboardRoutes(apiV1Group fiber.Router, uc remote_machine.UseCaseInterface, container usecase.Container, metricsRepo repo.SystemMetricsRepo, l logger.Interface) *DashboardHandler {
	h := &DashboardHandler{uc: uc, containerUC: container, metricsRepo: metricsRepo, l: l}

	dashboard := apiV1Group.Group("/dashboard")
	{
		dashboard.Get("/stats", h.Stats)
		dashboard.Get("/resources", h.Resources)
		dashboard.Get("/resources/history", h.ResourcesHistory)
		dashboard.Get("/resources/stream", websocket.New(h.ResourcesStreamWS))
	}

	return h
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
	sm, err := systemmetrics.GetSystemMetrics()
	if err != nil {
		h.l.Error(err, "DashboardHandler.Resources.GetSystemMetrics failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"cpu":    math.Round(sm.CPU*10) / 10,
			"memory": math.Round(sm.Memory*10) / 10,
			"disk":   math.Round(sm.Disk*10) / 10,
		},
	})
}

// ResourcesHistory handles GET /v1/dashboard/resources/history
// @Summary Get historical system resources metrics
// @Description Get historical metrics for the chart (last N minutes)
// @Tags dashboard
// @Produce json
// @Param minutes query int false "Number of minutes to look back" default(5)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /dashboard/resources/history [get]
func (h *DashboardHandler) ResourcesHistory(c *fiber.Ctx) error {
	minutes := c.QueryInt("minutes", 5)
	if minutes <= 0 || minutes > 60 {
		minutes = 5
	}

	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	metrics, err := h.metricsRepo.GetHistory(c.Context(), since)
	if err != nil {
		h.l.Error(err, "DashboardHandler.ResourcesHistory.GetHistory failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	data := make([]map[string]interface{}, 0, len(metrics))
	for _, m := range metrics {
		data = append(data, map[string]interface{}{
			"cpu":    math.Round(m.CPUPercent*10) / 10,
			"memory": math.Round(m.MemoryPercent*10) / 10,
			"disk":   math.Round(m.DiskPercent*10) / 10,
			"time":   m.RecordedAt.Format("15:04:05"),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func (h *DashboardHandler) ResourcesStreamWS(c *websocket.Conn) {
	h.register(c)
	defer h.unregister(c)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm, err := systemmetrics.GetSystemMetrics()
			if err != nil {
				h.l.Error(err, "DashboardHandler.ResourcesStreamWS.GetSystemMetrics failed")
				continue
			}

			data := map[string]interface{}{
				"cpu":    math.Round(sm.CPU*10) / 10,
				"memory": math.Round(sm.Memory*10) / 10,
				"disk":   math.Round(sm.Disk*10) / 10,
				"time":   sm.At.Format("15:04:05"),
			}

			if err := c.WriteJSON(data); err != nil {
				h.l.Warn("DashboardHandler.ResourcesStreamWS.WriteJSON failed: %v", err)
				return
			}
		}
	}
}

func (h *DashboardHandler) register(c *websocket.Conn) {
	h.activeConns.Store(c, struct{}{})
}

func (h *DashboardHandler) unregister(c *websocket.Conn) {
	h.activeConns.Delete(c)
}

func (h *DashboardHandler) CloseAllConnections() {
	h.activeConns.Range(func(key, _ interface{}) bool {
		if conn, ok := key.(*websocket.Conn); ok {
			conn.Close()
		}
		return true
	})
	h.l.Info("DashboardHandler: closed all active WebSocket connections")
}

