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

type ComposeRepo struct {
	db database.DB
}

func NewComposeRepo(db database.DB) *ComposeRepo {
	db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS compose_projects (
			id TEXT PRIMARY KEY,
			machine_id TEXT NOT NULL,
			name TEXT NOT NULL,
			file_path TEXT NOT NULL,
			project_name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'unknown',
			services TEXT NOT NULL DEFAULT '[]',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			cached_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(machine_id, project_name)
		)
	`)

	return &ComposeRepo{db: db}
}

func marshalServices(services []entity.ComposeService) ([]byte, error) {
	return json.Marshal(services)
}

func unmarshalServices(data []byte) ([]entity.ComposeService, error) {
	var services []entity.ComposeService
	if err := json.Unmarshal(data, &services); err != nil {
		return nil, err
	}

	return services, nil
}

const composeFileColumns = `id, machine_id, name, file_path, project_name, status, services, created_at, updated_at, cached_at`

func scanComposeFile(row interface{}) (entity.ComposeFile, error) {
	var cf entity.ComposeFile
	var servicesJSON []byte

	err := scanRow(
		row,
		&cf.ID,
		&cf.MachineID,
		&cf.Name,
		&cf.FilePath,
		&cf.ProjectName,
		&cf.Status,
		&servicesJSON,
		&cf.CreatedAt,
		&cf.UpdatedAt,
		&cf.CachedAt,
	)
	if err != nil {
		return cf, err
	}

	if servicesJSON != nil {
		services, err := unmarshalServices(servicesJSON)
		if err != nil {
			return cf, err
		}

		cf.Services = services
	}

	return cf, nil
}

func (r *ComposeRepo) ListByMachine(ctx context.Context, machineID string) ([]entity.ComposeFile, error) {
	query := `SELECT ` + composeFileColumns + ` FROM compose_projects WHERE machine_id = ? ORDER BY updated_at DESC`

	rowsInterface, err := r.db.Query(ctx, query, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "ComposeRepo.ListByMachine.Query")
	}

	scanner, ok := rowsInterface.(interface {
		Next() bool
		Scan(...any) error
		Close() error
		Err() error
	})
	if !ok {
		return nil, stdErrors.New("ComposeRepo.ListByMachine: rows does not implement scanner interface")
	}

	defer scanner.Close()

	composeFiles := make([]entity.ComposeFile, 0)

	for scanner.Next() {
		cf, err := scanComposeFile(scanner)
		if err != nil {
			return nil, errors.Wrap(err, "ComposeRepo.ListByMachine.scanComposeFile")
		}

		composeFiles = append(composeFiles, cf)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "ComposeRepo.ListByMachine.rowsErr")
	}

	return composeFiles, nil
}

func (r *ComposeRepo) GetByID(ctx context.Context, machineID, id string) (*entity.ComposeFile, error) {
	query := `SELECT ` + composeFileColumns + ` FROM compose_projects WHERE machine_id = ? AND id = ?`

	row := r.db.QueryRow(ctx, query, machineID, id)

	cf, err := scanComposeFile(row)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrNotFound
		}

		return nil, errors.Wrap(err, "ComposeRepo.GetByID.scanComposeFile")
	}

	return &cf, nil
}

func (r *ComposeRepo) GetByProjectName(ctx context.Context, machineID, projectName string) (*entity.ComposeFile, error) {
	query := `SELECT ` + composeFileColumns + ` FROM compose_projects WHERE machine_id = ? AND project_name = ?`

	row := r.db.QueryRow(ctx, query, machineID, projectName)

	cf, err := scanComposeFile(row)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrNotFound
		}

		return nil, errors.Wrap(err, "ComposeRepo.GetByProjectName.scanComposeFile")
	}

	return &cf, nil
}

func (r *ComposeRepo) Create(ctx context.Context, composeFile *entity.ComposeFile) error {
	servicesJSON, err := marshalServices(composeFile.Services)
	if err != nil {
		return errors.Wrap(err, "ComposeRepo.Create.marshalServices")
	}

	query := `
		INSERT INTO compose_projects (id, machine_id, name, file_path, project_name, status, services, created_at, updated_at, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()

	err = r.db.Exec(
		ctx, query,
		composeFile.ID,
		composeFile.MachineID,
		composeFile.Name,
		composeFile.FilePath,
		composeFile.ProjectName,
		composeFile.Status,
		servicesJSON,
		now,
		now,
		now,
	)
	if err != nil {
		return errors.Wrap(err, "ComposeRepo.Create.Exec")
	}

	return nil
}

func (r *ComposeRepo) Update(ctx context.Context, composeFile *entity.ComposeFile) error {
	servicesJSON, err := marshalServices(composeFile.Services)
	if err != nil {
		return errors.Wrap(err, "ComposeRepo.Update.marshalServices")
	}

	query := `
		UPDATE compose_projects
		SET file_path = ?, status = ?, services = ?, updated_at = CURRENT_TIMESTAMP
		WHERE machine_id = ? AND project_name = ?`

	err = r.db.Exec(
		ctx, query,
		composeFile.FilePath,
		composeFile.Status,
		servicesJSON,
		composeFile.MachineID,
		composeFile.ProjectName,
	)
	if err != nil {
		return errors.Wrap(err, "ComposeRepo.Update.Exec")
	}

	return nil
}

func (r *ComposeRepo) DeleteByMachine(ctx context.Context, machineID string) error {
	query := `DELETE FROM compose_projects WHERE machine_id = ?`

	err := r.db.Exec(ctx, query, machineID)
	if err != nil {
		return errors.Wrap(err, "ComposeRepo.DeleteByMachine.Exec")
	}

	return nil
}

func (r *ComposeRepo) DeleteByID(ctx context.Context, machineID, id string) error {
	query := `DELETE FROM compose_projects WHERE machine_id = ? AND id = ?`

	err := r.db.Exec(ctx, query, machineID, id)
	if err != nil {
		return errors.Wrap(err, "ComposeRepo.DeleteByID.Exec")
	}

	return nil
}

func (r *ComposeRepo) IsCacheValid(ctx context.Context, machineID string, maxAge time.Duration) (bool, error) {
	query := `SELECT cached_at FROM compose_projects WHERE machine_id = ? LIMIT 1`

	row := r.db.QueryRow(ctx, query, machineID)

	var cachedAt time.Time
	err := scanRow(row, &cachedAt)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, errors.Wrap(err, "ComposeRepo.IsCacheValid.QueryRow")
	}

	return time.Now().Before(cachedAt.Add(maxAge)), nil
}

func (r *ComposeRepo) UpsertBatch(ctx context.Context, composeFiles []entity.ComposeFile) error {
	if len(composeFiles) == 0 {
		return nil
	}

	machineID := composeFiles[0].MachineID

	deleteQuery := `DELETE FROM compose_projects WHERE machine_id = ?`
	err := r.db.Exec(ctx, deleteQuery, machineID)
	if err != nil {
		return errors.Wrap(err, "ComposeRepo.UpsertBatch.Delete")
	}

	insertQuery := `
		INSERT INTO compose_projects (id, machine_id, name, file_path, project_name, status, services, created_at, updated_at, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()

	for _, cf := range composeFiles {
		servicesJSON, err := marshalServices(cf.Services)
		if err != nil {
			return errors.Wrap(err, "ComposeRepo.UpsertBatch.marshalServices")
		}

		err = r.db.Exec(
			ctx, insertQuery,
			cf.ID,
			cf.MachineID,
			cf.Name,
			cf.FilePath,
			cf.ProjectName,
			cf.Status,
			servicesJSON,
			now,
			now,
			now,
		)
		if err != nil {
			return errors.Wrap(err, "ComposeRepo.UpsertBatch.Insert")
		}
	}

	return nil
}
