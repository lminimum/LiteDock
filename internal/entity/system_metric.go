package entity

import "time"

type SystemMetric struct {
	ID            string    `json:"id"`
	RecordedAt    time.Time `json:"recorded_at"`
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryPercent float64   `json:"memory_percent"`
	DiskPercent   float64   `json:"disk_percent"`
	CreatedAt     time.Time `json:"created_at"`
}
