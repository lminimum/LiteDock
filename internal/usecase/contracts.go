// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// Translation
	Translation interface {
		Translate(context.Context, entity.Translation) (entity.Translation, error)
		History(context.Context) (entity.TranslationHistory, error)
	}

	// Auth -.
	Auth interface {
		// Login - authenticates user and returns token
		Login(context.Context, string, string) (string, *entity.User, error)
		// Register - creates a new user
		Register(context.Context, string, string, string, string) (*entity.User, error)
		// GetCurrentUser - gets user info from token
		GetCurrentUser(context.Context, string) (*entity.User, error)
		// RefreshToken - refreshes authentication token
		RefreshToken(context.Context, string) (string, error)
	}
)
