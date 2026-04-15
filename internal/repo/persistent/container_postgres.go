package persistent

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

// ContainerRepo implements container repository.
type ContainerRepo struct{}

// NewContainerRepo creates a new container repository.
func NewContainerRepo(pg interface{}) *ContainerRepo {
	return &ContainerRepo{}
}

// List returns all containers (placeholder).
func (r *ContainerRepo) List(ctx context.Context) ([]entity.Container, error) {
	return []entity.Container{}, nil
}

// Get returns a container by ID (placeholder).
func (r *ContainerRepo) Get(ctx context.Context, id string) (*entity.Container, error) {
	return &entity.Container{ID: id, Name: "placeholder"}, nil
}
