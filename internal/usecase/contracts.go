// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/lminimum/LiteDock/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// Container -.
	Container interface {
		List(ctx context.Context) ([]entity.Container, error)
		Get(ctx context.Context, id string) (*entity.Container, error)
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
)
