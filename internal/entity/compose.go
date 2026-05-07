package entity

import "time"

// ComposeOperationStatus represents the status of a compose operation.
const (
	OpStarting = "starting"
	OpRunning  = "running"
	OpStopped  = "stopped"
	OpFailed   = "failed"
)

// ComposeFile represents a Docker Compose file.
type ComposeFile struct {
	ID          string           `json:"id"`
	MachineID   string           `json:"machine_id"`
	Name        string           `json:"name"`
	FilePath    string           `json:"file_path"`
	ProjectName string           `json:"project_name"`
	Status      string           `json:"status"`
	Services    []ComposeService `json:"services"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	CachedAt    time.Time        `json:"cached_at"`
}

// ComposeService represents a service within a Docker Compose file.
type ComposeService struct {
	Name       string      `json:"name"`
	Image      string      `json:"image"`
	Status     string      `json:"status"`
	Health     string      `json:"health"`
	Replicas   int         `json:"replicas"`
	Publishers []Publisher `json:"publishers"`
}

// Publisher represents a port mapping for a compose service.
type Publisher struct {
	URL           string `json:"url"`
	TargetPort    int    `json:"target_port"`
	PublishedPort int    `json:"published_port"`
}
