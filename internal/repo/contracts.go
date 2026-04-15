// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// ContainerRepo - manages Docker containers
	ContainerRepo interface {
		List(context.Context) ([]entity.Container, error)
		Get(context.Context, string) (*entity.Container, error)
	}

	// UserRepo - manages user accounts
	UserRepo interface {
		CreateUser(context.Context, entity.User) error
		GetUserByID(context.Context, string) (*entity.User, error)
		GetUserByUsername(context.Context, string) (*entity.User, error)
		GetUserByEmail(context.Context, string) (*entity.User, error)
		UpdateUser(context.Context, entity.User) error
		DeleteUser(context.Context, string) error
		VerifyPassword(context.Context, string, string) (bool, error)
		UpdatePassword(context.Context, string, string) error
	}
)
