package entity

import "time"

// Image represents a Docker image.
type Image struct {
	ID          string            `json:"id"`          // Docker image ID (sha256 prefix)
	MachineID   string            `json:"machine_id"`
	RepoTags    []string          `json:"repo_tags"`    // ["nginx:latest", "nginx:1.25"]
	RepoDigests []string          `json:"repo_digests"`
	Size        int64             `json:"size"`
	CreatedAt   time.Time         `json:"created_at"`
	Labels      map[string]string `json:"labels"`
	CachedAt    time.Time         `json:"cached_at"`
}