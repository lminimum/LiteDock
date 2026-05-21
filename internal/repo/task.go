package repo

import (
	"context"
	"github.com/lminimum/LiteDock/internal/entity"
)

type TaskRepo interface {
	Create(ctx context.Context, t *entity.Task) error
	GetByID(ctx context.Context, id string) (*entity.Task, error)
	Update(ctx context.Context, t *entity.Task) error
	AppendLogs(ctx context.Context, id string, logs string) error
	List(ctx context.Context, limit, offset int) ([]entity.Task, error)
}
