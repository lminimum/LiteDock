package auth

import (
	"context"

	"github.com/lminimum/LiteDock/internal/entity"
)

// UseCaseInterface -.
type UseCaseInterface interface {
	// Login - authenticates user and returns token
	Login(context.Context, string, string) (string, *entity.User, error)
	// Register - creates a new user
	Register(context.Context, string, string, string, string) (*entity.User, error)
	// GetCurrentUser - gets user info from token
	GetCurrentUser(context.Context, string) (*entity.User, error)
	// RefreshToken - refreshes authentication token
	RefreshToken(context.Context, string) (string, error)
}
