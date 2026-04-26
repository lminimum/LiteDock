// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"
	"time"

	"github.com/lminimum/LiteDock/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// ContainerRepo - manages Docker containers with caching
	ContainerRepo interface {
		List(context.Context) ([]entity.Container, error)
		Get(context.Context, string) (*entity.Container, error)
		ListByMachine(ctx context.Context, machineID string) ([]entity.Container, error)
		UpsertBatch(ctx context.Context, machineID string, containers []entity.Container) error
		DeleteByMachine(ctx context.Context, machineID string) error
		IsCacheValid(ctx context.Context, machineID string, maxAge time.Duration) (bool, error)
		CountAll(ctx context.Context) (int64, error)
		CountByStatus(ctx context.Context, status string) (int64, error)
	}

	// UserRepo - manages user accounts
	UserRepo interface {
		CreateUser(context.Context, *entity.User) error
		GetUserByID(context.Context, string) (*entity.User, error)
		GetUserByUsername(context.Context, string) (*entity.User, error)
		GetUserByEmail(context.Context, string) (*entity.User, error)
		UpdateUser(context.Context, *entity.User) error
		DeleteUser(context.Context, string) error
		VerifyPassword(context.Context, string, string) (bool, error)
		UpdatePassword(context.Context, string, string) error
		CountUsers(context.Context) (int64, error)
	}

	RemoteMachineRepo interface {
		Create(context.Context, *entity.RemoteMachine) error
		GetByID(context.Context, string) (*entity.RemoteMachine, error)
		List(context.Context) ([]entity.RemoteMachine, error)
		Count(context.Context) (int64, error)
		Update(context.Context, *entity.RemoteMachine) error
		Delete(context.Context, string) error
		GetByHost(context.Context, string) (*entity.RemoteMachine, error)
	}

	SystemMetricsRepo interface {
		Create(ctx context.Context, m *entity.SystemMetric) error
		GetHistory(ctx context.Context, since time.Time) ([]entity.SystemMetric, error)
		DeleteOlderThan(ctx context.Context, before time.Time) error
	}
)
