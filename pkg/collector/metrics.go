package collector

import (
	"context"
	"time"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/lminimum/LiteDock/pkg/systemmetrics"
)

type MetricsCollector struct {
	metricsRepo   repo.SystemMetricsRepo
	l             logger.Interface
	interval      time.Duration
	stopCh        chan struct{}
	cleanupStopCh chan struct{}
}

func NewMetricsCollector(metricsRepo repo.SystemMetricsRepo, l logger.Interface, interval time.Duration) *MetricsCollector {
	return &MetricsCollector{
		metricsRepo:   metricsRepo,
		l:             l,
		interval:      interval,
		stopCh:        make(chan struct{}),
		cleanupStopCh: make(chan struct{}),
	}
}

// Start begins periodic metric collection and scheduled cleanup.
// It blocks until Stop() is called.
func (mc *MetricsCollector) Start() {
	go func() {
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mc.cleanupOld()
			case <-mc.cleanupStopCh:
				return
			}
		}
	}()

	mc.collect()

	ticker := time.NewTicker(mc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.collect()
		case <-mc.stopCh:
			return
		}
	}
}

// Stop signals both the collection loop and the cleanup goroutine to stop.
func (mc *MetricsCollector) Stop() {
	close(mc.cleanupStopCh)
	close(mc.stopCh)
}

const (
	maxMetricsRecords = 3600
)

func (mc *MetricsCollector) collect() {
	ctx := context.Background()

	sm, err := systemmetrics.GetSystemMetrics()
	if err != nil {
		mc.l.Error(err, "MetricsCollector.collect failed")
		return
	}

	m := &entity.SystemMetric{
		RecordedAt:    sm.At,
		CPUPercent:    sm.CPU,
		MemoryPercent: sm.Memory,
		DiskPercent:   sm.Disk,
	}

	if err := mc.metricsRepo.Create(ctx, m); err != nil {
		mc.l.Error(err, "MetricsCollector.collect failed to save metric")
		return
	}

	if err := mc.metricsRepo.PruneToCount(ctx, maxMetricsRecords); err != nil {
		mc.l.Warn("MetricsCollector.collect.PruneToCount failed: %v", err)
	}

	mc.l.Debug("MetricsCollector: collected cpu=%.2f memory=%.2f disk=%.2f", sm.CPU, sm.Memory, sm.Disk)
}

func (mc *MetricsCollector) cleanupOld() {
	ctx := context.Background()
	threshold := time.Now().Add(-48 * time.Hour)
	if err := mc.metricsRepo.DeleteOlderThan(ctx, threshold); err != nil {
		mc.l.Warn("MetricsCollector.cleanupOld failed: %v", err)
	}
}
