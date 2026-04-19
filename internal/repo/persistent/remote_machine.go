package persistent

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/database"
	"github.com/lminimum/LiteDock/pkg/errors"
)

type RemoteMachineRepo struct {
	db database.DB
}

func NewRemoteMachineRepo(db database.DB) *RemoteMachineRepo {
	return &RemoteMachineRepo{db: db}
}

func (r *RemoteMachineRepo) Create(ctx context.Context, m *entity.RemoteMachine) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	now := time.Now()

	query := `
		INSERT INTO remote_machines (id, name, host, port, username, auth_method, password, ssh_key_path, ssh_key, docker_host, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	err := r.db.Exec(ctx, query,
		m.ID,
		m.Name,
		m.Host,
		m.Port,
		m.Username,
		m.AuthMethod,
		m.Password,
		m.SSHKeyPath,
		m.SSHKey,
		m.DockerHost,
		m.Status,
		now,
		now,
	)
	if err != nil {
		return errors.Wrap(err, "RemoteMachineRepo.Create.Exec")
	}

	return nil
}

func (r *RemoteMachineRepo) GetByID(ctx context.Context, id string) (*entity.RemoteMachine, error) {
	query := `
		SELECT id, name, host, port, username, auth_method, password, ssh_key_path, ssh_key, docker_host, status, created_at, updated_at
		FROM remote_machines WHERE id = ?`

	var m entity.RemoteMachine

	row := r.db.QueryRow(ctx, query, id)

	err := scanRow(row,
		&m.ID,
		&m.Name,
		&m.Host,
		&m.Port,
		&m.Username,
		&m.AuthMethod,
		&m.Password,
		&m.SSHKeyPath,
		&m.SSHKey,
		&m.DockerHost,
		&m.Status,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, errors.Wrap(errors.ErrRemoteMachineNotFound, "RemoteMachineRepo.GetByID")
		}
		return nil, errors.Wrap(err, "RemoteMachineRepo.GetByID.QueryRow")
	}

	return &m, nil
}

func (r *RemoteMachineRepo) List(ctx context.Context) ([]entity.RemoteMachine, error) {
	query := `
		SELECT id, name, host, port, username, auth_method, password, ssh_key_path, ssh_key, docker_host, status, created_at, updated_at
		FROM remote_machines ORDER BY name`

	rowsInterface, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "RemoteMachineRepo.List.Query")
	}

	scanner, ok := rowsInterface.(interface {
		Next() bool
		Scan(...any) error
		Close() error
	})
	if !ok {
		return nil, errors.Wrap(err, "RemoteMachineRepo.List.typeAssertion")
	}
	defer scanner.Close()

	machines := make([]entity.RemoteMachine, 0)

	for scanner.Next() {
		var m entity.RemoteMachine
		err := scanRow(scanner,
			&m.ID,
			&m.Name,
			&m.Host,
			&m.Port,
			&m.Username,
			&m.AuthMethod,
			&m.Password,
			&m.SSHKeyPath,
			&m.SSHKey,
			&m.DockerHost,
			&m.Status,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "RemoteMachineRepo.List.scanRow")
		}
		machines = append(machines, m)
	}

	return machines, nil
}

func (r *RemoteMachineRepo) Update(ctx context.Context, m *entity.RemoteMachine) error {
	query := `
		UPDATE remote_machines
		SET name=?, host=?, port=?, username=?, auth_method=?, password=?, ssh_key_path=?, ssh_key=?, docker_host=?, status=?, updated_at=?
		WHERE id=?`

	err := r.db.Exec(ctx, query,
		m.Name,
		m.Host,
		m.Port,
		m.Username,
		m.AuthMethod,
		m.Password,
		m.SSHKeyPath,
		m.SSHKey,
		m.DockerHost,
		m.Status,
		time.Now(),
		m.ID,
	)
	if err != nil {
		return errors.Wrap(err, "RemoteMachineRepo.Update.Exec")
	}

	return nil
}

func (r *RemoteMachineRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM remote_machines WHERE id = ?`

	err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "RemoteMachineRepo.Delete.Exec")
	}

	return nil
}

func (r *RemoteMachineRepo) GetByHost(ctx context.Context, host string) (*entity.RemoteMachine, error) {
	query := `
		SELECT id, name, host, port, username, auth_method, password, ssh_key_path, ssh_key, docker_host, status, created_at, updated_at
		FROM remote_machines WHERE host = ?`

	var m entity.RemoteMachine

	row := r.db.QueryRow(ctx, query, host)

	err := scanRow(row,
		&m.ID,
		&m.Name,
		&m.Host,
		&m.Port,
		&m.Username,
		&m.AuthMethod,
		&m.Password,
		&m.SSHKeyPath,
		&m.SSHKey,
		&m.DockerHost,
		&m.Status,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, errors.Wrap(errors.ErrRemoteMachineNotFound, "RemoteMachineRepo.GetByHost")
		}
		return nil, errors.Wrap(err, "RemoteMachineRepo.GetByHost.QueryRow")
	}

	return &m, nil
}
