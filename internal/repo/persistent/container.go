package persistent

import (
	"context"
	"database/sql"
	"encoding/json"
	stdErrors "errors"
	"time"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/database"
	"github.com/lminimum/LiteDock/pkg/errors"
)

type ContainerRepo struct {
	db database.DB
}

func NewContainerRepo(db database.DB) *ContainerRepo {
	return &ContainerRepo{db: db}
}

func (r *ContainerRepo) List(ctx context.Context) ([]entity.Container, error) {
	query := `
		SELECT id, machine_id, name, image, status, ports, created_at, cached_at
		FROM containers ORDER BY cached_at DESC`

	rowsInterface, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "ContainerRepo.List.Query")
	}

	scanner, ok := rowsInterface.(interface {
		Next() bool
		Scan(...any) error
		Close() error
		Err() error
	})
	if !ok {
		return nil, stdErrors.New("ContainerRepo.List: rows does not implement scanner interface")
	}
	defer scanner.Close()

	containers := make([]entity.Container, 0)

	for scanner.Next() {
		var c entity.Container
		var portsJSON []byte

		err := scanRow(
			scanner,
			&c.ID,
			&c.MachineID,
			&c.Name,
			&c.Image,
			&c.Status,
			&portsJSON,
			&c.Created,
			&c.CachedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "ContainerRepo.List.scanRow")
		}

		if portsJSON != nil {
			if err := json.Unmarshal(portsJSON, &c.Ports); err != nil {
				return nil, errors.Wrap(err, "ContainerRepo.List.UnmarshalPorts")
			}
		}

		containers = append(containers, c)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "ContainerRepo.List.rowsErr")
	}

	return containers, nil
}

func (r *ContainerRepo) Get(_ context.Context, id string) (*entity.Container, error) {
	return &entity.Container{ID: id, Name: "placeholder"}, nil
}

func (r *ContainerRepo) ListByMachine(ctx context.Context, machineID string) ([]entity.Container, error) {
	query := `
		SELECT id, machine_id, name, image, status, ports, created_at, cached_at
		FROM containers WHERE machine_id = ?`

	rowsInterface, err := r.db.Query(ctx, query, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "ContainerRepo.ListByMachine.Query")
	}

	scanner, ok := rowsInterface.(interface {
		Next() bool
		Scan(...any) error
		Close() error
		Err() error
	})
	if !ok {
		return nil, stdErrors.New("ContainerRepo.ListByMachine: rows does not implement scanner interface")
	}
	defer scanner.Close()

	containers := make([]entity.Container, 0)

	for scanner.Next() {
		var c entity.Container
		var portsJSON []byte

		err := scanRow(
			scanner,
			&c.ID,
			&c.MachineID,
			&c.Name,
			&c.Image,
			&c.Status,
			&portsJSON,
			&c.Created,
			&c.CachedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "ContainerRepo.ListByMachine.scanRow")
		}

		if portsJSON != nil {
			if err := json.Unmarshal(portsJSON, &c.Ports); err != nil {
				return nil, errors.Wrap(err, "ContainerRepo.ListByMachine.UnmarshalPorts")
			}
		}

		containers = append(containers, c)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "ContainerRepo.ListByMachine.rowsErr")
	}

	return containers, nil
}

func (r *ContainerRepo) UpsertBatch(ctx context.Context, machineID string, containers []entity.Container) error {
	if len(containers) == 0 {
		return nil
	}

	deleteQuery := `DELETE FROM containers WHERE machine_id = ?`
	err := r.db.Exec(ctx, deleteQuery, machineID)
	if err != nil {
		return errors.Wrap(err, "ContainerRepo.UpsertBatch.Delete")
	}

	insertQuery := `
		INSERT INTO containers (id, machine_id, name, image, status, ports, created_at, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()

	for _, c := range containers {
		portsJSON, err := json.Marshal(c.Ports)
		if err != nil {
			return errors.Wrap(err, "ContainerRepo.UpsertBatch.MarshalPorts")
		}

		err = r.db.Exec(
			ctx, insertQuery,
			c.ID,
			machineID,
			c.Name,
			c.Image,
			c.Status,
			portsJSON,
			c.Created,
			now,
		)
		if err != nil {
			return errors.Wrap(err, "ContainerRepo.UpsertBatch.Insert")
		}
	}

	return nil
}

func (r *ContainerRepo) DeleteByMachine(ctx context.Context, machineID string) error {
	query := `DELETE FROM containers WHERE machine_id = ?`

	err := r.db.Exec(ctx, query, machineID)
	if err != nil {
		return errors.Wrap(err, "ContainerRepo.DeleteByMachine.Exec")
	}

	return nil
}

func (r *ContainerRepo) IsCacheValid(ctx context.Context, machineID string, maxAge time.Duration) (bool, error) {
	query := `
		SELECT cached_at FROM containers WHERE machine_id = ? LIMIT 1`

	row := r.db.QueryRow(ctx, query, machineID)

	var cachedAt time.Time
	err := scanRow(row, &cachedAt)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, errors.Wrap(err, "ContainerRepo.IsCacheValid.QueryRow")
	}

	return time.Now().Before(cachedAt.Add(maxAge)), nil
}

func (r *ContainerRepo) CountAll(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM containers`

	var count int64
	row := r.db.QueryRow(ctx, query)
	err := scanRow(row, &count)
	if err != nil {
		return 0, errors.Wrap(err, "ContainerRepo.CountAll.Scan")
	}

	return count, nil
}

func (r *ContainerRepo) CountByStatus(ctx context.Context, status string) (int64, error) {
	query := `SELECT COUNT(*) FROM containers WHERE status = ?`

	var count int64
	row := r.db.QueryRow(ctx, query, status)
	err := scanRow(row, &count)
	if err != nil {
		return 0, errors.Wrap(err, "ContainerRepo.CountByStatus.Scan")
	}

	return count, nil
}
