package compose

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// --- mockComposeRepo ---

type mockComposeRepo struct {
	listByMachineFn    func(ctx context.Context, machineID string) ([]entity.ComposeFile, error)
	getByProjectNameFn func(ctx context.Context, machineID, projectName string) (*entity.ComposeFile, error)
	createFn           func(ctx context.Context, cf *entity.ComposeFile) error
	updateFn           func(ctx context.Context, cf *entity.ComposeFile) error
	deleteByIDFn       func(ctx context.Context, machineID, id string) error
	isCacheValidFn     func(ctx context.Context, machineID string, maxAge time.Duration) (bool, error)
	upsertBatchFn      func(ctx context.Context, cfs []entity.ComposeFile) error
}

func (m *mockComposeRepo) ListByMachine(ctx context.Context, machineID string) ([]entity.ComposeFile, error) {
	if m.listByMachineFn != nil {
		return m.listByMachineFn(ctx, machineID)
	}
	return nil, nil
}

func (m *mockComposeRepo) GetByID(_ context.Context, _, _ string) (*entity.ComposeFile, error) {
	return nil, errors.ErrNotFound
}

func (m *mockComposeRepo) GetByProjectName(ctx context.Context, machineID, projectName string) (*entity.ComposeFile, error) {
	if m.getByProjectNameFn != nil {
		return m.getByProjectNameFn(ctx, machineID, projectName)
	}
	return nil, errors.ErrNotFound
}

func (m *mockComposeRepo) Create(ctx context.Context, cf *entity.ComposeFile) error {
	if m.createFn != nil {
		return m.createFn(ctx, cf)
	}
	return nil
}

func (m *mockComposeRepo) Update(ctx context.Context, cf *entity.ComposeFile) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, cf)
	}
	return nil
}

func (m *mockComposeRepo) DeleteByMachine(_ context.Context, _ string) error {
	return nil
}

func (m *mockComposeRepo) DeleteByID(ctx context.Context, machineID, id string) error {
	if m.deleteByIDFn != nil {
		return m.deleteByIDFn(ctx, machineID, id)
	}
	return nil
}

func (m *mockComposeRepo) IsCacheValid(ctx context.Context, machineID string, maxAge time.Duration) (bool, error) {
	if m.isCacheValidFn != nil {
		return m.isCacheValidFn(ctx, machineID, maxAge)
	}
	return true, nil
}

func (m *mockComposeRepo) UpsertBatch(ctx context.Context, cfs []entity.ComposeFile) error {
	if m.upsertBatchFn != nil {
		return m.upsertBatchFn(ctx, cfs)
	}
	return nil
}

var _ repo.ComposeRepo = (*mockComposeRepo)(nil)

// --- mockComposeClient ---

type mockComposeClient struct {
	composeUpFn      func(ctx context.Context, machineID, projectName, composeFilePath string) error
	composeDownFn    func(ctx context.Context, machineID, projectName string, volumes bool) error
	composeBuildFn   func(ctx context.Context, machineID, composeFilePath string) error
	composeStartFn   func(ctx context.Context, machineID, projectName string) error
	composeStopFn    func(ctx context.Context, machineID, projectName string) error
	composeRestartFn func(ctx context.Context, machineID, projectName string) error
	composePsFn      func(ctx context.Context, machineID, projectName string) ([]dockerclient.ComposeServiceStatus, error)
	composeLogsFn    func(ctx context.Context, machineID, projectName string) (io.ReadCloser, error)
	composeConfigFn  func(ctx context.Context, machineID, composeFilePath string) (string, error)
	composeLsFn      func(ctx context.Context, machineID string) ([]dockerclient.ComposeProjectEntry, error)
}

func (m *mockComposeClient) ComposeUp(ctx context.Context, machineID, projectName, composeFilePath string) error {
	if m.composeUpFn != nil {
		return m.composeUpFn(ctx, machineID, projectName, composeFilePath)
	}
	return nil
}

func (m *mockComposeClient) ComposeDown(ctx context.Context, machineID, projectName string, volumes bool) error {
	if m.composeDownFn != nil {
		return m.composeDownFn(ctx, machineID, projectName, volumes)
	}
	return nil
}

func (m *mockComposeClient) ComposeBuild(ctx context.Context, machineID, composeFilePath string) error {
	if m.composeBuildFn != nil {
		return m.composeBuildFn(ctx, machineID, composeFilePath)
	}
	return nil
}

func (m *mockComposeClient) ComposeStart(ctx context.Context, machineID, projectName string) error {
	if m.composeStartFn != nil {
		return m.composeStartFn(ctx, machineID, projectName)
	}
	return nil
}

func (m *mockComposeClient) ComposeStop(ctx context.Context, machineID, projectName string) error {
	if m.composeStopFn != nil {
		return m.composeStopFn(ctx, machineID, projectName)
	}
	return nil
}

func (m *mockComposeClient) ComposeRestart(ctx context.Context, machineID, projectName string) error {
	if m.composeRestartFn != nil {
		return m.composeRestartFn(ctx, machineID, projectName)
	}
	return nil
}

func (m *mockComposeClient) ComposePs(ctx context.Context, machineID, projectName string) ([]dockerclient.ComposeServiceStatus, error) {
	if m.composePsFn != nil {
		return m.composePsFn(ctx, machineID, projectName)
	}
	return nil, nil
}

func (m *mockComposeClient) ComposeLogs(ctx context.Context, machineID, projectName string) (io.ReadCloser, error) {
	if m.composeLogsFn != nil {
		return m.composeLogsFn(ctx, machineID, projectName)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *mockComposeClient) ComposeConfig(ctx context.Context, machineID, composeFilePath string) (string, error) {
	if m.composeConfigFn != nil {
		return m.composeConfigFn(ctx, machineID, composeFilePath)
	}
	return "", nil
}

func (m *mockComposeClient) ComposeLs(ctx context.Context, machineID string) ([]dockerclient.ComposeProjectEntry, error) {
	if m.composeLsFn != nil {
		return m.composeLsFn(ctx, machineID)
	}
	return nil, nil
}

var _ dockerclient.DockerComposeClient = (*mockComposeClient)(nil)

// --- mockRemoteMachineRepo ---

type mockRemoteMachineRepo struct {
	getByIDFn func(ctx context.Context, id string) (*entity.RemoteMachine, error)
}

func (m *mockRemoteMachineRepo) Create(_ context.Context, _ *entity.RemoteMachine) error { return nil }

func (m *mockRemoteMachineRepo) GetByID(ctx context.Context, id string) (*entity.RemoteMachine, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

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

// --- helpers ---

func newUseCaseForTest(
	compRepo repo.ComposeRepo,
	rmRepo repo.RemoteMachineRepo,
	composeCli dockerclient.DockerComposeClient,
	composeDir string,
) *ComposeUseCase {
	return &ComposeUseCase{
		composeRepo:             compRepo,
		remoteMachineRepo:       rmRepo,
		cacheMaxAge:             10 * time.Second,
		composeDir:              composeDir,
		l:                       &mockLogger{},
		testDockerComposeClient: composeCli,
	}
}

// makeTestComposeFile creates a ComposeFile entity for use in tests.
func makeTestComposeFile(id, machineID, projectName, filePath string) *entity.ComposeFile {
	now := time.Now()
	return &entity.ComposeFile{
		ID:          id,
		MachineID:   machineID,
		Name:        projectName,
		FilePath:    filePath,
		ProjectName: projectName,
		Status:      entity.OpStopped,
		Services:    []entity.ComposeService{},
		CreatedAt:   now,
		UpdatedAt:   now,
		CachedAt:    now,
	}
}

// makeTestService creates a ComposeService for use in tests.
func makeTestService(name, status string) entity.ComposeService {
	return entity.ComposeService{
		Name:   name,
		Status: status,
	}
}

// ============================================================================
// Tests
// ============================================================================

func TestNewComposeUseCase(t *testing.T) {
	uc := NewComposeUseCase(
		&mockComposeRepo{},
		&mockRemoteMachineRepo{},
		30*time.Second,
		t.TempDir(),
		&mockLogger{},
	)
	require.NotNil(t, uc)
	require.Equal(t, 30*time.Second, uc.cacheMaxAge)
	require.Nil(t, uc.testDockerComposeClient)
}

func TestHasRunningService(t *testing.T) {
	tests := []struct {
		name     string
		services []entity.ComposeService
		want     bool
	}{
		{
			name:     "empty services",
			services: nil,
			want:     false,
		},
		{
			name: "no running service",
			services: []entity.ComposeService{
				makeTestService("web", "exited"),
				makeTestService("db", "exited"),
			},
			want: false,
		},
		{
			name: "has running service",
			services: []entity.ComposeService{
				makeTestService("web", "exited"),
				makeTestService("db", "running"),
			},
			want: true,
		},
		{
			name: "all running",
			services: []entity.ComposeService{
				makeTestService("web", "running"),
				makeTestService("db", "running"),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasRunningService(tt.services)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestValidateComposeName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid names
		{name: "simple lowercase", input: "myapp", wantErr: false},
		{name: "simple uppercase", input: "MyApp", wantErr: false},
		{name: "with numbers", input: "app123", wantErr: false},
		{name: "with underscore", input: "my_app", wantErr: false},
		{name: "with dot", input: "my.app", wantErr: false},
		{name: "with hyphen", input: "my-app", wantErr: false},
		{name: "alphanumeric mix", input: "My_App.123", wantErr: false},

		// Invalid names
		{name: "empty string", input: "", wantErr: true},
		{name: "whitespace only", input: "   ", wantErr: true},
		{name: "starts with special char", input: "-myapp", wantErr: true},
		{name: "starts with underscore", input: "_myapp", wantErr: true},
		{name: "contains slash", input: "my/app", wantErr: true},
		{name: "contains space", input: "my app", wantErr: true},
		{name: "contains at sign", input: "my@app", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateComposeName(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, errors.ErrInvalidInput)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// --- List ---

func TestList_CacheHit(t *testing.T) {
	cachedFiles := []entity.ComposeFile{
		{ID: "f1", MachineID: "local", ProjectName: "app1", Status: entity.OpRunning},
		{ID: "f2", MachineID: "local", ProjectName: "app2", Status: entity.OpStopped},
	}

	mockCompRepo := &mockComposeRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.ComposeFile, error) {
			return cachedFiles, nil
		},
		isCacheValidFn: func(_ context.Context, _ string, _ time.Duration) (bool, error) {
			return true, nil
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, nil, t.TempDir())

	files, err := uc.List(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, files, 2)
	require.Equal(t, "app1", files[0].ProjectName)
	require.Equal(t, "app2", files[1].ProjectName)
}

func TestList_StaleCache_TriggersRefresh(t *testing.T) {
	upsertCalled := make(chan struct{}, 1)
	cachedFiles := []entity.ComposeFile{
		{ID: "f1", MachineID: "local", ProjectName: "stale-app", CachedAt: time.Now().Add(-1 * time.Hour)},
	}

	mockCli := &mockComposeClient{
		composePsFn: func(_ context.Context, _, _ string) ([]dockerclient.ComposeServiceStatus, error) {
			return []dockerclient.ComposeServiceStatus{
				{Name: "web", ServiceName: "web", Status: "running", Health: "healthy"},
			}, nil
		},
	}

	mockCompRepo := &mockComposeRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.ComposeFile, error) {
			return cachedFiles, nil
		},
		isCacheValidFn: func(_ context.Context, _ string, _ time.Duration) (bool, error) {
			return false, nil
		},
		upsertBatchFn: func(_ context.Context, cfs []entity.ComposeFile) error {
			require.Len(t, cfs, 1)
			require.Equal(t, "running", cfs[0].Status)
			upsertCalled <- struct{}{}
			return nil
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, mockCli, t.TempDir())

	files, err := uc.List(context.Background(), "local")
	require.NoError(t, err)
	require.Len(t, files, 1)
	require.Equal(t, "stale-app", files[0].ProjectName)

	select {
	case <-upsertCalled:
	case <-time.After(2 * time.Second):
		t.Fatal("refresh goroutine did not call UpsertBatch within 2s")
	}
}

func TestList_EmptyCache(t *testing.T) {
	mockCompRepo := &mockComposeRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.ComposeFile, error) {
			return nil, nil
		},
	}
	mockRMRepo := &mockRemoteMachineRepo{
		getByIDFn: func(_ context.Context, id string) (*entity.RemoteMachine, error) {
			return nil, errors.ErrNotFound
		},
	}

	uc := newUseCaseForTest(mockCompRepo, mockRMRepo, &mockComposeClient{}, t.TempDir())

	files, err := uc.List(context.Background(), "local")
	require.NoError(t, err)
	require.Empty(t, files)
}

func TestList_RepoError(t *testing.T) {
	mockCompRepo := &mockComposeRepo{
		listByMachineFn: func(_ context.Context, _ string) ([]entity.ComposeFile, error) {
			return nil, errors.ErrNotFound
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, nil, t.TempDir())

	files, err := uc.List(context.Background(), "local")
	require.Error(t, err)
	require.Nil(t, files)
	require.ErrorIs(t, err, errors.ErrNotFound)
}

// --- Get ---

func TestGet_Success(t *testing.T) {
	expected := makeTestComposeFile("proj-1", "local", "my-app", "/tmp/docker-compose.yml")
	expected.Status = entity.OpRunning
	expected.Services = []entity.ComposeService{
		makeTestService("web", "running"),
	}

	mockCompRepo := &mockComposeRepo{
		getByProjectNameFn: func(_ context.Context, machineID, projectName string) (*entity.ComposeFile, error) {
			require.Equal(t, "local", machineID)
			require.Equal(t, "my-app", projectName)
			return expected, nil
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, nil, t.TempDir())

	cf, err := uc.Get(context.Background(), "local", "my-app")
	require.NoError(t, err)
	require.NotNil(t, cf)
	require.Equal(t, "my-app", cf.ProjectName)
	require.Equal(t, entity.OpRunning, cf.Status)
	require.Len(t, cf.Services, 1)
	require.Equal(t, "web", cf.Services[0].Name)
}

func TestGet_NotFound(t *testing.T) {
	mockCompRepo := &mockComposeRepo{
		getByProjectNameFn: func(_ context.Context, _, _ string) (*entity.ComposeFile, error) {
			return nil, errors.ErrNotFound
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, nil, t.TempDir())

	cf, err := uc.Get(context.Background(), "local", "non-existent")
	require.Error(t, err)
	require.Nil(t, cf)
	require.ErrorIs(t, err, errors.ErrNotFound)
}

// --- Create ---

func TestCreate_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "docker-compose.yml")
	yamlContent := "version: '3.8'\nservices:\n  web:\n    image: nginx:latest\n"

	var createdCF *entity.ComposeFile
	mockCompRepo := &mockComposeRepo{
		createFn: func(_ context.Context, cf *entity.ComposeFile) error {
			createdCF = cf
			return nil
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, nil, t.TempDir())

	cf, err := uc.Create(context.Background(), "local", "my-app", yamlContent, filePath)
	require.NoError(t, err)
	require.NotNil(t, cf)
	require.Equal(t, "my-app", cf.Name)
	require.Equal(t, "my-app", cf.ProjectName)
	require.Equal(t, filePath, cf.FilePath)
	require.Equal(t, entity.OpStopped, cf.Status)
	require.NotEmpty(t, cf.ID)

	require.NotNil(t, createdCF, "expected Create to be called on repo")
	require.Equal(t, cf.ID, createdCF.ID)

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Contains(t, string(data), "version: '3.8'")
	require.Contains(t, string(data), "image: nginx:latest")
}

func TestCreate_InvalidName(t *testing.T) {
	uc := newUseCaseForTest(&mockComposeRepo{}, &mockRemoteMachineRepo{}, nil, t.TempDir())

	cf, err := uc.Create(context.Background(), "local", "my app!", "version: '3'", "/tmp/test.yml")
	require.Error(t, err)
	require.Nil(t, cf)
	require.ErrorIs(t, err, errors.ErrInvalidInput)
}

func TestCreate_EmptyYAML(t *testing.T) {
	uc := newUseCaseForTest(&mockComposeRepo{}, &mockRemoteMachineRepo{}, nil, t.TempDir())

	cf, err := uc.Create(context.Background(), "local", "my-app", "   ", "/tmp/test.yml")
	require.Error(t, err)
	require.Nil(t, cf)
	require.Contains(t, err.Error(), "yaml content is empty")
}

// --- Update ---

func TestUpdate_Success(t *testing.T) {
	tmpDir := t.TempDir()
	existingPath := filepath.Join(tmpDir, "docker-compose.yml")
	_ = os.WriteFile(existingPath, []byte("version: '3'"), 0o644)

	updateCalled := false
	mockCompRepo := &mockComposeRepo{
		getByProjectNameFn: func(_ context.Context, machineID, projectName string) (*entity.ComposeFile, error) {
			return makeTestComposeFile("proj-1", machineID, projectName, existingPath), nil
		},
		updateFn: func(_ context.Context, cf *entity.ComposeFile) error {
			updateCalled = true
			return nil
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, nil, t.TempDir())

	err := uc.Update(context.Background(), "local", "my-app", "version: '3.8'\nservices:\n  web:\n    image: nginx:latest")
	require.NoError(t, err)
	require.True(t, updateCalled, "expected Update to be called on repo")

	data, err := os.ReadFile(existingPath)
	require.NoError(t, err)
	require.Contains(t, string(data), "version: '3.8'")
}

func TestUpdate_EmptyYAML(t *testing.T) {
	mockCompRepo := &mockComposeRepo{
		getByProjectNameFn: func(_ context.Context, _, _ string) (*entity.ComposeFile, error) {
			return makeTestComposeFile("proj-1", "local", "my-app", "/tmp/test.yml"), nil
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, nil, t.TempDir())

	err := uc.Update(context.Background(), "local", "my-app", "   ")
	require.Error(t, err)
	require.Contains(t, err.Error(), "yaml content is empty")
}

// --- Delete ---

func TestDelete_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "docker-compose.yml")
	_ = os.WriteFile(filePath, []byte("version: '3'"), 0o644)

	var deletedID string
	mockCompRepo := &mockComposeRepo{
		getByProjectNameFn: func(_ context.Context, _, projectName string) (*entity.ComposeFile, error) {
			return makeTestComposeFile("proj-1", "local", projectName, filePath), nil
		},
		deleteByIDFn: func(_ context.Context, machineID, id string) error {
			deletedID = id
			require.Equal(t, "local", machineID)
			return nil
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, nil, t.TempDir())

	err := uc.Delete(context.Background(), "local", "my-app")
	require.NoError(t, err)
	require.Equal(t, "proj-1", deletedID)

	_, err = os.Stat(filePath)
	require.True(t, os.IsNotExist(err), "expected file to be removed")
}

func TestDelete_FileRemovalIgnored(t *testing.T) {
	mockCompRepo := &mockComposeRepo{
		getByProjectNameFn: func(_ context.Context, _, _ string) (*entity.ComposeFile, error) {
			return makeTestComposeFile("proj-1", "local", "my-app", ""), nil
		},
		deleteByIDFn: func(_ context.Context, _, _ string) error { return nil },
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, nil, t.TempDir())

	err := uc.Delete(context.Background(), "local", "my-app")
	require.NoError(t, err)
}

// --- Up ---

func TestUp_Success(t *testing.T) {
	refreshDone := make(chan struct{}, 1)
	upCalled := false
	updateCalled := false

	mockCli := &mockComposeClient{
		composeUpFn: func(_ context.Context, machineID, projectName, composeFilePath string) error {
			upCalled = true
			require.Equal(t, "local", machineID)
			require.Equal(t, "my-app", projectName)
			return nil
		},
		composePsFn: func(_ context.Context, _, _ string) ([]dockerclient.ComposeServiceStatus, error) {
			return []dockerclient.ComposeServiceStatus{
				{Name: "web", ServiceName: "web", Status: "running", Health: "healthy"},
			}, nil
		},
	}

	mockCompRepo := &mockComposeRepo{
		getByProjectNameFn: func(_ context.Context, _, projectName string) (*entity.ComposeFile, error) {
			return makeTestComposeFile("proj-1", "local", projectName, "/tmp/docker-compose.yml"), nil
		},
		updateFn: func(_ context.Context, cf *entity.ComposeFile) error {
			updateCalled = true
			require.Equal(t, entity.OpRunning, cf.Status)
			return nil
		},
		listByMachineFn: func(_ context.Context, _ string) ([]entity.ComposeFile, error) {
			return []entity.ComposeFile{
				{ID: "proj-1", MachineID: "local", ProjectName: "my-app"},
			}, nil
		},
		upsertBatchFn: func(_ context.Context, cfs []entity.ComposeFile) error {
			require.Len(t, cfs, 1)
			refreshDone <- struct{}{}
			return nil
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, mockCli, t.TempDir())

	err := uc.Up(context.Background(), "local", "my-app")
	require.NoError(t, err)
	require.True(t, upCalled, "expected ComposeUp to be called")
	require.True(t, updateCalled, "expected status update after up")

	select {
	case <-refreshDone:
	case <-time.After(2 * time.Second):
		t.Fatal("refresh goroutine did not complete within 2s")
	}
}

func TestUp_NotFound(t *testing.T) {
	mockCompRepo := &mockComposeRepo{
		getByProjectNameFn: func(_ context.Context, _, _ string) (*entity.ComposeFile, error) {
			return nil, errors.ErrNotFound
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, nil, t.TempDir())

	err := uc.Up(context.Background(), "local", "non-existent")
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrNotFound)
}

// --- Down ---

func TestDown_Success(t *testing.T) {
	downCalled := false
	volumesFlag := false

	mockCli := &mockComposeClient{
		composeDownFn: func(_ context.Context, machineID, projectName string, volumes bool) error {
			downCalled = true
			volumesFlag = volumes
			require.Equal(t, "local", machineID)
			require.Equal(t, "my-app", projectName)
			return nil
		},
	}

	mockCompRepo := &mockComposeRepo{
		getByProjectNameFn: func(_ context.Context, _, projectName string) (*entity.ComposeFile, error) {
			return makeTestComposeFile("proj-1", "local", projectName, "/tmp/docker-compose.yml"), nil
		},
		updateFn: func(_ context.Context, cf *entity.ComposeFile) error {
			require.Equal(t, entity.OpStopped, cf.Status)
			return nil
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, mockCli, t.TempDir())

	err := uc.Down(context.Background(), "local", "my-app", false)
	require.NoError(t, err)
	require.True(t, downCalled, "expected ComposeDown to be called")
	require.False(t, volumesFlag, "expected volumes=false")
}

func TestDown_WithVolumes(t *testing.T) {
	volumesFlag := false

	mockCli := &mockComposeClient{
		composeDownFn: func(_ context.Context, _, _ string, volumes bool) error {
			volumesFlag = volumes
			return nil
		},
	}

	mockCompRepo := &mockComposeRepo{
		getByProjectNameFn: func(_ context.Context, _, projectName string) (*entity.ComposeFile, error) {
			return makeTestComposeFile("proj-1", "local", projectName, "/tmp/docker-compose.yml"), nil
		},
		updateFn: func(_ context.Context, _ *entity.ComposeFile) error { return nil },
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, mockCli, t.TempDir())

	err := uc.Down(context.Background(), "local", "my-app", true)
	require.NoError(t, err)
	require.True(t, volumesFlag, "expected volumes=true")
}

// --- Build ---

func TestBuild_Success(t *testing.T) {
	buildCalled := false

	mockCli := &mockComposeClient{
		composeBuildFn: func(_ context.Context, machineID, composeFilePath string) error {
			buildCalled = true
			require.Equal(t, "local", machineID)
			require.Equal(t, "/tmp/docker-compose.yml", composeFilePath)
			return nil
		},
	}

	mockCompRepo := &mockComposeRepo{
		getByProjectNameFn: func(_ context.Context, _, _ string) (*entity.ComposeFile, error) {
			return makeTestComposeFile("proj-1", "local", "my-app", "/tmp/docker-compose.yml"), nil
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, mockCli, t.TempDir())

	err := uc.Build(context.Background(), "local", "my-app")
	require.NoError(t, err)
	require.True(t, buildCalled, "expected ComposeBuild to be called")
}

// --- Start / Stop / Restart ---

func TestStart_Stop_Restart(t *testing.T) {
	startCalled := false
	stopCalled := false
	restartCalled := false
	updateCalled := 0

	mockCli := &mockComposeClient{
		composeStartFn: func(_ context.Context, _, projectName string) error {
			startCalled = true
			require.Equal(t, "my-app", projectName)
			return nil
		},
		composeStopFn: func(_ context.Context, _, projectName string) error {
			stopCalled = true
			require.Equal(t, "my-app", projectName)
			return nil
		},
		composeRestartFn: func(_ context.Context, _, projectName string) error {
			restartCalled = true
			require.Equal(t, "my-app", projectName)
			return nil
		},
	}

	proj := makeTestComposeFile("proj-1", "local", "my-app", "/tmp/docker-compose.yml")
	mockCompRepo := &mockComposeRepo{
		getByProjectNameFn: func(_ context.Context, _, _ string) (*entity.ComposeFile, error) {
			return proj, nil
		},
		updateFn: func(_ context.Context, _ *entity.ComposeFile) error {
			updateCalled++
			return nil
		},
	}

	uc := newUseCaseForTest(mockCompRepo, &mockRemoteMachineRepo{}, mockCli, t.TempDir())

	err := uc.Start(context.Background(), "local", "my-app")
	require.NoError(t, err)
	require.True(t, startCalled, "expected ComposeStart to be called")
	require.Equal(t, entity.OpRunning, proj.Status)

	err = uc.Stop(context.Background(), "local", "my-app")
	require.NoError(t, err)
	require.True(t, stopCalled, "expected ComposeStop to be called")
	require.Equal(t, entity.OpStopped, proj.Status)

	err = uc.Restart(context.Background(), "local", "my-app")
	require.NoError(t, err)
	require.True(t, restartCalled, "expected ComposeRestart to be called")
	require.Equal(t, entity.OpRunning, proj.Status)

	require.Equal(t, 3, updateCalled, "expected 3 Update calls (Start, Stop, Restart)")
}

// --- Logs ---

func TestLogs_Success(t *testing.T) {
	logContent := "[web] 2024-01-01T00:00:00Z Server started\n[db] 2024-01-01T00:00:01Z Ready for connections\n"

	mockCli := &mockComposeClient{
		composeLogsFn: func(_ context.Context, machineID, projectName string) (io.ReadCloser, error) {
			require.Equal(t, "local", machineID)
			require.Equal(t, "my-app", projectName)
			return io.NopCloser(strings.NewReader(logContent)), nil
		},
	}

	uc := newUseCaseForTest(&mockComposeRepo{}, &mockRemoteMachineRepo{}, mockCli, t.TempDir())

	logs, err := uc.Logs(context.Background(), "local", "my-app")
	require.NoError(t, err)
	require.Equal(t, logContent, logs)
}

func TestLogs_Error(t *testing.T) {
	mockCli := &mockComposeClient{
		composeLogsFn: func(_ context.Context, _, _ string) (io.ReadCloser, error) {
			return nil, errors.ErrDockerOperation
		},
	}

	uc := newUseCaseForTest(&mockComposeRepo{}, &mockRemoteMachineRepo{}, mockCli, t.TempDir())

	logs, err := uc.Logs(context.Background(), "local", "my-app")
	require.Error(t, err)
	require.Empty(t, logs)
	require.ErrorIs(t, err, errors.ErrDockerOperation)
}

// --- Ps ---

func TestPs_Success(t *testing.T) {
	mockCli := &mockComposeClient{
		composePsFn: func(_ context.Context, machineID, projectName string) ([]dockerclient.ComposeServiceStatus, error) {
			require.Equal(t, "local", machineID)
			require.Equal(t, "my-app", projectName)
			return []dockerclient.ComposeServiceStatus{
				{
					Name:        "my-app-web-1",
					ServiceName: "web",
					Image:       "nginx:latest",
					Status:      "running",
					Health:      "healthy",
					Replicas:    1,
					Publishers: []dockerclient.PublishInfo{
						{URL: "0.0.0.0", TargetPort: 80, PublishedPort: 8080},
					},
				},
				{
					Name:        "my-app-db-1",
					ServiceName: "db",
					Image:       "postgres:16",
					Status:      "running",
					Health:      "healthy",
					Replicas:    1,
				},
			}, nil
		},
	}

	uc := newUseCaseForTest(&mockComposeRepo{}, &mockRemoteMachineRepo{}, mockCli, t.TempDir())

	services, err := uc.Ps(context.Background(), "local", "my-app")
	require.NoError(t, err)
	require.Len(t, services, 2)

	require.Equal(t, "my-app-web-1", services[0].Name)
	require.Equal(t, "nginx:latest", services[0].Image)
	require.Equal(t, "running", services[0].Status)
	require.Equal(t, "healthy", services[0].Health)
	require.Equal(t, 1, services[0].Replicas)
	require.Len(t, services[0].Publishers, 1)
	require.Equal(t, "0.0.0.0", services[0].Publishers[0].URL)
	require.Equal(t, 80, services[0].Publishers[0].TargetPort)
	require.Equal(t, 8080, services[0].Publishers[0].PublishedPort)

	require.Equal(t, "my-app-db-1", services[1].Name)
	require.Equal(t, "postgres:16", services[1].Image)
}

func TestPs_Error(t *testing.T) {
	mockCli := &mockComposeClient{
		composePsFn: func(_ context.Context, _, _ string) ([]dockerclient.ComposeServiceStatus, error) {
			return nil, errors.ErrDockerOperation
		},
	}

	uc := newUseCaseForTest(&mockComposeRepo{}, &mockRemoteMachineRepo{}, mockCli, t.TempDir())

	services, err := uc.Ps(context.Background(), "local", "my-app")
	require.Error(t, err)
	require.Nil(t, services)
	require.ErrorIs(t, err, errors.ErrDockerOperation)
}
