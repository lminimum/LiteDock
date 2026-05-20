package remote_machine

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerImage "github.com/docker/docker/api/types/image"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/pkg/dockerclient"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/stretchr/testify/require"
)

type mockRemoteMachineRepo struct {
	getByIDFn   func(ctx context.Context, id string) (*entity.RemoteMachine, error)
	listFn      func(ctx context.Context) ([]entity.RemoteMachine, error)
	createFn    func(ctx context.Context, m *entity.RemoteMachine) error
	updateFn    func(ctx context.Context, m *entity.RemoteMachine) error
	deleteFn    func(ctx context.Context, id string) error
	getByHostFn func(ctx context.Context, host string) (*entity.RemoteMachine, error)
	countFn     func(ctx context.Context) (int64, error)
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
	return nil, nil
}

func (m *mockRemoteMachineRepo) UpdateStatus(_ context.Context, _ string, _ string) error {
	return nil
}

func (m *mockRemoteMachineRepo) Count(ctx context.Context) (int64, error) {
	if m.countFn != nil {
		return m.countFn(ctx)
	}
	return 0, nil
}

var _ repo.RemoteMachineRepo = (*mockRemoteMachineRepo)(nil)

type mockContainerRepo struct {
	deleteByMachineFn func(ctx context.Context, machineID string) error
}

func (m *mockContainerRepo) List(context.Context) ([]entity.Container, error) { return nil, nil }

func (m *mockContainerRepo) Get(context.Context, string) (*entity.Container, error) { return nil, nil }

func (m *mockContainerRepo) ListByMachine(context.Context, string) ([]entity.Container, error) {
	return nil, nil
}

func (m *mockContainerRepo) UpsertBatch(context.Context, string, []entity.Container) error {
	return nil
}

func (m *mockContainerRepo) DeleteByMachine(ctx context.Context, machineID string) error {
	if m.deleteByMachineFn != nil {
		return m.deleteByMachineFn(ctx, machineID)
	}
	return nil
}

func (m *mockContainerRepo) IsCacheValid(context.Context, string, time.Duration) (bool, error) {
	return true, nil
}

func (m *mockContainerRepo) CountAll(context.Context) (int64, error) { return 0, nil }

func (m *mockContainerRepo) CountByStatus(context.Context, string) (int64, error) { return 0, nil }

var _ repo.ContainerRepo = (*mockContainerRepo)(nil)

type mockDockerClient struct {
	containerCreateFn func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, containerName string) (*container.CreateResponse, error)
	imagePullFn       func(ctx context.Context, ref string, opts dockerImage.PullOptions) error
	imageInspectFn    func(ctx context.Context, id string) (dockerImage.InspectResponse, error)
	closeFn           func() error
}

func (m *mockDockerClient) Ping(context.Context) error { return nil }

func (m *mockDockerClient) ContainerList(context.Context) ([]entity.Container, error) {
	return nil, nil
}

func (m *mockDockerClient) ContainerLogs(context.Context, string, string) (string, error) {
	return "", nil
}

func (m *mockDockerClient) ContainerExec(context.Context, string, []string) (string, error) {
	return "", nil
}

func (m *mockDockerClient) ContainerStart(context.Context, string) error { return nil }

func (m *mockDockerClient) ContainerStop(context.Context, string, time.Duration) error { return nil }

func (m *mockDockerClient) ContainerRestart(context.Context, string, time.Duration) error { return nil }

func (m *mockDockerClient) ContainerRemove(context.Context, string, bool) error { return nil }

func (m *mockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, containerName string) (*container.CreateResponse, error) {
	if m.containerCreateFn != nil {
		return m.containerCreateFn(ctx, config, hostConfig, containerName)
	}
	return &container.CreateResponse{ID: "created"}, nil
}

func (m *mockDockerClient) ContainerInspect(context.Context, string) (*container.InspectResponse, error) {
	return nil, nil
}

func (m *mockDockerClient) NetworkList(context.Context) ([]entity.Network, error) { return nil, nil }

func (m *mockDockerClient) NetworkCreate(context.Context, string, string) (*entity.Network, error) {
	return nil, nil
}

func (m *mockDockerClient) NetworkDelete(context.Context, string) error { return nil }

func (m *mockDockerClient) NetworkInspect(context.Context, string) (*entity.Network, error) {
	return nil, nil
}

func (m *mockDockerClient) VolumeList(context.Context) ([]entity.Volume, error) { return nil, nil }

func (m *mockDockerClient) VolumeCreate(context.Context, string, string) (*entity.Volume, error) {
	return nil, nil
}

func (m *mockDockerClient) VolumeDelete(context.Context, string) error { return nil }

func (m *mockDockerClient) VolumeInspect(context.Context, string) (*entity.Volume, error) {
	return nil, nil
}

func (m *mockDockerClient) ImageList(context.Context, dockerImage.ListOptions) ([]entity.Image, error) {
	return nil, nil
}

func (m *mockDockerClient) ImagePull(ctx context.Context, ref string, opts dockerImage.PullOptions) error {
	if m.imagePullFn != nil {
		return m.imagePullFn(ctx, ref, opts)
	}
	return nil
}

func (m *mockDockerClient) ImageRemove(context.Context, string, dockerImage.RemoveOptions) ([]dockerImage.DeleteResponse, error) {
	return nil, nil
}

func (m *mockDockerClient) ImageInspect(ctx context.Context, id string) (dockerImage.InspectResponse, error) {
	if m.imageInspectFn != nil {
		return m.imageInspectFn(ctx, id)
	}
	return dockerImage.InspectResponse{}, nil
}

func (m *mockDockerClient) ImagePrune(context.Context, filters.Args) (dockerImage.PruneReport, error) {
	return dockerImage.PruneReport{}, nil
}

func (m *mockDockerClient) ComposeUp(context.Context, string, string, string) error { return nil }

func (m *mockDockerClient) ComposeDown(context.Context, string, string, bool) error { return nil }

func (m *mockDockerClient) ComposeBuild(context.Context, string, string) error { return nil }

func (m *mockDockerClient) ComposeStart(context.Context, string, string) error { return nil }

func (m *mockDockerClient) ComposeStop(context.Context, string, string) error { return nil }

func (m *mockDockerClient) ComposeRestart(context.Context, string, string) error { return nil }

func (m *mockDockerClient) ComposePs(context.Context, string, string) ([]dockerclient.ComposeServiceStatus, error) {
	return nil, nil
}

func (m *mockDockerClient) ComposeLogs(context.Context, string, string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockDockerClient) ComposeConfig(context.Context, string, string) (string, error) {
	return "", nil
}

func (m *mockDockerClient) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

var _ dockerclient.Client = (*mockDockerClient)(nil)

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

func TestCreateContainer_PullsMissingImageBeforeCreate(t *testing.T) {
	calls := make([]string, 0, 4)
	cacheInvalidated := false

	mockCli := &mockDockerClient{
		imageInspectFn: func(_ context.Context, id string) (dockerImage.InspectResponse, error) {
			calls = append(calls, "inspect")
			require.Equal(t, "nginx:alpine", id)
			return dockerImage.InspectResponse{}, fmt.Errorf("image not found")
		},
		imagePullFn: func(_ context.Context, ref string, _ dockerImage.PullOptions) error {
			calls = append(calls, "pull")
			require.Equal(t, "nginx:alpine", ref)
			return nil
		},
		containerCreateFn: func(_ context.Context, cfg *container.Config, _ *container.HostConfig, name string) (*container.CreateResponse, error) {
			calls = append(calls, "create")
			require.Equal(t, "nginx:alpine", cfg.Image)
			require.Equal(t, "test-nginx", name)
			return &container.CreateResponse{ID: "container-id"}, nil
		},
	}

	uc := &UseCase{
		containerRepo: &mockContainerRepo{deleteByMachineFn: func(_ context.Context, machineID string) error {
			cacheInvalidated = true
			require.Equal(t, LocalMachineID, machineID)
			return nil
		}},
		l:                &mockLogger{},
		testDockerClient: mockCli,
	}

	resp, err := uc.CreateContainer(
		context.Background(),
		LocalMachineID,
		&container.Config{Image: "nginx:alpine"},
		&container.HostConfig{},
		"test-nginx",
	)

	require.NoError(t, err)
	require.Equal(t, "container-id", resp.ID)
	require.Equal(t, []string{"inspect", "pull", "create"}, calls)
	require.True(t, cacheInvalidated)
}

func TestCreateContainer_UsesExistingImageWithoutPull(t *testing.T) {
	pullCalled := false

	mockCli := &mockDockerClient{
		imageInspectFn: func(_ context.Context, id string) (dockerImage.InspectResponse, error) {
			require.Equal(t, "nginx:alpine", id)
			return dockerImage.InspectResponse{}, nil
		},
		imagePullFn: func(context.Context, string, dockerImage.PullOptions) error {
			pullCalled = true
			return nil
		},
	}

	uc := &UseCase{
		containerRepo:    &mockContainerRepo{},
		l:                &mockLogger{},
		testDockerClient: mockCli,
	}

	resp, err := uc.CreateContainer(
		context.Background(),
		LocalMachineID,
		&container.Config{Image: "nginx:alpine"},
		&container.HostConfig{},
		"test-nginx",
	)

	require.NoError(t, err)
	require.Equal(t, "created", resp.ID)
	require.False(t, pullCalled)
}
