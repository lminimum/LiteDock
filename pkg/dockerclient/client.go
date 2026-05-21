package dockerclient

import (
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/lminimum/LiteDock/internal/entity"
)

// ComposeServiceStatus represents a compose service's runtime status
type ComposeServiceStatus struct {
	Name        string        `json:"name"`
	ServiceName string        `json:"service_name"`
	Image       string        `json:"image"`
	Status      string        `json:"status"` // running, exited, paused
	Health      string        `json:"health"` // healthy, unhealthy, none
	Replicas    int           `json:"replicas"`
	Publishers  []PublishInfo `json:"publishers,omitempty"`
}

// PublishInfo represents a port publishing entry
type PublishInfo struct {
	URL           string `json:"url"`
	TargetPort    int    `json:"target_port"`
	PublishedPort int    `json:"published_port"`
}

// Client defines the Docker operations interface.
// Both RemoteClient (SSH-based) and LocalClient (direct socket) implement this.
type Client interface {
	Ping(ctx context.Context) error
	ContainerList(ctx context.Context) ([]entity.Container, error)
	ContainerLogs(ctx context.Context, containerID, tail string) (string, error)
	ContainerExec(ctx context.Context, containerID string, cmd []string) (string, error)
	ContainerStart(ctx context.Context, containerID string) error
	ContainerStop(ctx context.Context, containerID string, timeout time.Duration) error
	ContainerRestart(ctx context.Context, containerID string, timeout time.Duration) error
	ContainerPause(ctx context.Context, containerID string) error
	ContainerUnpause(ctx context.Context, containerID string) error
	ContainerKill(ctx context.Context, containerID string) error
	ContainerRemove(ctx context.Context, containerID string, force bool) error
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, containerName string) (*container.CreateResponse, error)
	ContainerInspect(ctx context.Context, containerID string) (*container.InspectResponse, error)
	NetworkList(ctx context.Context) ([]entity.Network, error)
	NetworkCreate(ctx context.Context, name, driver string) (*entity.Network, error)
	NetworkDelete(ctx context.Context, networkID string) error
	NetworkInspect(ctx context.Context, networkID string) (*entity.Network, error)
	VolumeList(ctx context.Context) ([]entity.Volume, error)
	VolumeCreate(ctx context.Context, name, driver string) (*entity.Volume, error)
	VolumeDelete(ctx context.Context, volumeID string) error
	VolumeInspect(ctx context.Context, volumeID string) (*entity.Volume, error)
	ImageList(ctx context.Context, opts image.ListOptions) ([]entity.Image, error)
	ImagePull(ctx context.Context, ref string, opts image.PullOptions) error
	ImageRemove(ctx context.Context, id string, opts image.RemoveOptions) ([]image.DeleteResponse, error)
	ImageInspect(ctx context.Context, id string) (image.InspectResponse, error)
	ImagePrune(ctx context.Context, opts filters.Args) (image.PruneReport, error)
	// ComposeUp starts compose project services
	ComposeUp(ctx context.Context, machineID, projectName, composeFilePath string) error
	// ComposeDown stops and removes compose project resources
	ComposeDown(ctx context.Context, machineID, projectName string, volumes bool) error
	// ComposeBuild builds or rebuilds compose project services
	ComposeBuild(ctx context.Context, machineID, composeFilePath string) error
	// ComposeStart starts existing compose project containers
	ComposeStart(ctx context.Context, machineID, projectName string) error
	// ComposeStop stops running compose project containers
	ComposeStop(ctx context.Context, machineID, projectName string) error
	// ComposeRestart restarts compose project containers
	ComposeRestart(ctx context.Context, machineID, projectName string) error
	// ComposePs lists compose project services with status
	ComposePs(ctx context.Context, machineID, projectName string) ([]ComposeServiceStatus, error)
	// ComposeLogs returns compose project logs
	ComposeLogs(ctx context.Context, machineID, projectName string) (io.ReadCloser, error)
	// ComposeConfig validates and returns compose file config
	ComposeConfig(ctx context.Context, machineID, composeFilePath string) (string, error)
	Close() error
}
