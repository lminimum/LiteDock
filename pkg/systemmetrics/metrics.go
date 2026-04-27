// Package systemmetrics provides a unified interface for collecting
// system resource usage metrics (CPU, memory, disk).
package systemmetrics

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

// SystemMetrics holds a point-in-time snapshot of system resource usage.
type SystemMetrics struct {
	CPU    float64
	Memory float64
	Disk   float64
	At     time.Time
}

// GetSystemMetrics collects all three metrics at once.
// Uses 200ms for CPU sampling (unified from 100ms/500ms).
func GetSystemMetrics() (SystemMetrics, error) {
	cpuVal, err := getCPUUsage()
	if err != nil {
		return SystemMetrics{}, fmt.Errorf("systemmetrics: cpu: %w", err)
	}

	memVal, err := getMemoryUsage()
	if err != nil {
		return SystemMetrics{}, fmt.Errorf("systemmetrics: memory: %w", err)
	}

	diskVal, err := getDiskUsage()
	if err != nil {
		return SystemMetrics{}, fmt.Errorf("systemmetrics: disk: %w", err)
	}

	return SystemMetrics{
		CPU:    cpuVal,
		Memory: memVal,
		Disk:   diskVal,
		At:     time.Now(),
	}, nil
}

func getCPUUsage() (float64, error) {
	percent, err := cpu.Percent(200*time.Millisecond, false)
	if err != nil {
		return 0, fmt.Errorf("systemmetrics: cpu: %w", err)
	}
	if len(percent) == 0 {
		return 0, fmt.Errorf("systemmetrics: cpu: no percent values returned")
	}
	return percent[0], nil
}

func getMemoryUsage() (float64, error) {
	m, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("systemmetrics: memory: %w", err)
	}
	return m.UsedPercent, nil
}

func getDiskUsage() (float64, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return 0, fmt.Errorf("systemmetrics: disk partitions: %w", err)
	}
	if len(parts) == 0 {
		return 0, fmt.Errorf("systemmetrics: disk: no partitions found")
	}
	usage, err := disk.Usage(parts[0].Mountpoint)
	if err != nil {
		return 0, fmt.Errorf("systemmetrics: disk usage: %w", err)
	}
	return usage.UsedPercent, nil
}
