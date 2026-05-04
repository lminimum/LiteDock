// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	dockerImage "github.com/docker/docker/api/types/image"

	"github.com/lminimum/LiteDock/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// Container -.
	Container interface {
		List(ctx context.Context) ([]entity.Container, error)
		Get(ctx context.Context, id string) (*entity.Container, error)
		CountAll(ctx context.Context) (int64, error)
		CountByStatus(ctx context.Context, status string) (int64, error)
	}

	// Auth -.
	Auth interface {
		Login(ctx context.Context, username, password string) (string, *entity.User, error)
		Register(ctx context.Context, username, email, password, role string) (*entity.User, error)
		GetCurrentUser(ctx context.Context, token string) (*entity.User, error)
		RefreshToken(ctx context.Context, token string) (string, error)
		IsSetupComplete(ctx context.Context) (bool, error)
	}

	// RemoteMachine -.
	RemoteMachine interface {
		Create(ctx context.Context, machine *entity.RemoteMachine) (*entity.RemoteMachine, error)
		GetByID(ctx context.Context, id string) (*entity.RemoteMachine, error)
		List(ctx context.Context) ([]entity.RemoteMachine, error)
		Update(ctx context.Context, machine *entity.RemoteMachine) error
		Delete(ctx context.Context, id string) error
		GetByHost(ctx context.Context, host string) (*entity.RemoteMachine, error)
	}

	// Network -.
	Network interface {
		ListNetworks(ctx context.Context, machineID string) ([]entity.Network, error)
		CreateNetwork(ctx context.Context, machineID string, name, driver string) (*entity.Network, error)
		DeleteNetwork(ctx context.Context, machineID string, networkName string) error
		InspectNetwork(ctx context.Context, machineID string, networkName string) (*entity.Network, error)
	}

	// Volume -.
	Volume interface {
		ListVolumes(ctx context.Context, machineID string) ([]entity.Volume, error)
		CreateVolume(ctx context.Context, machineID string, name, driver string) (*entity.Volume, error)
		DeleteVolume(ctx context.Context, machineID string, volumeName string) error
		InspectVolume(ctx context.Context, machineID string, volumeName string) (*entity.Volume, error)
	}

	// Image -.
	Image interface {
		List(ctx context.Context, machineID string) ([]entity.Image, error)
		Inspect(ctx context.Context, machineID, imageID string) (*entity.Image, error)
		Pull(ctx context.Context, machineID, repository, tag string) (*entity.Image, error)
		Delete(ctx context.Context, machineID, imageID string) ([]dockerImage.DeleteResponse, error)
		Prune(ctx context.Context, machineID string) (*dockerImage.PruneReport, error)
	}
)
