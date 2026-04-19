package persistent

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/sqlite"
)

func setupTestDB(t *testing.T) *sqlite.SQLite {
	db, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS remote_machines (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			host VARCHAR(255) NOT NULL,
			port INTEGER DEFAULT 22,
			username VARCHAR(255) NOT NULL,
			auth_method VARCHAR(20) NOT NULL DEFAULT 'password',
			password VARCHAR(255),
			ssh_key_path VARCHAR(512),
			ssh_key TEXT,
			docker_host VARCHAR(512) DEFAULT '/var/run/docker.sock',
			status VARCHAR(20) DEFAULT 'unknown',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func TestRemoteMachineRepoCreate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRemoteMachineRepo(db)
	ctx := context.Background()

	machine := &entity.RemoteMachine{
		ID:         uuid.New().String(),
		Name:       "test-server",
		Host:       "192.168.1.100",
		Port:       22,
		Username:   "root",
		AuthMethod: entity.AuthMethodPassword,
		Password:   "secret",
		DockerHost: "/var/run/docker.sock",
		Status:     "unknown",
	}

	err := repo.Create(ctx, machine)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, machine.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Name != machine.Name {
		t.Errorf("expected name %s, got %s", machine.Name, retrieved.Name)
	}
	if retrieved.Host != machine.Host {
		t.Errorf("expected host %s, got %s", machine.Host, retrieved.Host)
	}
}

func TestRemoteMachineRepoGetByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRemoteMachineRepo(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent id")
	}
}

func TestRemoteMachineRepoList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRemoteMachineRepo(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		machine := &entity.RemoteMachine{
			ID:         uuid.New().String(),
			Name:       "server",
			Host:       "192.168.1.100",
			Port:       22,
			Username:   "root",
			AuthMethod: entity.AuthMethodPassword,
			Password:   "secret",
		}
		if i == 1 {
			machine.Host = "192.168.1.101"
		}
		if i == 2 {
			machine.Host = "192.168.1.102"
		}
		repo.Create(ctx, machine)
	}

	machines, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(machines) != 3 {
		t.Errorf("expected 3 machines, got %d", len(machines))
	}
}

func TestRemoteMachineRepoUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRemoteMachineRepo(db)
	ctx := context.Background()

	machine := &entity.RemoteMachine{
		ID:         uuid.New().String(),
		Name:       "original-name",
		Host:       "192.168.1.100",
		Port:       22,
		Username:   "root",
		AuthMethod: entity.AuthMethodPassword,
		Password:   "secret",
	}

	repo.Create(ctx, machine)

	machine.Name = "updated-name"
	err := repo.Update(ctx, machine)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, machine.ID)
	if retrieved.Name != "updated-name" {
		t.Errorf("expected updated-name, got %s", retrieved.Name)
	}
}

func TestRemoteMachineRepoDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRemoteMachineRepo(db)
	ctx := context.Background()

	machine := &entity.RemoteMachine{
		ID:         uuid.New().String(),
		Name:       "to-delete",
		Host:       "192.168.1.100",
		Port:       22,
		Username:   "root",
		AuthMethod: entity.AuthMethodPassword,
		Password:   "secret",
	}

	repo.Create(ctx, machine)
	err := repo.Delete(ctx, machine.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, machine.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestRemoteMachineRepoGetByHost(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRemoteMachineRepo(db)
	ctx := context.Background()

	machine := &entity.RemoteMachine{
		ID:         uuid.New().String(),
		Name:       "test-server",
		Host:       "192.168.1.100",
		Port:       22,
		Username:   "root",
		AuthMethod: entity.AuthMethodPassword,
		Password:   "secret",
	}

	repo.Create(ctx, machine)

	retrieved, err := repo.GetByHost(ctx, "192.168.1.100")
	if err != nil {
		t.Fatalf("GetByHost failed: %v", err)
	}

	if retrieved.ID != machine.ID {
		t.Errorf("expected id %s, got %s", machine.ID, retrieved.ID)
	}
}
