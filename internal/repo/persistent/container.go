package persistent

import (
	"context"

	"github.com/lminimum/LiteDock/internal/entity"
)

// ContainerRepo implements container repository.
type ContainerRepo struct{}

// NewContainerRepo creates a new container repository.
func NewContainerRepo(_ interface{}) *ContainerRepo {
	return &ContainerRepo{}
}

// List returns all containers (placeholder).
func (r *ContainerRepo) List(_ context.Context) ([]entity.Container, error) {
	return []entity.Container{}, nil
}

// Get returns a container by ID (placeholder).
func (r *ContainerRepo) Get(_ context.Context, id string) (*entity.Container, error) {
	return &entity.Container{ID: id, Name: "placeholder"}, nil
}
