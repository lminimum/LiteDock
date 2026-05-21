package persistent

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/database"
	pkgErrors "github.com/lminimum/LiteDock/pkg/errors"
)

type TaskRepo struct {
	db database.DB
}

func NewTaskRepo(db database.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, t *entity.Task) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now

	query := `
		INSERT INTO tasks (id, type, status, machine_id, payload, result, error, logs, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	err := r.db.Exec(
		ctx, query,
		t.ID, t.Type, t.Status, t.MachineID, t.Payload, t.Result, t.Error, t.Logs, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return pkgErrors.Wrap(err, "TaskRepo.Create.Exec")
	}
	return nil
}

func (r *TaskRepo) GetByID(ctx context.Context, id string) (*entity.Task, error) {
	query := `SELECT id, type, status, machine_id, payload, result, error, logs, created_at, updated_at FROM tasks WHERE id = ?`
	var t entity.Task
	row := r.db.QueryRow(ctx, query, id)
	err := scanRow(row, &t.ID, &t.Type, &t.Status, &t.MachineID, &t.Payload, &t.Result, &t.Error, &t.Logs, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, pkgErrors.Wrap(err, "TaskRepo.GetByID.Scan")
	}
	return &t, nil
}

func (r *TaskRepo) Update(ctx context.Context, t *entity.Task) error {
	t.UpdatedAt = time.Now()
	query := `UPDATE tasks SET status = ?, result = ?, error = ?, logs = ?, updated_at = ? WHERE id = ?`
	err := r.db.Exec(ctx, query, t.Status, t.Result, t.Error, t.Logs, t.UpdatedAt, t.ID)
	if err != nil {
		return pkgErrors.Wrap(err, "TaskRepo.Update.Exec")
	}
	return nil
}

func (r *TaskRepo) AppendLogs(ctx context.Context, id string, logs string) error {
	query := `UPDATE tasks SET logs = COALESCE(logs, '') || ?, updated_at = ? WHERE id = ?`
	err := r.db.Exec(ctx, query, logs, time.Now(), id)
	if err != nil {
		return pkgErrors.Wrap(err, "TaskRepo.AppendLogs.Exec")
	}
	return nil
}

func (r *TaskRepo) List(ctx context.Context, limit, offset int) ([]entity.Task, error) {
	query := `SELECT id, type, status, machine_id, payload, result, error, logs, created_at, updated_at FROM tasks ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rowsInterface, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, pkgErrors.Wrap(err, "TaskRepo.List.Query")
	}

	scanner, ok := rowsInterface.(interface {
		Next() bool
		Scan(...any) error
		Close() error
		Err() error
	})
	if !ok {
		return nil, errors.New("TaskRepo.List: rows does not implement scanner interface")
	}
	defer scanner.Close()

	tasks := make([]entity.Task, 0)
	for scanner.Next() {
		var t entity.Task
		err := scanRow(scanner, &t.ID, &t.Type, &t.Status, &t.MachineID, &t.Payload, &t.Result, &t.Error, &t.Logs, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, pkgErrors.Wrap(err, "TaskRepo.List.scanRow")
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
