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

type VolumeRepo struct {
	db database.DB
}

func NewVolumeRepo(db database.DB) *VolumeRepo {
	return &VolumeRepo{db: db}
}

func (r *VolumeRepo) ListByMachine(ctx context.Context, machineID string) ([]entity.Volume, error) {
	query := `
		SELECT name, machine_id, driver, mountpoint, created_at, scope, labels, size, cached_at
		FROM volumes WHERE machine_id = ?`

	rowsInterface, err := r.db.Query(ctx, query, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "VolumeRepo.ListByMachine.Query")
	}

	scanner, ok := rowsInterface.(interface {
		Next() bool
		Scan(...any) error
		Close() error
		Err() error
	})
	if !ok {
		return nil, stdErrors.New("VolumeRepo.ListByMachine: rows does not implement scanner interface")
	}
	defer scanner.Close()

	volumes := make([]entity.Volume, 0)

	for scanner.Next() {
		var v entity.Volume
		var labelsJSON []byte

		err := scanRow(scanner,
			&v.Name,
			&v.MachineID,
			&v.Driver,
			&v.Mountpoint,
			&v.CreatedAt,
			&v.Scope,
			&labelsJSON,
			&v.Size,
			&v.CachedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "VolumeRepo.ListByMachine.scanRow")
		}

		if labelsJSON != nil {
			if err := json.Unmarshal(labelsJSON, &v.Labels); err != nil {
				return nil, errors.Wrap(err, "VolumeRepo.ListByMachine.UnmarshalLabels")
			}
		}

		volumes = append(volumes, v)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "VolumeRepo.ListByMachine.rowsErr")
	}

	return volumes, nil
}

func (r *VolumeRepo) GetByName(ctx context.Context, machineID, name string) (*entity.Volume, error) {
	query := `
		SELECT name, machine_id, driver, mountpoint, created_at, scope, labels, size, cached_at
		FROM volumes WHERE machine_id = ? AND name = ?`

	row := r.db.QueryRow(ctx, query, machineID, name)

	var v entity.Volume
	var labelsJSON []byte

	err := scanRow(row,
		&v.Name,
		&v.MachineID,
		&v.Driver,
		&v.Mountpoint,
		&v.CreatedAt,
		&v.Scope,
		&labelsJSON,
		&v.Size,
		&v.CachedAt,
	)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrVolumeNotFound
		}
		return nil, errors.Wrap(err, "VolumeRepo.GetByName.Scan")
	}

	if labelsJSON != nil {
		if err := json.Unmarshal(labelsJSON, &v.Labels); err != nil {
			return nil, errors.Wrap(err, "VolumeRepo.GetByName.UnmarshalLabels")
		}
	}

	return &v, nil
}

func (r *VolumeRepo) UpsertBatch(ctx context.Context, machineID string, volumes []entity.Volume) error {
	if len(volumes) == 0 {
		return nil
	}

	deleteQuery := `DELETE FROM volumes WHERE machine_id = ?`
	err := r.db.Exec(ctx, deleteQuery, machineID)
	if err != nil {
		return errors.Wrap(err, "VolumeRepo.UpsertBatch.Delete")
	}

	insertQuery := `
		INSERT INTO volumes (name, machine_id, driver, mountpoint, created_at, scope, labels, size, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()

	for _, v := range volumes {
		labelsJSON, err := json.Marshal(v.Labels)
		if err != nil {
			return errors.Wrap(err, "VolumeRepo.UpsertBatch.MarshalLabels")
		}

		err = r.db.Exec(ctx, insertQuery,
			v.Name,
			machineID,
			v.Driver,
			v.Mountpoint,
			v.CreatedAt,
			v.Scope,
			labelsJSON,
			v.Size,
			now,
		)
		if err != nil {
			return errors.Wrap(err, "VolumeRepo.UpsertBatch.Insert")
		}
	}

	return nil
}

func (r *VolumeRepo) DeleteByMachine(ctx context.Context, machineID string) error {
	query := `DELETE FROM volumes WHERE machine_id = ?`

	err := r.db.Exec(ctx, query, machineID)
	if err != nil {
		return errors.Wrap(err, "VolumeRepo.DeleteByMachine.Exec")
	}

	return nil
}

func (r *VolumeRepo) IsCacheValid(ctx context.Context, machineID string, maxAge time.Duration) (bool, error) {
	query := `
		SELECT cached_at FROM volumes WHERE machine_id = ? LIMIT 1`

	row := r.db.QueryRow(ctx, query, machineID)

	var cachedAt time.Time
	err := scanRow(row, &cachedAt)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, errors.Wrap(err, "VolumeRepo.IsCacheValid.QueryRow")
	}

	return time.Now().Before(cachedAt.Add(maxAge)), nil
}
