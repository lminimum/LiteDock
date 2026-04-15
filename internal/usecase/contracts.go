// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
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
	}
)
