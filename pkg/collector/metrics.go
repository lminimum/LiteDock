package collector

import (
	"context"
	"time"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type MetricsCollector struct {
	metricsRepo repo.SystemMetricsRepo
	l           logger.Interface
	interval    time.Duration
	stopCh      chan struct{}
}

func NewMetricsCollector(metricsRepo repo.SystemMetricsRepo, l logger.Interface, interval time.Duration) *MetricsCollector {
	return &MetricsCollector{
		metricsRepo: metricsRepo,
		l:           l,
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

func (mc *MetricsCollector) Start() {
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

func (mc *MetricsCollector) Stop() {
	close(mc.stopCh)
}

func (mc *MetricsCollector) collect() {
	ctx := context.Background()

	cpuVal := getCPUUsage()
	memoryVal := getMemoryUsage()
	diskVal := getDiskUsage()

	m := &entity.SystemMetric{
		RecordedAt:    time.Now(),
		CPUPercent:    cpuVal,
		MemoryPercent: memoryVal,
		DiskPercent:   diskVal,
	}

	if err := mc.metricsRepo.Create(ctx, m); err != nil {
		mc.l.Error(err, "MetricsCollector.collect failed to save metric")
		return
	}

	mc.l.Debug("MetricsCollector: collected cpu=%.2f memory=%.2f disk=%.2f", cpuVal, memoryVal, diskVal)

	go mc.cleanupOld()
}

func (mc *MetricsCollector) cleanupOld() {
	ctx := context.Background()
	threshold := time.Now().Add(-48 * time.Hour)
	if err := mc.metricsRepo.DeleteOlderThan(ctx, threshold); err != nil {
		mc.l.Warn("MetricsCollector.cleanupOld failed: %v", err)
	}
}

func getCPUUsage() float64 {
	percent, err := cpu.Percent(500*time.Millisecond, false)
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
