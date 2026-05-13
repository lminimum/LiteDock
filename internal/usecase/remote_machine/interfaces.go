package remote_machine

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/lminimum/LiteDock/internal/entity"
)

type UseCaseInterface interface {
	Create(context.Context, *entity.RemoteMachine) (*entity.RemoteMachine, error)
	GetByID(context.Context, string) (*entity.RemoteMachine, error)
	List(context.Context) ([]entity.RemoteMachine, error)
	Count(context.Context) (int64, error)
	Update(context.Context, *entity.RemoteMachine) error
	Delete(context.Context, string) error
	GetByHost(context.Context, string) (*entity.RemoteMachine, error)
	TestConnection(context.Context, string) error
	ListContainers(context.Context, string) ([]entity.Container, error)
	GetContainerLogs(context.Context, string, string, string) (string, error)
	ExecContainer(context.Context, string, string, []string) (string, error)
	CreateContainer(context.Context, string, *container.Config, *container.HostConfig, string) (*container.CreateResponse, error)
	StartContainer(context.Context, string, string) error
	StopContainer(context.Context, string, string) error
	RestartContainer(context.Context, string, string) error
	RemoveContainer(context.Context, string, string, bool) error
	InspectContainer(context.Context, string, string) (interface{}, error)
}
