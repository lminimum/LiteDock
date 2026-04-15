// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// TranslationRepo
	TranslationRepo interface {
		Store(context.Context, entity.Translation) error
		GetHistory(context.Context) ([]entity.Translation, error)
	}

	// TranslationWebAPI
	TranslationWebAPI interface {
		Translate(entity.Translation) (entity.Translation, error)
	}

	// UserRepo - manages user accounts
	UserRepo interface {
		// CreateUser creates a new user
		CreateUser(context.Context, entity.User) error
		// GetUserByID retrieves a user by ID
		GetUserByID(context.Context, string) (*entity.User, error)
		// GetUserByUsername retrieves a user by username
		GetUserByUsername(context.Context, string) (*entity.User, error)
		// GetUserByEmail retrieves a user by email
		GetUserByEmail(context.Context, string) (*entity.User, error)
		// UpdateUser updates a user's information
		UpdateUser(context.Context, entity.User) error
		// DeleteUser deletes a user
		DeleteUser(context.Context, string) error
		// VerifyPassword verifies if the provided password matches the user's hash
		VerifyPassword(context.Context, string, string) (bool, error) // username, password
		// UpdatePassword updates a user's password
		UpdatePassword(context.Context, string, string) error // userID, newPasswordHash
	}
)
