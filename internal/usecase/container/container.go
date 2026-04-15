package container

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/logger"
)

// UseCase handles container business logic.
type UseCase struct {
	repo interface {
		List(ctx context.Context) ([]entity.Container, error)
		Get(ctx context.Context, id string) (*entity.Container, error)
	}
	l logger.Interface
}

// New creates a new container use case.
func New(repo interface {
	List(ctx context.Context) ([]entity.Container, error)
	Get(ctx context.Context, id string) (*entity.Container, error)
}, l logger.Interface) *UseCase {
	return &UseCase{repo: repo, l: l}
}

// List returns all containers.
func (uc *UseCase) List(ctx context.Context) ([]entity.Container, error) {
	return uc.repo.List(ctx)
}

// Get returns a container by ID.
func (uc *UseCase) Get(ctx context.Context, id string) (*entity.Container, error) {
	return uc.repo.Get(ctx, id)
}
