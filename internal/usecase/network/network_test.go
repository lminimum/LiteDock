package network

import (
	"context"
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

// --- mockLogger ---

type mockLogger struct{}

func (m *mockLogger) Debug(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Info(_ string, _ ...interface{})       {}
func (m *mockLogger) Warn(_ string, _ ...interface{})       {}
func (m *mockLogger) Error(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Fatal(_ interface{}, _ ...interface{}) {}

// --- mockNetworkRepo ---

type mockNetworkRepo struct {
	listByMachineFn   func(ctx context.Context, machineID string) ([]entity.Network, error)
	getByNameFn       func(ctx context.Context, machineID, name string) (*entity.Network, error)
	upsertBatchFn     func(ctx context.Context, machineID string, networks []entity.Network) error
	deleteByMachineFn func(ctx context.Context, machineID string) error
	isCacheValidFn    func(ctx context.Context, machineID string, maxAge time.Duration) (bool, error)
}

func (m *mockNetworkRepo) ListByMachine(ctx context.Context, machineID string) ([]entity.Network, error) {
	return m.listByMachineFn(ctx, machineID)
}

func (m *mockNetworkRepo) GetByName(ctx context.Context, machineID, name string) (*entity.Network, error) {
	return m.getByNameFn(ctx, machineID, name)
}

func (m *mockNetworkRepo) UpsertBatch(ctx context.Context, machineID string, networks []entity.Network) error {
	if m.upsertBatchFn != nil {
		return m.upsertBatchFn(ctx, machineID, networks)
	}
	return nil
}

func (m *mockNetworkRepo) DeleteByMachine(ctx context.Context, machineID string) error {
	if m.deleteByMachineFn != nil {
		return m.deleteByMachineFn(ctx, machineID)
	}
	return nil
}

func (m *mockNetworkRepo) IsCacheValid(ctx context.Context, machineID string, maxAge time.Duration) (bool, error) {
	if m.isCacheValidFn != nil {
		return m.isCacheValidFn(ctx, machineID, maxAge)
	}
	return true, nil
}

var _ repo.NetworkRepo = (*mockNetworkRepo)(nil)

// --- mockRemoteMachineRepo ---

type mockRemoteMachineRepo struct {
	getByIDFn func(ctx context.Context, id string) (*entity.RemoteMachine, error)
}

func (m *mockRemoteMachineRepo) GetByID(ctx context.Context, id string) (*entity.RemoteMachine, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockRemoteMachineRepo) Create(_ context.Context, _ *entity.RemoteMachine) error { return nil }

func (m *mockRemoteMachineRepo) List(_ context.Context) ([]entity.RemoteMachine, error) {
	return nil, nil
}

func (m *mockRemoteMachineRepo) Count(_ context.Context) (int64, error) { return 0, nil }

func (m *mockRemoteMachineRepo) Update(_ context.Context, _ *entity.RemoteMachine) error { return nil }

func (m *mockRemoteMachineRepo) Delete(_ context.Context, _ string) error { return nil }

func (m *mockRemoteMachineRepo) GetByHost(_ context.Context, _ string) (*entity.RemoteMachine, error) {
	return nil, nil
}

func (m *mockRemoteMachineRepo) UpdateStatus(_ context.Context, _ string, _ string) error {
	return nil
}

var _ repo.RemoteMachineRepo = (*mockRemoteMachineRepo)(nil)

// --- mockDockerClient ---

type mockDockerClient struct {
	networkListFn    func(ctx context.Context) ([]entity.Network, error)
	networkCreateFn  func(ctx context.Context, name, driver string) (*entity.Network, error)
	networkDeleteFn  func(ctx context.Context, networkID string) error
	networkInspectFn func(ctx context.Context, networkID string) (*entity.Network, error)
	closeFn          func() error
}

func (m *mockDockerClient) Ping(_ context.Context) error { return nil }
func (m *mockDockerClient) ContainerList(_ context.Context) ([]entity.Container, error) {
	return nil, nil
}

func (m *mockDockerClient) ContainerLogs(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func (m *mockDockerClient) ContainerExec(_ context.Context, _ string, _ []string) (string, error) {
	return "", nil
}

func (m *mockDockerClient) ContainerStart(_ context.Context, _ string) error { return nil }

func (m *mockDockerClient) ContainerStop(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

func (m *mockDockerClient) ContainerRestart(_ context.Context, _ string, _ time.Duration) error {
	return nil
}
func (m *mockDockerClient) ContainerRemove(_ context.Context, _ string, _ bool) error { return nil }
func (m *mockDockerClient) ContainerInspect(_ context.Context, _ string) (*container.InspectResponse, error) {
	return nil, nil
}

func (m *mockDockerClient) ContainerCreate(_ context.Context, _ *container.Config, _ *container.HostConfig, _ string) (*container.CreateResponse, error) {
	return nil, nil
}

func (m *mockDockerClient) NetworkList(ctx context.Context) ([]entity.Network, error) {
	if m.networkListFn != nil {
		return m.networkListFn(ctx)
	}
	return nil, nil
}

func (m *mockDockerClient) NetworkCreate(ctx context.Context, name, driver string) (*entity.Network, error) {
	if m.networkCreateFn != nil {
		return m.networkCreateFn(ctx, name, driver)
	}
	return nil, nil
}

func (m *mockDockerClient) NetworkDelete(ctx context.Context, networkID string) error {
	if m.networkDeleteFn != nil {
		return m.networkDeleteFn(ctx, networkID)
	}
	return nil
}

func (m *mockDockerClient) NetworkInspect(ctx context.Context, networkID string) (*entity.Network, error) {
	if m.networkInspectFn != nil {
		return m.networkInspectFn(ctx, networkID)
	}
	return nil, nil
}

func (m *mockDockerClient) VolumeList(_ context.Context) ([]entity.Volume, error) { return nil, nil }

func (m *mockDockerClient) VolumeCreate(_ context.Context, _, _ string) (*entity.Volume, error) {
	return nil, nil
}
func (m *mockDockerClient) VolumeDelete(_ context.Context, _ string) error { return nil }
func (m *mockDockerClient) VolumeInspect(_ context.Context, _ string) (*entity.Volume, error) {
	return nil, nil
}

func (m *mockDockerClient) ImageList(_ context.Context, _ dockerImage.ListOptions) ([]entity.Image, error) {
	return nil, nil
}

func (m *mockDockerClient) ImagePull(_ context.Context, _ string, _ dockerImage.PullOptions) error {
	return nil
}

func (m *mockDockerClient) ImageRemove(_ context.Context, _ string, _ dockerImage.RemoveOptions) ([]dockerImage.DeleteResponse, error) {
	return nil, nil
}

func (m *mockDockerClient) ImageInspect(_ context.Context, _ string) (dockerImage.InspectResponse, error) {
	return dockerImage.InspectResponse{}, nil
}

func (m *mockDockerClient) ImagePrune(_ context.Context, _ filters.Args) (dockerImage.PruneReport, error) {
	return dockerImage.PruneReport{}, nil
}

func (m *mockDockerClient) ComposeUp(_ context.Context, _, _, _ string) error { return nil }
func (m *mockDockerClient) ComposeDown(_ context.Context, _, _ string, _ bool) error {
	return nil
}

func (m *mockDockerClient) ComposeBuild(_ context.Context, _, _ string) error { return nil }
func (m *mockDockerClient) ComposeStart(_ context.Context, _, _ string) error { return nil }
func (m *mockDockerClient) ComposeStop(_ context.Context, _, _ string) error  { return nil }
func (m *mockDockerClient) ComposeRestart(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockDockerClient) ComposePs(_ context.Context, _, _ string) ([]dockerclient.ComposeServiceStatus, error) {
	return nil, nil
}

func (m *mockDockerClient) ComposeLogs(_ context.Context, _, _ string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockDockerClient) ComposeConfig(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func (m *mockDockerClient) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

var _ dockerclient.Client = (*mockDockerClient)(nil)

// --- helpers ---

func newUseCaseForTest(cli dockerclient.Client, netRepo repo.NetworkRepo, rmRepo repo.RemoteMachineRepo) *NetworkUseCase {
	uc := New(netRepo, rmRepo, 10*time.Second, &mockLogger{})
	uc.testDockerClient = cli
	return uc
}

// newUseCaseWithoutDockerClient creates a usecase without injecting a test docker client,
// so getDockerClient will exercise the production path (remoteMachineRepo.GetByID).
func newUseCaseWithoutDockerClient(netRepo repo.NetworkRepo, rmRepo repo.RemoteMachineRepo) *NetworkUseCase {
	return New(netRepo, rmRepo, 10*time.Second, &mockLogger{})
}

// --- tests ---

func TestListNetworks_EmptyCache_FetchesFromDocker(t *testing.T) {
	dockerCalled := false
	upsertCalled := false

	mockCli := &mockDockerClient{
		networkListFn: func(_ context.Context) ([]entity.Network, error) {
			dockerCalled = true
			return []entity.Network{
				{Name: "my-network", Driver: "bridge", Scope: "local"},
			}, nil
		},
	}

	mockNetRepo := &mockNetworkRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Network, error) {
			return nil, nil // empty cache
		},
		upsertBatchFn: func(_ context.Context, _ string, networks []entity.Network) error {
			upsertCalled = true
			require.Len(t, networks, 1)
			require.Equal(t, "my-network", networks[0].Name)
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockNetRepo, &mockRemoteMachineRepo{})

	networks, err := uc.ListNetworks(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, networks, 1)
	require.Equal(t, "my-network", networks[0].Name)
	require.True(t, dockerCalled, "expected Docker to be called for empty cache")
	require.True(t, upsertCalled, "expected UpsertBatch to be called after fetch")
}

func TestListNetworks_CacheHit_ReturnsCached(t *testing.T) {
	dockerCalled := false

	mockCli := &mockDockerClient{
		networkListFn: func(_ context.Context) ([]entity.Network, error) {
			dockerCalled = true
			return nil, nil
		},
	}

	cachedNetworks := []entity.Network{
		{Name: "cached-net", Driver: "bridge"},
	}

	mockNetRepo := &mockNetworkRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Network, error) {
			return cachedNetworks, nil
		},
		isCacheValidFn: func(_ context.Context, _ string, _ time.Duration) (bool, error) {
			return true, nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockNetRepo, &mockRemoteMachineRepo{})

	networks, err := uc.ListNetworks(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, networks, 1)
	require.Equal(t, "cached-net", networks[0].Name)
	require.False(t, dockerCalled, "expected Docker NOT to be called when cache is valid")
}

func TestListNetworks_StaleCache_TriggersRefresh(t *testing.T) {
	upsertCalled := make(chan struct{}, 1)

	mockCli := &mockDockerClient{
		networkListFn: func(_ context.Context) ([]entity.Network, error) {
			return []entity.Network{
				{Name: "refreshed-net", Driver: "bridge"},
			}, nil
		},
	}

	cachedNetworks := []entity.Network{
		{Name: "stale-net", Driver: "bridge"},
	}

	mockNetRepo := &mockNetworkRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Network, error) {
			return cachedNetworks, nil
		},
		isCacheValidFn: func(_ context.Context, _ string, _ time.Duration) (bool, error) {
			return false, nil // stale
		},
		upsertBatchFn: func(_ context.Context, _ string, networks []entity.Network) error {
			require.Len(t, networks, 1)
			require.Equal(t, "refreshed-net", networks[0].Name)
			upsertCalled <- struct{}{}
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockNetRepo, &mockRemoteMachineRepo{})

	// First call: returns cached (stale), triggers background refresh
	networks, err := uc.ListNetworks(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, networks, 1)
	require.Equal(t, "stale-net", networks[0].Name)

	// Wait for the goroutine to complete
	select {
	case <-upsertCalled:
		// goroutine ran and upserted refreshed data
	case <-time.After(2 * time.Second):
		t.Fatal("refresh goroutine did not call UpsertBatch within 2s")
	}
}

func TestCreateNetwork_Success(t *testing.T) {
	createCalled := false
	cacheInvalidated := false

	mockCli := &mockDockerClient{
		networkCreateFn: func(_ context.Context, name, driver string) (*entity.Network, error) {
			createCalled = true
			require.Equal(t, "test-net", name)
			require.Equal(t, "overlay", driver)
			return &entity.Network{Name: "test-net", Driver: "overlay", Scope: "local"}, nil
		},
	}

	mockNetRepo := &mockNetworkRepo{
		deleteByMachineFn: func(_ context.Context, machineID string) error {
			cacheInvalidated = true
			require.Equal(t, "local", machineID)
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockNetRepo, &mockRemoteMachineRepo{})

	network, err := uc.CreateNetwork(context.Background(), "local", "test-net", "overlay")
	require.NoError(t, err)
	require.NotNil(t, network)
	require.Equal(t, "test-net", network.Name)
	require.True(t, createCalled, "expected Docker NetworkCreate to be called")
	require.True(t, cacheInvalidated, "expected cache invalidation on create")
}

func TestDeleteNetwork_Success(t *testing.T) {
	deleteCalled := false
	cacheInvalidated := false

	mockCli := &mockDockerClient{
		networkDeleteFn: func(_ context.Context, networkID string) error {
			deleteCalled = true
			require.Equal(t, "user-network", networkID)
			return nil
		},
	}

	mockNetRepo := &mockNetworkRepo{
		deleteByMachineFn: func(_ context.Context, machineID string) error {
			cacheInvalidated = true
			require.Equal(t, "local", machineID)
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockNetRepo, &mockRemoteMachineRepo{})

	err := uc.DeleteNetwork(context.Background(), "local", "user-network")
	require.NoError(t, err)
	require.True(t, deleteCalled, "expected Docker NetworkDelete to be called")
	require.True(t, cacheInvalidated, "expected cache invalidation on delete")
}

func TestDeleteNetwork_BuiltIn_Error(t *testing.T) {
	dockerCalled := false

	mockCli := &mockDockerClient{
		networkDeleteFn: func(_ context.Context, _ string) error {
			dockerCalled = true
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, &mockNetworkRepo{}, &mockRemoteMachineRepo{})

	builtInNames := []string{"bridge", "host", "none"}
	for _, name := range builtInNames {
		err := uc.DeleteNetwork(context.Background(), "local", name)
		require.Error(t, err, "expected error for built-in network %q", name)
		require.ErrorIs(t, err, errors.ErrInvalidInput, "expected ErrInvalidInput for built-in network %q", name)
	}

	require.False(t, dockerCalled, "expected Docker NOT to be called for built-in networks")
}

func TestInspectNetwork_Success(t *testing.T) {
	inspectCalled := false

	mockCli := &mockDockerClient{
		networkInspectFn: func(_ context.Context, networkID string) (*entity.Network, error) {
			inspectCalled = true
			require.Equal(t, "my-net", networkID)
			return &entity.Network{Name: "my-net", Driver: "bridge", Scope: "local"}, nil
		},
	}

	uc := newUseCaseForTest(mockCli, &mockNetworkRepo{}, &mockRemoteMachineRepo{})

	network, err := uc.InspectNetwork(context.Background(), "local", "my-net")
	require.NoError(t, err)
	require.NotNil(t, network)
	require.Equal(t, "my-net", network.Name)
	require.Equal(t, "bridge", network.Driver)
	require.True(t, inspectCalled, "expected Docker NetworkInspect to be called")
}

func TestListNetworks_IsCacheValidError_StillReturnsCached(t *testing.T) {
	mockCli := &mockDockerClient{
		networkListFn: func(_ context.Context) ([]entity.Network, error) {
			return nil, nil
		},
	}

	cachedNetworks := []entity.Network{
		{Name: "cached-despite-error", Driver: "bridge"},
	}

	mockNetRepo := &mockNetworkRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Network, error) {
			return cachedNetworks, nil
		},
		isCacheValidFn: func(_ context.Context, _ string, _ time.Duration) (bool, error) {
			return false, errors.ErrVolumeNotFound
		},
	}

	uc := newUseCaseForTest(mockCli, mockNetRepo, &mockRemoteMachineRepo{})

	networks, err := uc.ListNetworks(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, networks, 1)
	require.Equal(t, "cached-despite-error", networks[0].Name)
}

func TestListNetworks_EmptyDocker_NoUpsert(t *testing.T) {
	upsertCalled := false

	mockCli := &mockDockerClient{
		networkListFn: func(_ context.Context) ([]entity.Network, error) {
			return []entity.Network{}, nil
		},
	}

	mockNetRepo := &mockNetworkRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Network, error) {
			return nil, nil
		},
		upsertBatchFn: func(_ context.Context, _ string, _ []entity.Network) error {
			upsertCalled = true
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockNetRepo, &mockRemoteMachineRepo{})

	networks, err := uc.ListNetworks(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, networks, 0)
	require.False(t, upsertCalled, "expected no UpsertBatch when Docker returns empty list")
}

func TestListNetworks_UpsertBatchError_ReturnsNetworks(t *testing.T) {
	mockCli := &mockDockerClient{
		networkListFn: func(_ context.Context) ([]entity.Network, error) {
			return []entity.Network{
				{Name: "error-but-returned", Driver: "bridge"},
			}, nil
		},
	}

	mockNetRepo := &mockNetworkRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Network, error) {
			return nil, nil
		},
		upsertBatchFn: func(_ context.Context, _ string, _ []entity.Network) error {
			return errors.ErrVolumeNotFound
		},
	}

	uc := newUseCaseForTest(mockCli, mockNetRepo, &mockRemoteMachineRepo{})

	networks, err := uc.ListNetworks(context.Background(), "local")
	require.Error(t, err)
	require.Len(t, networks, 1)
	require.Equal(t, "error-but-returned", networks[0].Name)
}

func TestListNetworks_StaleCache_RefreshError(t *testing.T) {
	refreshAttempted := make(chan struct{}, 1)

	mockCli := &mockDockerClient{
		networkListFn: func(_ context.Context) ([]entity.Network, error) {
			refreshAttempted <- struct{}{}
			return nil, errors.ErrVolumeNotFound
		},
	}

	cachedNetworks := []entity.Network{
		{Name: "stale-net-err", Driver: "bridge"},
	}

	mockNetRepo := &mockNetworkRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Network, error) {
			return cachedNetworks, nil
		},
		isCacheValidFn: func(_ context.Context, _ string, _ time.Duration) (bool, error) {
			return false, nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockNetRepo, &mockRemoteMachineRepo{})

	networks, err := uc.ListNetworks(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, networks, 1)
	require.Equal(t, "stale-net-err", networks[0].Name)

	select {
	case <-refreshAttempted:
	case <-time.After(2 * time.Second):
		t.Fatal("refresh goroutine did not run within 2s")
	}
}

func TestListNetworks_GetDockerClientError(t *testing.T) {
	rmRepo := &mockRemoteMachineRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.RemoteMachine, error) {
			return nil, errors.ErrVolumeNotFound
		},
	}

	mockNetRepo := &mockNetworkRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Network, error) {
			return nil, nil
		},
	}

	uc := newUseCaseWithoutDockerClient(mockNetRepo, rmRepo)

	networks, err := uc.ListNetworks(context.Background(), "some-machine")
	require.Error(t, err)
	require.Nil(t, networks)
}

func TestCreateNetwork_GetDockerClientError(t *testing.T) {
	rmRepo := &mockRemoteMachineRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.RemoteMachine, error) {
			return nil, errors.ErrVolumeNotFound
		},
	}

	uc := newUseCaseWithoutDockerClient(&mockNetworkRepo{}, rmRepo)

	network, err := uc.CreateNetwork(context.Background(), "some-machine", "test-net", "bridge")
	require.Error(t, err)
	require.Nil(t, network)
}

func TestDeleteNetwork_GetDockerClientError(t *testing.T) {
	rmRepo := &mockRemoteMachineRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.RemoteMachine, error) {
			return nil, errors.ErrVolumeNotFound
		},
	}

	uc := newUseCaseWithoutDockerClient(&mockNetworkRepo{}, rmRepo)

	err := uc.DeleteNetwork(context.Background(), "some-machine", "user-network")
	require.Error(t, err)
}

func TestInspectNetwork_GetDockerClientError(t *testing.T) {
	rmRepo := &mockRemoteMachineRepo{
		getByIDFn: func(_ context.Context, _ string) (*entity.RemoteMachine, error) {
			return nil, errors.ErrVolumeNotFound
		},
	}

	uc := newUseCaseWithoutDockerClient(&mockNetworkRepo{}, rmRepo)

	network, err := uc.InspectNetwork(context.Background(), "some-machine", "my-net")
	require.Error(t, err)
	require.Nil(t, network)
}

func TestInspectNetwork_NotFound(t *testing.T) {
	mockCli := &mockDockerClient{
		networkInspectFn: func(_ context.Context, _ string) (*entity.Network, error) {
			return nil, errors.ErrVolumeNotFound
		},
	}

	uc := newUseCaseForTest(mockCli, &mockNetworkRepo{}, &mockRemoteMachineRepo{})

	network, err := uc.InspectNetwork(context.Background(), "local", "non-existent-net")
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrVolumeNotFound)
	require.Nil(t, network)
}
