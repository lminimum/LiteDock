package volume

import (
	"context"
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

// --- mockVolumeRepo ---

type mockVolumeRepo struct {
	listByMachineFn   func(ctx context.Context, machineID string) ([]entity.Volume, error)
	getByNameFn       func(ctx context.Context, machineID, name string) (*entity.Volume, error)
	upsertBatchFn     func(ctx context.Context, machineID string, volumes []entity.Volume) error
	deleteByMachineFn func(ctx context.Context, machineID string) error
	isCacheValidFn    func(ctx context.Context, machineID string, maxAge time.Duration) (bool, error)
}

func (m *mockVolumeRepo) ListByMachine(ctx context.Context, machineID string) ([]entity.Volume, error) {
	return m.listByMachineFn(ctx, machineID)
}

func (m *mockVolumeRepo) GetByName(ctx context.Context, machineID, name string) (*entity.Volume, error) {
	return m.getByNameFn(ctx, machineID, name)
}

func (m *mockVolumeRepo) UpsertBatch(ctx context.Context, machineID string, volumes []entity.Volume) error {
	if m.upsertBatchFn != nil {
		return m.upsertBatchFn(ctx, machineID, volumes)
	}
	return nil
}

func (m *mockVolumeRepo) DeleteByMachine(ctx context.Context, machineID string) error {
	if m.deleteByMachineFn != nil {
		return m.deleteByMachineFn(ctx, machineID)
	}
	return nil
}

func (m *mockVolumeRepo) IsCacheValid(ctx context.Context, machineID string, maxAge time.Duration) (bool, error) {
	if m.isCacheValidFn != nil {
		return m.isCacheValidFn(ctx, machineID, maxAge)
	}
	return true, nil
}

var _ repo.VolumeRepo = (*mockVolumeRepo)(nil)

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
func (m *mockRemoteMachineRepo) Count(_ context.Context) (int64, error)                  { return 0, nil }
func (m *mockRemoteMachineRepo) Update(_ context.Context, _ *entity.RemoteMachine) error { return nil }
func (m *mockRemoteMachineRepo) Delete(_ context.Context, _ string) error                { return nil }
func (m *mockRemoteMachineRepo) GetByHost(_ context.Context, _ string) (*entity.RemoteMachine, error) {
	return nil, nil
}

var _ repo.RemoteMachineRepo = (*mockRemoteMachineRepo)(nil)

// --- mockDockerClient ---

type mockDockerClient struct {
	volumeListFn    func(ctx context.Context) ([]entity.Volume, error)
	volumeCreateFn  func(ctx context.Context, name, driver string) (*entity.Volume, error)
	volumeDeleteFn  func(ctx context.Context, volumeID string) error
	volumeInspectFn func(ctx context.Context, volumeID string) (*entity.Volume, error)
	closeFn         func() error
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
func (m *mockDockerClient) NetworkList(_ context.Context) ([]entity.Network, error) { return nil, nil }
func (m *mockDockerClient) NetworkCreate(_ context.Context, _, _ string) (*entity.Network, error) {
	return nil, nil
}
func (m *mockDockerClient) NetworkDelete(_ context.Context, _ string) error { return nil }
func (m *mockDockerClient) NetworkInspect(_ context.Context, _ string) (*entity.Network, error) {
	return nil, nil
}

func (m *mockDockerClient) VolumeList(ctx context.Context) ([]entity.Volume, error) {
	if m.volumeListFn != nil {
		return m.volumeListFn(ctx)
	}
	return nil, nil
}

func (m *mockDockerClient) VolumeCreate(ctx context.Context, name, driver string) (*entity.Volume, error) {
	if m.volumeCreateFn != nil {
		return m.volumeCreateFn(ctx, name, driver)
	}
	return nil, nil
}

func (m *mockDockerClient) VolumeDelete(ctx context.Context, volumeID string) error {
	if m.volumeDeleteFn != nil {
		return m.volumeDeleteFn(ctx, volumeID)
	}
	return nil
}

func (m *mockDockerClient) VolumeInspect(ctx context.Context, volumeID string) (*entity.Volume, error) {
	if m.volumeInspectFn != nil {
		return m.volumeInspectFn(ctx, volumeID)
	}
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

func (m *mockDockerClient) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

var _ dockerclient.Client = (*mockDockerClient)(nil)

// --- helpers ---

func newUseCaseForTest(client dockerclient.Client, volRepo repo.VolumeRepo, rmRepo repo.RemoteMachineRepo) *VolumeUseCase {
	return &VolumeUseCase{
		volumeRepo:        volRepo,
		remoteMachineRepo: rmRepo,
		cacheMaxAge:       10 * time.Second,
		l:                 &mockLogger{},
		testDockerClient:  client,
	}
}

// --- tests ---

func TestListVolumes_EmptyCache_FetchesFromDocker(t *testing.T) {
	dockerCalled := false
	upsertCalled := false

	mockCli := &mockDockerClient{
		volumeListFn: func(_ context.Context) ([]entity.Volume, error) {
			dockerCalled = true
			return []entity.Volume{
				{Name: "my-volume", Driver: "local", Mountpoint: "/var/lib/docker/volumes/my-volume/_data"},
			}, nil
		},
	}

	mockVolRepo := &mockVolumeRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Volume, error) {
			return nil, nil // empty cache
		},
		upsertBatchFn: func(_ context.Context, _ string, volumes []entity.Volume) error {
			upsertCalled = true
			require.Len(t, volumes, 1)
			require.Equal(t, "my-volume", volumes[0].Name)
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockVolRepo, &mockRemoteMachineRepo{})

	volumes, err := uc.ListVolumes(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, volumes, 1)
	require.Equal(t, "my-volume", volumes[0].Name)
	require.True(t, dockerCalled, "expected Docker to be called for empty cache")
	require.True(t, upsertCalled, "expected UpsertBatch to be called after fetch")
}

func TestListVolumes_CacheHit_ReturnsCached(t *testing.T) {
	dockerCalled := false

	mockCli := &mockDockerClient{
		volumeListFn: func(_ context.Context) ([]entity.Volume, error) {
			dockerCalled = true
			return nil, nil
		},
	}

	cachedVolumes := []entity.Volume{
		{Name: "cached-vol", Driver: "local"},
	}

	mockVolRepo := &mockVolumeRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Volume, error) {
			return cachedVolumes, nil
		},
		isCacheValidFn: func(_ context.Context, _ string, _ time.Duration) (bool, error) {
			return true, nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockVolRepo, &mockRemoteMachineRepo{})

	volumes, err := uc.ListVolumes(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, volumes, 1)
	require.Equal(t, "cached-vol", volumes[0].Name)
	require.False(t, dockerCalled, "expected Docker NOT to be called when cache is valid")
}

func TestListVolumes_StaleCache_TriggersRefresh(t *testing.T) {
	upsertCalled := make(chan struct{}, 1)

	mockCli := &mockDockerClient{
		volumeListFn: func(_ context.Context) ([]entity.Volume, error) {
			return []entity.Volume{
				{Name: "refreshed-vol", Driver: "local"},
			}, nil
		},
	}

	cachedVolumes := []entity.Volume{
		{Name: "stale-vol", Driver: "local"},
	}

	mockVolRepo := &mockVolumeRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Volume, error) {
			return cachedVolumes, nil
		},
		isCacheValidFn: func(_ context.Context, _ string, _ time.Duration) (bool, error) {
			return false, nil // stale
		},
		upsertBatchFn: func(_ context.Context, _ string, volumes []entity.Volume) error {
			require.Len(t, volumes, 1)
			require.Equal(t, "refreshed-vol", volumes[0].Name)
			upsertCalled <- struct{}{}
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockVolRepo, &mockRemoteMachineRepo{})

	// First call: returns cached (stale), triggers background refresh
	volumes, err := uc.ListVolumes(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, volumes, 1)
	require.Equal(t, "stale-vol", volumes[0].Name)

	// Wait for the goroutine to complete
	select {
	case <-upsertCalled:
		// goroutine ran and upserted refreshed data
	case <-time.After(2 * time.Second):
		t.Fatal("refresh goroutine did not call UpsertBatch within 2s")
	}
}

func TestCreateVolume(t *testing.T) {
	createCalled := false
	cacheInvalidated := false

	mockCli := &mockDockerClient{
		volumeCreateFn: func(_ context.Context, name, driver string) (*entity.Volume, error) {
			createCalled = true
			require.Equal(t, "test-vol", name)
			require.Equal(t, "local", driver)
			return &entity.Volume{Name: "test-vol", Driver: "local", Mountpoint: "/var/lib/docker/volumes/test-vol/_data"}, nil
		},
	}

	mockVolRepo := &mockVolumeRepo{
		deleteByMachineFn: func(_ context.Context, machineID string) error {
			cacheInvalidated = true
			require.Equal(t, "local", machineID)
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockVolRepo, &mockRemoteMachineRepo{})

	volume, err := uc.CreateVolume(context.Background(), "local", "test-vol", "local")
	require.NoError(t, err)
	require.NotNil(t, volume)
	require.Equal(t, "test-vol", volume.Name)
	require.True(t, createCalled, "expected Docker VolumeCreate to be called")
	require.True(t, cacheInvalidated, "expected cache invalidation on create")
}

func TestDeleteVolume(t *testing.T) {
	deleteCalled := false
	cacheInvalidated := false

	mockCli := &mockDockerClient{
		volumeDeleteFn: func(_ context.Context, volumeID string) error {
			deleteCalled = true
			require.Equal(t, "user-vol", volumeID)
			return nil
		},
	}

	mockVolRepo := &mockVolumeRepo{
		deleteByMachineFn: func(_ context.Context, machineID string) error {
			cacheInvalidated = true
			require.Equal(t, "local", machineID)
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockVolRepo, &mockRemoteMachineRepo{})

	err := uc.DeleteVolume(context.Background(), "local", "user-vol")
	require.NoError(t, err)
	require.True(t, deleteCalled, "expected Docker VolumeDelete to be called")
	require.True(t, cacheInvalidated, "expected cache invalidation on delete")
}

func TestDeleteVolume_NotFound(t *testing.T) {
	mockCli := &mockDockerClient{
		volumeDeleteFn: func(_ context.Context, _ string) error {
			return errors.ErrVolumeNotFound
		},
	}

	uc := newUseCaseForTest(mockCli, &mockVolumeRepo{}, &mockRemoteMachineRepo{})

	err := uc.DeleteVolume(context.Background(), "local", "non-existent-vol")
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrVolumeNotFound)
}

func TestInspectVolume(t *testing.T) {
	inspectCalled := false

	mockCli := &mockDockerClient{
		volumeInspectFn: func(_ context.Context, volumeID string) (*entity.Volume, error) {
			inspectCalled = true
			require.Equal(t, "my-vol", volumeID)
			return &entity.Volume{Name: "my-vol", Driver: "local", Mountpoint: "/var/lib/docker/volumes/my-vol/_data"}, nil
		},
	}

	uc := newUseCaseForTest(mockCli, &mockVolumeRepo{}, &mockRemoteMachineRepo{})

	volume, err := uc.InspectVolume(context.Background(), "local", "my-vol")
	require.NoError(t, err)
	require.NotNil(t, volume)
	require.Equal(t, "my-vol", volume.Name)
	require.Equal(t, "local", volume.Driver)
	require.True(t, inspectCalled, "expected Docker VolumeInspect to be called")
}

func TestInspectVolume_NotFound(t *testing.T) {
	mockCli := &mockDockerClient{
		volumeInspectFn: func(_ context.Context, _ string) (*entity.Volume, error) {
			return nil, errors.ErrVolumeNotFound
		},
	}

	uc := newUseCaseForTest(mockCli, &mockVolumeRepo{}, &mockRemoteMachineRepo{})

	volume, err := uc.InspectVolume(context.Background(), "local", "non-existent-vol")
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrVolumeNotFound)
	require.Nil(t, volume)
}
