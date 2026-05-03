package entity

import "time"

// Volume represents a Docker volume.
type Volume struct {
	Name       string            `json:"name"`
	MachineID  string            `json:"machine_id"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	CreatedAt  string            `json:"created_at"`
	Scope      string            `json:"scope"`
	Labels     map[string]string `json:"labels"`
	Size       int64             `json:"size"`
	CachedAt   time.Time         `json:"cached_at"`
}
