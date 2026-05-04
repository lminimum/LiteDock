package image

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerImage "github.com/docker/docker/api/types/image"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/pkg/dockerclient"
	"github.com/stretchr/testify/require"
)

// --- mockLogger ---

type mockLogger struct{}

func (m *mockLogger) Debug(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Info(_ string, _ ...interface{})       {}
func (m *mockLogger) Warn(_ string, _ ...interface{})       {}
func (m *mockLogger) Error(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Fatal(_ interface{}, _ ...interface{}) {}

// --- mockImageRepo ---

type mockImageRepo struct {
	listByMachineFn func(ctx context.Context, machineID string) ([]entity.Image, error)
	getByIDFn       func(ctx context.Context, machineID, imageID string) (*entity.Image, error)
	upsertBatchFn   func(ctx context.Context, machineID string, images []entity.Image) error
	deleteByMachineFn func(ctx context.Context, machineID string) error
	deleteByIDFn    func(ctx context.Context, machineID, imageID string) error
	isCacheValidFn  func(ctx context.Context, machineID string, maxAge time.Duration) (bool, error)
	countAllFn      func(ctx context.Context) (int64, error)
}

func (m *mockImageRepo) ListByMachine(ctx context.Context, machineID string) ([]entity.Image, error) {
	if m.listByMachineFn != nil {
		return m.listByMachineFn(ctx, machineID)
	}
	return nil, nil
}

func (m *mockImageRepo) GetByID(ctx context.Context, machineID, imageID string) (*entity.Image, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, machineID, imageID)
	}
	return nil, nil
}

func (m *mockImageRepo) UpsertBatch(ctx context.Context, machineID string, images []entity.Image) error {
	if m.upsertBatchFn != nil {
		return m.upsertBatchFn(ctx, machineID, images)
	}
	return nil
}

func (m *mockImageRepo) DeleteByMachine(ctx context.Context, machineID string) error {
	if m.deleteByMachineFn != nil {
		return m.deleteByMachineFn(ctx, machineID)
	}
	return nil
}

func (m *mockImageRepo) DeleteByID(ctx context.Context, machineID, imageID string) error {
	if m.deleteByIDFn != nil {
		return m.deleteByIDFn(ctx, machineID, imageID)
	}
	return nil
}

func (m *mockImageRepo) IsCacheValid(ctx context.Context, machineID string, maxAge time.Duration) (bool, error) {
	if m.isCacheValidFn != nil {
		return m.isCacheValidFn(ctx, machineID, maxAge)
	}
	return true, nil
}

func (m *mockImageRepo) CountAll(ctx context.Context) (int64, error) {
	if m.countAllFn != nil {
		return m.countAllFn(ctx)
	}
	return 0, nil
}

var _ repo.ImageRepo = (*mockImageRepo)(nil)

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
func (m *mockRemoteMachineRepo) List(_ context.Context) ([]entity.RemoteMachine, error)  { return nil, nil }
func (m *mockRemoteMachineRepo) Count(_ context.Context) (int64, error)                  { return 0, nil }
func (m *mockRemoteMachineRepo) Update(_ context.Context, _ *entity.RemoteMachine) error { return nil }
func (m *mockRemoteMachineRepo) Delete(_ context.Context, _ string) error                { return nil }
func (m *mockRemoteMachineRepo) GetByHost(_ context.Context, _ string) (*entity.RemoteMachine, error) {
	return nil, nil
}

var _ repo.RemoteMachineRepo = (*mockRemoteMachineRepo)(nil)

// --- mockDockerClient ---

type mockDockerClient struct {
	imageListFn    func(ctx context.Context, opts dockerImage.ListOptions) ([]entity.Image, error)
	imagePullFn    func(ctx context.Context, ref string, opts dockerImage.PullOptions) error
	imageRemoveFn  func(ctx context.Context, id string, opts dockerImage.RemoveOptions) ([]dockerImage.DeleteResponse, error)
	imageInspectFn func(ctx context.Context, id string) (dockerImage.InspectResponse, error)
	imagePruneFn   func(ctx context.Context, opts filters.Args) (dockerImage.PruneReport, error)
	closeFn        func() error
}

func (m *mockDockerClient) Ping(_ context.Context) error { return nil }
func (m *mockDockerClient) ContainerList(_ context.Context) ([]entity.Container, error) {
	return nil, nil
}
func (m *mockDockerClient) ContainerLogs(_ context.Context, _, _ string) (string, error) { return "", nil }
func (m *mockDockerClient) ContainerExec(_ context.Context, _ string, _ []string) (string, error) {
	return "", nil
}
func (m *mockDockerClient) ContainerStart(_ context.Context, _ string) error               { return nil }
func (m *mockDockerClient) ContainerStop(_ context.Context, _ string, _ time.Duration) error  { return nil }
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
func (m *mockDockerClient) VolumeList(_ context.Context) ([]entity.Volume, error) { return nil, nil }
func (m *mockDockerClient) VolumeCreate(_ context.Context, _, _ string) (*entity.Volume, error) {
	return nil, nil
}
func (m *mockDockerClient) VolumeDelete(_ context.Context, _ string) error { return nil }
func (m *mockDockerClient) VolumeInspect(_ context.Context, _ string) (*entity.Volume, error) {
	return nil, nil
}
func (m *mockDockerClient) ImageList(ctx context.Context, opts dockerImage.ListOptions) ([]entity.Image, error) {
	if m.imageListFn != nil {
		return m.imageListFn(ctx, opts)
	}
	return nil, nil
}
func (m *mockDockerClient) ImagePull(ctx context.Context, ref string, opts dockerImage.PullOptions) error {
	if m.imagePullFn != nil {
		return m.imagePullFn(ctx, ref, opts)
	}
	return nil
}
func (m *mockDockerClient) ImageRemove(ctx context.Context, id string, opts dockerImage.RemoveOptions) ([]dockerImage.DeleteResponse, error) {
	if m.imageRemoveFn != nil {
		return m.imageRemoveFn(ctx, id, opts)
	}
	return nil, nil
}
func (m *mockDockerClient) ImageInspect(ctx context.Context, id string) (dockerImage.InspectResponse, error) {
	if m.imageInspectFn != nil {
		return m.imageInspectFn(ctx, id)
	}
	return dockerImage.InspectResponse{}, nil
}
func (m *mockDockerClient) ImagePrune(ctx context.Context, opts filters.Args) (dockerImage.PruneReport, error) {
	if m.imagePruneFn != nil {
		return m.imagePruneFn(ctx, opts)
	}
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

func newUseCaseForTest(client dockerclient.Client, imgRepo repo.ImageRepo, rmRepo repo.RemoteMachineRepo) *ImageUseCase {
	return &ImageUseCase{
		imageRepo:         imgRepo,
		remoteMachineRepo: rmRepo,
		cacheMaxAge:       30 * time.Second,
		l:                 &mockLogger{},
		testDockerClient:  client,
	}
}

// makeTestImage is a helper to create an entity.Image with common defaults.
func makeTestImage(id, machineID string, repoTags []string) entity.Image {
	return entity.Image{
		ID:        id,
		MachineID: machineID,
		RepoTags:  repoTags,
		Size:      100 * 1024 * 1024, // 100 MB
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CachedAt:  time.Now(),
	}
}

// makeInspectResponse creates a dockerImage.InspectResponse for testing.
func makeInspectResponse(id, created string, repoTags []string) dockerImage.InspectResponse {
	return dockerImage.InspectResponse{
		ID:          id,
		RepoTags:    repoTags,
		RepoDigests: []string{},
		Size:        100 * 1024 * 1024,
		Created:     created,
		Config: &container.Config{
			Labels: map[string]string{"maintainer": "test"},
		},
	}
}

// --- tests ---

func TestList_CacheHit(t *testing.T) {
	tests := []struct {
		name        string
		cachedImages []entity.Image
	}{
		{
			name: "returns cached images when cache is valid",
			cachedImages: []entity.Image{
				makeTestImage("abc123", "local", []string{"nginx:latest"}),
				makeTestImage("def456", "local", []string{"alpine:latest"}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockImgRepo := &mockImageRepo{
				listByMachineFn: func(_ context.Context, _ string) ([]entity.Image, error) {
					return tt.cachedImages, nil
				},
				isCacheValidFn: func(_ context.Context, _ string, _ time.Duration) (bool, error) {
					return true, nil
				},
			}

			uc := newUseCaseForTest(&mockDockerClient{}, mockImgRepo, &mockRemoteMachineRepo{})

			images, err := uc.List(context.Background(), "local")
			require.NoError(t, err)
			require.Len(t, images, 2)
			require.Equal(t, "nginx:latest", images[0].RepoTags[0])
			require.Equal(t, "alpine:latest", images[1].RepoTags[0])
		})
	}
}

func TestList_CacheMiss_FetchesFromDocker(t *testing.T) {
	dockerCalled := false
	upsertCalled := false

	mockCli := &mockDockerClient{
		imageListFn: func(_ context.Context, _ dockerImage.ListOptions) ([]entity.Image, error) {
			dockerCalled = true
			return []entity.Image{
				makeTestImage("abc123", "local", []string{"nginx:latest"}),
			}, nil
		},
	}

	mockImgRepo := &mockImageRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Image, error) {
			return nil, nil // empty cache
		},
		upsertBatchFn: func(_ context.Context, _ string, images []entity.Image) error {
			upsertCalled = true
			require.Len(t, images, 1)
			require.Equal(t, "nginx:latest", images[0].RepoTags[0])
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockImgRepo, &mockRemoteMachineRepo{})

	images, err := uc.List(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Equal(t, "nginx:latest", images[0].RepoTags[0])
	require.True(t, dockerCalled, "expected Docker to be called for empty cache")
	require.True(t, upsertCalled, "expected UpsertBatch to be called after fetch")
}

func TestList_DockerError(t *testing.T) {
	mockCli := &mockDockerClient{
		imageListFn: func(_ context.Context, _ dockerImage.ListOptions) ([]entity.Image, error) {
			return nil, fmt.Errorf("docker: connection refused")
		},
	}

	mockImgRepo := &mockImageRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Image, error) {
			return nil, nil // empty cache → triggers Docker call
		},
	}

	uc := newUseCaseForTest(mockCli, mockImgRepo, &mockRemoteMachineRepo{})

	images, err := uc.List(context.Background(), "local")
	require.Error(t, err)
	require.Nil(t, images)
	require.Contains(t, err.Error(), "connection refused")
}

func TestList_CacheValid_ListByMachineError(t *testing.T) {
	mockImgRepo := &mockImageRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Image, error) {
			return nil, fmt.Errorf("database: connection failed")
		},
	}

	uc := newUseCaseForTest(&mockDockerClient{}, mockImgRepo, &mockRemoteMachineRepo{})

	images, err := uc.List(context.Background(), "local")
	require.Error(t, err)
	require.Nil(t, images)
	require.Contains(t, err.Error(), "connection failed")
}

func TestPull_Success(t *testing.T) {
	pullCalled := false
	inspectCalled := false
	cacheInvalidated := false

	mockCli := &mockDockerClient{
		imagePullFn: func(_ context.Context, ref string, _ dockerImage.PullOptions) error {
			pullCalled = true
			require.Equal(t, "nginx:latest", ref)
			return nil
		},
		imageInspectFn: func(_ context.Context, id string) (dockerImage.InspectResponse, error) {
			inspectCalled = true
			require.Equal(t, "nginx:latest", id)
			return makeInspectResponse("sha256:abc123def456abc123def456abc123def456abc1", "2024-01-01T00:00:00.000000000Z", []string{"nginx:latest"}), nil
		},
	}

	mockImgRepo := &mockImageRepo{
		deleteByMachineFn: func(_ context.Context, machineID string) error {
			cacheInvalidated = true
			require.Equal(t, "local", machineID)
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockImgRepo, &mockRemoteMachineRepo{})

	img, err := uc.Pull(context.Background(), "local", "nginx", "latest")
	require.NoError(t, err)
	require.NotNil(t, img)
	require.Equal(t, "abc123def456", img.ID)
	require.Equal(t, "nginx:latest", img.RepoTags[0])
	require.True(t, pullCalled, "expected Docker ImagePull to be called")
	require.True(t, inspectCalled, "expected Docker ImageInspect to be called after pull")
	require.True(t, cacheInvalidated, "expected cache invalidation on pull")
}

func TestPull_InvalidRef(t *testing.T) {
	mockCli := &mockDockerClient{
		imagePullFn: func(_ context.Context, _ string, _ dockerImage.PullOptions) error {
			t.Fatal("Docker should not be called for empty repository")
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, &mockImageRepo{}, &mockRemoteMachineRepo{})

	img, err := uc.Pull(context.Background(), "local", "", "")
	require.Error(t, err)
	require.Nil(t, img)
	require.Contains(t, err.Error(), "repository is required")
}

func TestDelete_Success(t *testing.T) {
	removeCalled := false
	cacheInvalidated := false

	mockCli := &mockDockerClient{
		imageRemoveFn: func(_ context.Context, id string, _ dockerImage.RemoveOptions) ([]dockerImage.DeleteResponse, error) {
			removeCalled = true
			require.Equal(t, "abc123", id)
			return []dockerImage.DeleteResponse{
				{Deleted: "sha256:abc123def456"},
			}, nil
		},
	}

	mockImgRepo := &mockImageRepo{
		deleteByMachineFn: func(_ context.Context, machineID string) error {
			cacheInvalidated = true
			require.Equal(t, "local", machineID)
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockImgRepo, &mockRemoteMachineRepo{})

	resp, err := uc.Delete(context.Background(), "local", "abc123")
	require.NoError(t, err)
	require.Len(t, resp, 1)
	require.Equal(t, "sha256:abc123def456", resp[0].Deleted)
	require.True(t, removeCalled, "expected Docker ImageRemove to be called")
	require.True(t, cacheInvalidated, "expected cache invalidation on delete")
}

func TestDelete_ImageInUse(t *testing.T) {
	removeCalled := false
	cacheInvalidated := false

	deleteErr := errors.New("docker: conflict: unable to delete abc123 (must be forced) - image is being used by running container")

	mockCli := &mockDockerClient{
		imageRemoveFn: func(_ context.Context, id string, _ dockerImage.RemoveOptions) ([]dockerImage.DeleteResponse, error) {
			removeCalled = true
			require.Equal(t, "abc123", id)
			return nil, deleteErr
		},
	}

	mockImgRepo := &mockImageRepo{
		deleteByMachineFn: func(_ context.Context, machineID string) error {
			cacheInvalidated = true
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockImgRepo, &mockRemoteMachineRepo{})

	resp, err := uc.Delete(context.Background(), "local", "abc123")
	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorIs(t, err, deleteErr)
	require.True(t, removeCalled, "expected Docker ImageRemove to be called")
	require.False(t, cacheInvalidated, "expected cache NOT to be invalidated on error")
}

func TestPrune_Success(t *testing.T) {
	pruneCalled := false
	cacheInvalidated := false

	mockCli := &mockDockerClient{
		imagePruneFn: func(_ context.Context, _ filters.Args) (dockerImage.PruneReport, error) {
			pruneCalled = true
			return dockerImage.PruneReport{
				ImagesDeleted:  []dockerImage.DeleteResponse{{Deleted: "sha256:abc"}, {Deleted: "sha256:def"}},
				SpaceReclaimed: 200 * 1024 * 1024,
			}, nil
		},
	}

	mockImgRepo := &mockImageRepo{
		deleteByMachineFn: func(_ context.Context, machineID string) error {
			cacheInvalidated = true
			require.Equal(t, "local", machineID)
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockImgRepo, &mockRemoteMachineRepo{})

	report, err := uc.Prune(context.Background(), "local")
	require.NoError(t, err)
	require.NotNil(t, report)
	require.Len(t, report.ImagesDeleted, 2)
	require.Equal(t, uint64(200*1024*1024), report.SpaceReclaimed)
	require.True(t, pruneCalled, "expected Docker ImagePrune to be called")
	require.True(t, cacheInvalidated, "expected cache invalidation on prune")
}

func TestInspect_Success(t *testing.T) {
	inspectCalled := false

	mockCli := &mockDockerClient{
		imageInspectFn: func(_ context.Context, id string) (dockerImage.InspectResponse, error) {
			inspectCalled = true
			require.Equal(t, "abc123", id)
			return makeInspectResponse("sha256:abc123def456abc123def456abc123def456abc1", "2024-01-01T00:00:00.000000000Z", []string{"nginx:latest"}), nil
		},
	}

	uc := newUseCaseForTest(mockCli, &mockImageRepo{}, &mockRemoteMachineRepo{})

	img, err := uc.Inspect(context.Background(), "local", "abc123")
	require.NoError(t, err)
	require.NotNil(t, img)
	require.Equal(t, "abc123def456", img.ID)            // sha256: stripped, first 12 chars
	require.Equal(t, "nginx:latest", img.RepoTags[0])
	require.Equal(t, "test", img.Labels["maintainer"])
	require.False(t, img.CreatedAt.IsZero())
	require.True(t, inspectCalled, "expected Docker ImageInspect to be called")
}

func TestInspect_NotFound(t *testing.T) {
	notFoundErr := errors.New("docker: image not found")

	mockCli := &mockDockerClient{
		imageInspectFn: func(_ context.Context, _ string) (dockerImage.InspectResponse, error) {
			return dockerImage.InspectResponse{}, notFoundErr
		},
	}

	uc := newUseCaseForTest(mockCli, &mockImageRepo{}, &mockRemoteMachineRepo{})

	img, err := uc.Inspect(context.Background(), "local", "nonexistent")
	require.Error(t, err)
	require.Nil(t, img)
	require.ErrorIs(t, err, notFoundErr)
}

func TestRefresh(t *testing.T) {
	upsertCalled := make(chan struct{}, 1)

	mockCli := &mockDockerClient{
		imageListFn: func(_ context.Context, _ dockerImage.ListOptions) ([]entity.Image, error) {
			return []entity.Image{
				makeTestImage("refreshed1", "local", []string{"alpine:latest"}),
			}, nil
		},
	}

	cachedImages := []entity.Image{
		makeTestImage("stale1", "local", []string{"nginx:old"}),
	}

	mockImgRepo := &mockImageRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Image, error) {
			return cachedImages, nil
		},
		isCacheValidFn: func(_ context.Context, _ string, _ time.Duration) (bool, error) {
			return false, nil // stale cache → triggers background refresh
		},
		upsertBatchFn: func(_ context.Context, _ string, images []entity.Image) error {
			require.Len(t, images, 1)
			require.Equal(t, "alpine:latest", images[0].RepoTags[0])
			upsertCalled <- struct{}{}
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockImgRepo, &mockRemoteMachineRepo{})

	// First call: returns cached (stale) data, triggers background refresh
	images, err := uc.List(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Equal(t, "nginx:old", images[0].RepoTags[0])

	// Wait for the background refresh goroutine to complete
	select {
	case <-upsertCalled:
		// goroutine ran and upserted refreshed data
	case <-time.After(2 * time.Second):
		t.Fatal("refresh goroutine did not call UpsertBatch within 2s")
	}
}

func TestList_StaleCache_IsCacheValidError(t *testing.T) {
	dockerCalled := false

	mockCli := &mockDockerClient{
		imageListFn: func(_ context.Context, _ dockerImage.ListOptions) ([]entity.Image, error) {
			dockerCalled = true
			return []entity.Image{
				makeTestImage("abc123", "local", []string{"nginx:latest"}),
			}, nil
		},
	}

	cachedImages := []entity.Image{
		makeTestImage("stale1", "local", []string{"nginx:old"}),
	}

	mockImgRepo := &mockImageRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Image, error) {
			return cachedImages, nil
		},
		isCacheValidFn: func(_ context.Context, _ string, _ time.Duration) (bool, error) {
			return false, fmt.Errorf("cache check failed")
		},
		upsertBatchFn: func(_ context.Context, _ string, images []entity.Image) error {
			require.Len(t, images, 1)
			require.Equal(t, "nginx:latest", images[0].RepoTags[0])
			return nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockImgRepo, &mockRemoteMachineRepo{})

	// When IsCacheValid returns error, we still return cached images and trigger refresh
	images, err := uc.List(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Equal(t, "nginx:old", images[0].RepoTags[0])

	// Wait for background refresh
	time.Sleep(100 * time.Millisecond)
	require.True(t, dockerCalled, "expected Docker to be called in background refresh")
}

func TestPrune_DockerError(t *testing.T) {
	pruneErr := errors.New("docker: prune failed")
	pruneCalled := false

	mockCli := &mockDockerClient{
		imagePruneFn: func(_ context.Context, _ filters.Args) (dockerImage.PruneReport, error) {
			pruneCalled = true
			return dockerImage.PruneReport{}, pruneErr
		},
	}

	uc := newUseCaseForTest(mockCli, &mockImageRepo{}, &mockRemoteMachineRepo{})

	report, err := uc.Prune(context.Background(), "local")
	require.Error(t, err)
	require.Nil(t, report)
	require.ErrorIs(t, err, pruneErr)
	require.True(t, pruneCalled, "expected Docker ImagePrune to be called")
}

func TestPull_ImageInspectError(t *testing.T) {
	inspectErr := errors.New("docker: image inspect failed")
	pullCalled := false
	inspectCalled := false

	mockCli := &mockDockerClient{
		imagePullFn: func(_ context.Context, ref string, _ dockerImage.PullOptions) error {
			pullCalled = true
			require.Equal(t, "nginx:latest", ref)
			return nil
		},
		imageInspectFn: func(_ context.Context, _ string) (dockerImage.InspectResponse, error) {
			inspectCalled = true
			return dockerImage.InspectResponse{}, inspectErr
		},
	}

	uc := newUseCaseForTest(mockCli, &mockImageRepo{}, &mockRemoteMachineRepo{})

	img, err := uc.Pull(context.Background(), "local", "nginx", "latest")
	require.Error(t, err)
	require.Nil(t, img)
	require.ErrorIs(t, err, inspectErr)
	require.True(t, pullCalled, "expected Docker ImagePull to be called")
	require.True(t, inspectCalled, "expected Docker ImageInspect to be called after pull")
}

func TestList_EmptyCache_UpsertError(t *testing.T) {
	upsertErr := errors.New("database: upsert failed")

	mockCli := &mockDockerClient{
		imageListFn: func(_ context.Context, _ dockerImage.ListOptions) ([]entity.Image, error) {
			return []entity.Image{
				makeTestImage("abc123", "local", []string{"nginx:latest"}),
			}, nil
		},
	}

	mockImgRepo := &mockImageRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Image, error) {
			return nil, nil
		},
		upsertBatchFn: func(_ context.Context, _ string, _ []entity.Image) error {
			return upsertErr
		},
	}

	uc := newUseCaseForTest(mockCli, mockImgRepo, &mockRemoteMachineRepo{})

	images, err := uc.List(context.Background(), "local")
	require.Error(t, err)
	require.Len(t, images, 1) // images returned even when upsert fails
	require.ErrorIs(t, err, upsertErr)
}

func TestNewImageUseCase(t *testing.T) {
	mockImgRepo := &mockImageRepo{}
	mockRMRepo := &mockRemoteMachineRepo{}
	mockLogger := &mockLogger{}
	cacheTTL := 30 * time.Second

	uc := NewImageUseCase(mockImgRepo, mockRMRepo, cacheTTL, mockLogger)

	require.NotNil(t, uc)
	require.Equal(t, mockImgRepo, uc.imageRepo)
	require.Equal(t, mockRMRepo, uc.remoteMachineRepo)
	require.Equal(t, cacheTTL, uc.cacheMaxAge)
	require.Equal(t, mockLogger, uc.l)
	require.Nil(t, uc.testDockerClient)
}

func TestInspectToEntity(t *testing.T) {
	tests := []struct {
		name     string
		resp     dockerImage.InspectResponse
		machineID string
		wantID   string
		wantTags []string
	}{
		{
			name: "normal response with sha256 prefix",
			resp: makeInspectResponse(
				"sha256:abcdef1234567890abcdef1234567890abcdef12",
				"2024-01-01T00:00:00.000000000Z",
				[]string{"alpine:latest", "alpine:3.19"},
			),
			machineID: "local",
			wantID:    "abcdef123456",
			wantTags:  []string{"alpine:latest", "alpine:3.19"},
		},
		{
			name: "bad created time falls back to zero",
			resp: dockerImage.InspectResponse{
				ID:      "sha256:abcdef123456",
				Created: "not-a-valid-timestamp",
				Size:    50 * 1024 * 1024,
				RepoTags: []string{"busybox:latest"},
			},
			machineID: "remote-1",
			wantID:    "abcdef123456",
			wantTags:  []string{"busybox:latest"},
		},
		{
			name: "nil repoTags handled",
			resp: dockerImage.InspectResponse{
				ID:        "sha256:abcdef123456",
				RepoTags:  nil,
				Size:      1024,
				Created:   "2024-01-01T00:00:00.000000000Z",
				Config:    &container.Config{Labels: nil},
			},
			machineID: "local",
			wantID:    "abcdef123456",
			wantTags:  []string{},
		},
		{
			name: "short id without sha256 prefix",
			resp: dockerImage.InspectResponse{
				ID:      "abc123",
				Size:    2048,
				Created: "2024-01-01T00:00:00.000000000Z",
			},
			machineID: "local",
			wantID:    "abc123",
			wantTags:  []string{},
		},
		{
			name: "nil config defaults labels to empty map",
			resp: dockerImage.InspectResponse{
				ID:      "sha256:abcdef123456",
				Size:    4096,
				Created: "2024-01-01T00:00:00.000000000Z",
				RepoTags: []string{"test:latest"},
				Config:    nil,
			},
			machineID: "local",
			wantID:    "abcdef123456",
			wantTags:  []string{"test:latest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := inspectToEntity(&tt.resp, tt.machineID)

			require.Equal(t, tt.wantID, img.ID)
			require.Equal(t, tt.machineID, img.MachineID)
			require.Equal(t, tt.wantTags, img.RepoTags)
			require.Equal(t, tt.resp.Size, img.Size)
			require.NotNil(t, img.Labels)

			if tt.resp.Config != nil && tt.resp.Config.Labels != nil {
				require.Equal(t, tt.resp.Config.Labels, img.Labels)
			} else {
				require.Empty(t, img.Labels)
			}
		})
	}
}

func TestPull_WithTag(t *testing.T) {
	pullCalled := false

	mockCli := &mockDockerClient{
		imagePullFn: func(_ context.Context, ref string, _ dockerImage.PullOptions) error {
			pullCalled = true
			require.Equal(t, "nginx:1.25", ref)
			return nil
		},
		imageInspectFn: func(_ context.Context, _ string) (dockerImage.InspectResponse, error) {
			return makeInspectResponse("sha256:abc123def456abc123def456abc123def456abc1", "2024-01-01T00:00:00.000000000Z", []string{"nginx:1.25"}), nil
		},
	}

	uc := newUseCaseForTest(mockCli, &mockImageRepo{}, &mockRemoteMachineRepo{})

	img, err := uc.Pull(context.Background(), "local", "nginx", "1.25")
	require.NoError(t, err)
	require.NotNil(t, img)
	require.True(t, pullCalled, "expected Docker ImagePull to be called with tag")
}

func TestList_EmptyCache_ZeroImages(t *testing.T) {
	dockerCalled := false

	mockCli := &mockDockerClient{
		imageListFn: func(_ context.Context, _ dockerImage.ListOptions) ([]entity.Image, error) {
			dockerCalled = true
			return []entity.Image{}, nil
		},
	}

	mockImgRepo := &mockImageRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.Image, error) {
			return nil, nil
		},
	}

	uc := newUseCaseForTest(mockCli, mockImgRepo, &mockRemoteMachineRepo{})

	images, err := uc.List(context.Background(), "local")
	require.NoError(t, err)
	require.Empty(t, images)
	require.True(t, dockerCalled)
}
