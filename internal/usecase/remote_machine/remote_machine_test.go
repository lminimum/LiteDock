package remote_machine

import (
	"context"
	"testing"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/stretchr/testify/require"
)

type mockRemoteMachineRepo struct {
	getByIDFn  func(ctx context.Context, id string) (*entity.RemoteMachine, error)
	listFn     func(ctx context.Context) ([]entity.RemoteMachine, error)
	createFn   func(ctx context.Context, m *entity.RemoteMachine) error
	updateFn   func(ctx context.Context, m *entity.RemoteMachine) error
	deleteFn   func(ctx context.Context, id string) error
	getByHostFn func(ctx context.Context, host string) (*entity.RemoteMachine, error)
	countFn    func(ctx context.Context) (int64, error)
}

func (m *mockRemoteMachineRepo) GetByID(ctx context.Context, id string) (*entity.RemoteMachine, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockRemoteMachineRepo) List(ctx context.Context) ([]entity.RemoteMachine, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockRemoteMachineRepo) Create(ctx context.Context, machine *entity.RemoteMachine) error {
	if m.createFn != nil {
		return m.createFn(ctx, machine)
	}
	return nil
}

func (m *mockRemoteMachineRepo) Update(ctx context.Context, machine *entity.RemoteMachine) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, machine)
	}
	return nil
}

func (m *mockRemoteMachineRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockRemoteMachineRepo) GetByHost(ctx context.Context, host string) (*entity.RemoteMachine, error) {
	if m.getByHostFn != nil {
		return m.getByHostFn(ctx, host)
	}
	return nil, nil
}

func (m *mockRemoteMachineRepo) Count(ctx context.Context) (int64, error) {
	if m.countFn != nil {
		return m.countFn(ctx)
	}
	return 0, nil
}

type mockLogger struct{}

func (m *mockLogger) Debug(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Info(_ string, _ ...interface{})       {}
func (m *mockLogger) Warn(_ string, _ ...interface{})       {}
func (m *mockLogger) Error(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Fatal(_ interface{}, _ ...interface{}) {}

func TestIsLocalMachine(t *testing.T) {
	tests := []struct {
		name     string
		machine  *entity.RemoteMachine
		expected bool
	}{
		{
			name:     "local with any host",
			machine:  &entity.RemoteMachine{ID: LocalMachineID, Host: "anything"},
			expected: true,
		},
		{
			name:     "local with localhost",
			machine:  &entity.RemoteMachine{ID: LocalMachineID, Host: LocalMachineHost},
			expected: true,
		},
		{
			name:     "non-local with localhost",
			machine:  &entity.RemoteMachine{ID: "not-local", Host: LocalMachineHost},
			expected: false,
		},
		{
			name:     "non-local with remote host",
			machine:  &entity.RemoteMachine{ID: "not-local", Host: "remote.example.com"},
			expected: false,
		},
		{
			name:     "case-sensitive - LOCAL is not local",
			machine:  &entity.RemoteMachine{ID: "LOCAL", Host: LocalMachineHost},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLocalMachine(tt.machine)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDockerClient_LocalMachine(t *testing.T) {
	mockRepo := &mockRemoteMachineRepo{
		getByIDFn: func(_ context.Context, id string) (*entity.RemoteMachine, error) {
			return &entity.RemoteMachine{
				ID:         LocalMachineID,
				Name:       "local",
				Host:       LocalMachineHost,
				Port:       0,
				Username:   "local",
				AuthMethod: entity.AuthMethodPassword,
				DockerHost: "/var/run/docker.sock",
			}, nil
		},
	}

	uc := &UseCase{
		repo: mockRepo,
		l:    &mockLogger{},
	}

	cli, err := uc.getDockerClient(context.Background(), LocalMachineID)
	require.NoError(t, err)
	require.NotNil(t, cli)
	cli.Close()
}

func TestGetDockerClient_RemoteMachine(t *testing.T) {
	mockRepo := &mockRemoteMachineRepo{
		getByIDFn: func(_ context.Context, id string) (*entity.RemoteMachine, error) {
			return &entity.RemoteMachine{
				ID:         "remote-uuid",
				Name:       "remote-server",
				Host:       "192.168.1.100",
				Port:       22,
				Username:   "root",
				AuthMethod: entity.AuthMethodPassword,
				Password:   "bad-password",
				DockerHost: "/var/run/docker.sock",
			}, nil
		},
	}

	uc := &UseCase{
		repo: mockRepo,
		l:    &mockLogger{},
	}

	cli, err := uc.getDockerClient(context.Background(), "remote-uuid")
	require.Error(t, err)
	require.Nil(t, cli)
	require.Contains(t, err.Error(), "UseCase.getDockerClient.sshclient.New")
}

func TestDelete_LocalMachineRejected(t *testing.T) {
	uc := &UseCase{
		repo: &mockRemoteMachineRepo{},
		l:    &mockLogger{},
	}

	err := uc.Delete(context.Background(), LocalMachineID)
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrInvalidInput)
}

func TestDelete_RemoteMachineAllowed(t *testing.T) {
	deleted := false

	mockRepo := &mockRemoteMachineRepo{
		deleteFn: func(_ context.Context, id string) error {
			deleted = true
			require.Equal(t, "remote-uuid", id)
			return nil
		},
	}

	uc := &UseCase{
		repo: mockRepo,
		l:    &mockLogger{},
	}

	err := uc.Delete(context.Background(), "remote-uuid")
	require.NoError(t, err)
	require.True(t, deleted, "expected deleteFn to be called for remote machine")
}

func TestUpdate_LocalMachineRejected(t *testing.T) {
	uc := &UseCase{
		repo: &mockRemoteMachineRepo{},
		l:    &mockLogger{},
	}

	m := &entity.RemoteMachine{ID: LocalMachineID}
	err := uc.Update(context.Background(), m)
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrInvalidInput)
}

func TestUpdate_RemoteMachineAllowed(t *testing.T) {
	updated := false

	mockRepo := &mockRemoteMachineRepo{
		updateFn: func(_ context.Context, m *entity.RemoteMachine) error {
			updated = true
			require.Equal(t, "remote-uuid", m.ID)
			return nil
		},
	}

	uc := &UseCase{
		repo: mockRepo,
		l:    &mockLogger{},
	}

	err := uc.Update(context.Background(), &entity.RemoteMachine{ID: "remote-uuid"})
	require.NoError(t, err)
	require.True(t, updated, "expected updateFn to be called for remote machine")
}
