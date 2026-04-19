package remote_machine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lminimum/LiteDock/internal/entity"
)

type mockRepo struct {
	machines map[string]*entity.RemoteMachine
}

func (m *mockRepo) Create(ctx context.Context, machine *entity.RemoteMachine) error {
	if machine.ID == "" {
		machine.ID = uuid.New().String()
	}
	m.machines[machine.ID] = machine
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*entity.RemoteMachine, error) {
	if m, ok := m.machines[id]; ok {
		return m, nil
	}
	return nil, errors.New("not found")
}

func (m *mockRepo) List(ctx context.Context) ([]entity.RemoteMachine, error) {
	result := make([]entity.RemoteMachine, 0, len(m.machines))
	for _, v := range m.machines {
		result = append(result, *v)
	}
	return result, nil
}

func (m *mockRepo) Update(ctx context.Context, machine *entity.RemoteMachine) error {
	m.machines[machine.ID] = machine
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	delete(m.machines, id)
	return nil
}

func (m *mockRepo) GetByHost(ctx context.Context, host string) (*entity.RemoteMachine, error) {
	for _, v := range m.machines {
		if v.Host == host {
			return v, nil
		}
	}
	return nil, errors.New("not found")
}

type mockLogger struct{}

func (m *mockLogger) Error(message interface{}, args ...interface{})     {}
func (m *mockLogger) Info(msg string, args ...interface{})              {}
func (m *mockLogger) Debug(msg interface{}, args ...interface{})        {}
func (m *mockLogger) Warn(msg string, args ...interface{})                {}
func (m *mockLogger) Fatal(message interface{}, args ...interface{})     {}
func (m *mockLogger) With(key, value string) interface{}               { return m }

type mockDockerClient struct {
	pingCalled bool
	pingErr    error
	containers []entity.Container
}

func (m *mockDockerClient) Ping(ctx context.Context) error {
	m.pingCalled = true
	return m.pingErr
}

func (m *mockDockerClient) ContainerList(ctx context.Context) ([]entity.Container, error) {
	return m.containers, nil
}

func (m *mockDockerClient) ContainerLogs(ctx context.Context, containerID, tail string) (string, error) {
	return "log output", nil
}

func (m *mockDockerClient) ContainerExec(ctx context.Context, containerID string, cmd []string) (string, error) {
	return "exec output", nil
}

func (m *mockDockerClient) ContainerStart(ctx context.Context, containerID string) error {
	return nil
}

func (m *mockDockerClient) ContainerStop(ctx context.Context, containerID string, timeout time.Duration) error {
	return nil
}

func (m *mockDockerClient) ContainerRestart(ctx context.Context, containerID string, timeout time.Duration) error {
	return nil
}

func (m *mockDockerClient) ContainerRemove(ctx context.Context, containerID string, force bool) error {
	return nil
}

func (m *mockDockerClient) ContainerInspect(ctx context.Context, containerID string) (*interface{}, error) {
	return nil, nil
}

func (m *mockDockerClient) Close() error {
	return nil
}

func TestUseCaseCreate(t *testing.T) {
	repo := &mockRepo{machines: make(map[string]*entity.RemoteMachine)}
	logger := &mockLogger{}
	uc := New(repo, logger)

	machine := &entity.RemoteMachine{
		Name:       "test-server",
		Host:       "192.168.1.100",
		Port:       22,
		Username:   "root",
		AuthMethod: entity.AuthMethodPassword,
		Password:   "secret",
	}

	result, err := uc.Create(context.Background(), machine)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if result.ID == "" {
		t.Error("expected ID to be set")
	}
	if result.Port != 22 {
		t.Errorf("expected default port 22, got %d", result.Port)
	}
	if result.Status != "unknown" {
		t.Errorf("expected status unknown, got %s", result.Status)
	}
}

func TestUseCaseGetByID(t *testing.T) {
	repo := &mockRepo{machines: make(map[string]*entity.RemoteMachine)}
	logger := &mockLogger{}
	uc := New(repo, logger)

	machine := &entity.RemoteMachine{
		ID:         uuid.New().String(),
		Name:       "test-server",
		Host:       "192.168.1.100",
		Username:   "root",
		AuthMethod: entity.AuthMethodPassword,
	}
	repo.Create(context.Background(), machine)

	result, err := uc.GetByID(context.Background(), machine.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if result.Name != machine.Name {
		t.Errorf("expected name %s, got %s", machine.Name, result.Name)
	}
}

func TestUseCaseList(t *testing.T) {
	repo := &mockRepo{machines: make(map[string]*entity.RemoteMachine)}
	logger := &mockLogger{}
	uc := New(repo, logger)

	for range 3 {
		uc.Create(context.Background(), &entity.RemoteMachine{
			Name:       "server",
			Host:       "192.168.1.100",
			Username:   "root",
			AuthMethod: entity.AuthMethodPassword,
		})
	}

	machines, err := uc.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(machines) != 3 {
		t.Errorf("expected 3 machines, got %d", len(machines))
	}
}

func TestUseCaseUpdate(t *testing.T) {
	repo := &mockRepo{machines: make(map[string]*entity.RemoteMachine)}
	logger := &mockLogger{}
	uc := New(repo, logger)

	machine, _ := uc.Create(context.Background(), &entity.RemoteMachine{
		Name:       "original",
		Host:       "192.168.1.100",
		Username:   "root",
		AuthMethod: entity.AuthMethodPassword,
	})

	machine.Name = "updated"
	err := uc.Update(context.Background(), machine)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	result, _ := uc.GetByID(context.Background(), machine.ID)
	if result.Name != "updated" {
		t.Errorf("expected updated name, got %s", result.Name)
	}
}

func TestUseCaseDelete(t *testing.T) {
	repo := &mockRepo{machines: make(map[string]*entity.RemoteMachine)}
	logger := &mockLogger{}
	uc := New(repo, logger)

	machine, _ := uc.Create(context.Background(), &entity.RemoteMachine{
		Name:       "to-delete",
		Host:       "192.168.1.100",
		Username:   "root",
		AuthMethod: entity.AuthMethodPassword,
	})

	err := uc.Delete(context.Background(), machine.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = uc.GetByID(context.Background(), machine.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestUseCaseGetByHost(t *testing.T) {
	repo := &mockRepo{machines: make(map[string]*entity.RemoteMachine)}
	logger := &mockLogger{}
	uc := New(repo, logger)

	uc.Create(context.Background(), &entity.RemoteMachine{
		Name:       "test-server",
		Host:       "192.168.1.100",
		Username:   "root",
		AuthMethod: entity.AuthMethodPassword,
	})

	result, err := uc.GetByHost(context.Background(), "192.168.1.100")
	if err != nil {
		t.Fatalf("GetByHost failed: %v", err)
	}

	if result.Host != "192.168.1.100" {
		t.Errorf("expected host 192.168.1.100, got %s", result.Host)
	}
}

func TestUseCaseCreateWithDefaults(t *testing.T) {
	repo := &mockRepo{machines: make(map[string]*entity.RemoteMachine)}
	logger := &mockLogger{}
	uc := New(repo, logger)

	machine := &entity.RemoteMachine{
		Name:       "test",
		Host:       "localhost",
		Username:   "user",
		AuthMethod: entity.AuthMethodKey,
		SSHKey:     "key-content",
	}

	result, err := uc.Create(context.Background(), machine)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if result.ID == "" {
		t.Error("expected ID to be generated")
	}
	if result.Port != 22 {
		t.Errorf("expected default port 22, got %d", result.Port)
	}
	if result.DockerHost != "/var/run/docker.sock" {
		t.Errorf("expected default docker host, got %s", result.DockerHost)
	}
	if result.Status != "unknown" {
		t.Errorf("expected unknown status, got %s", result.Status)
	}
}
