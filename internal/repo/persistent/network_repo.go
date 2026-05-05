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

type NetworkRepo struct {
	db database.DB
}

func NewNetworkRepo(db database.DB) *NetworkRepo {
	return &NetworkRepo{db: db}
}

func (r *NetworkRepo) ListByMachine(ctx context.Context, machineID string) ([]entity.Network, error) {
	query := `
		SELECT id, machine_id, name, driver, scope, internal, attachable, enable_ipv6, created, labels, cached_at
		FROM networks WHERE machine_id = ?`

	rowsInterface, err := r.db.Query(ctx, query, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "NetworkRepo.ListByMachine.Query")
	}

	scanner, ok := rowsInterface.(interface {
		Next() bool
		Scan(...any) error
		Close() error
		Err() error
	})
	if !ok {
		return nil, stdErrors.New("NetworkRepo.ListByMachine: rows does not implement scanner interface")
	}
	defer scanner.Close()

	networks := make([]entity.Network, 0)

	for scanner.Next() {
		var n entity.Network
		var labelsJSON []byte

		err := scanRow(scanner,
			&n.ID,
			&n.MachineID,
			&n.Name,
			&n.Driver,
			&n.Scope,
			&n.Internal,
			&n.Attachable,
			&n.EnableIPv6,
			&n.Created,
			&labelsJSON,
			&n.CachedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "NetworkRepo.ListByMachine.scanRow")
		}

		if labelsJSON != nil {
			if err := json.Unmarshal(labelsJSON, &n.Labels); err != nil {
				return nil, errors.Wrap(err, "NetworkRepo.ListByMachine.UnmarshalLabels")
			}
		}

		networks = append(networks, n)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "NetworkRepo.ListByMachine.rowsErr")
	}

	return networks, nil
}

func (r *NetworkRepo) GetByName(ctx context.Context, machineID, name string) (*entity.Network, error) {
	query := `
		SELECT id, machine_id, name, driver, scope, internal, attachable, enable_ipv6, created, labels, cached_at
		FROM networks WHERE machine_id = ? AND name = ?`

	row := r.db.QueryRow(ctx, query, machineID, name)

	var n entity.Network
	var labelsJSON []byte

	err := scanRow(row,
		&n.ID,
		&n.MachineID,
		&n.Name,
		&n.Driver,
		&n.Scope,
		&n.Internal,
		&n.Attachable,
		&n.EnableIPv6,
		&n.Created,
		&labelsJSON,
		&n.CachedAt,
	)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrNetworkNotFound
		}
		return nil, errors.Wrap(err, "NetworkRepo.GetByName.Scan")
	}

	if labelsJSON != nil {
		if err := json.Unmarshal(labelsJSON, &n.Labels); err != nil {
			return nil, errors.Wrap(err, "NetworkRepo.GetByName.UnmarshalLabels")
		}
	}

	return &n, nil
}

func (r *NetworkRepo) UpsertBatch(ctx context.Context, machineID string, networks []entity.Network) error {
	if len(networks) == 0 {
		return nil
	}

	deleteQuery := `DELETE FROM networks WHERE machine_id = ?`
	err := r.db.Exec(ctx, deleteQuery, machineID)
	if err != nil {
		return errors.Wrap(err, "NetworkRepo.UpsertBatch.Delete")
	}

	insertQuery := `
		INSERT INTO networks (id, machine_id, name, driver, scope, internal, attachable, enable_ipv6, created, labels, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()

	for _, n := range networks {
		labelsJSON, err := json.Marshal(n.Labels)
		if err != nil {
			return errors.Wrap(err, "NetworkRepo.UpsertBatch.MarshalLabels")
		}

		err = r.db.Exec(ctx, insertQuery,
			n.ID,
			machineID,
			n.Name,
			n.Driver,
			n.Scope,
			n.Internal,
			n.Attachable,
			n.EnableIPv6,
			n.Created,
			labelsJSON,
			now,
		)
		if err != nil {
			return errors.Wrap(err, "NetworkRepo.UpsertBatch.Insert")
		}
	}

	return nil
}

func (r *NetworkRepo) DeleteByMachine(ctx context.Context, machineID string) error {
	query := `DELETE FROM networks WHERE machine_id = ?`

	err := r.db.Exec(ctx, query, machineID)
	if err != nil {
		return errors.Wrap(err, "NetworkRepo.DeleteByMachine.Exec")
	}

	return nil
}

func (r *NetworkRepo) IsCacheValid(ctx context.Context, machineID string, maxAge time.Duration) (bool, error) {
	query := `
		SELECT cached_at FROM networks WHERE machine_id = ? LIMIT 1`

	row := r.db.QueryRow(ctx, query, machineID)

	var cachedAt time.Time
	err := scanRow(row, &cachedAt)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, errors.Wrap(err, "NetworkRepo.IsCacheValid.QueryRow")
	}

	return time.Now().Before(cachedAt.Add(maxAge)), nil
}
