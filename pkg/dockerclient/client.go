package dockerclient

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/lminimum/LiteDock/internal/entity"
)

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
	ContainerRemove(ctx context.Context, containerID string, force bool) error
	ContainerInspect(ctx context.Context, containerID string) (*container.InspectResponse, error)
	NetworkList(ctx context.Context) ([]entity.Network, error)
	NetworkCreate(ctx context.Context, name, driver string) (*entity.Network, error)
	NetworkDelete(ctx context.Context, networkID string) error
	NetworkInspect(ctx context.Context, networkID string) (*entity.Network, error)
	Close() error
}
