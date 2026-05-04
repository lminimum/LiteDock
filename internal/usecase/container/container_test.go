package container

import (
	"context"
	"fmt"
	"testing"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/stretchr/testify/require"
)

// --- mockLogger ---

type mockLogger struct{}

func (m *mockLogger) Debug(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Info(_ string, _ ...interface{})       {}
func (m *mockLogger) Warn(_ string, _ ...interface{})       {}
func (m *mockLogger) Error(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Fatal(_ interface{}, _ ...interface{}) {}

// --- containerRepo (named interface matching the anonymous repo interface in UseCase) ---

type containerRepo interface {
	List(ctx context.Context) ([]entity.Container, error)
	Get(ctx context.Context, id string) (*entity.Container, error)
	CountAll(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
}

// --- mockContainerRepo ---

type mockContainerRepo struct {
	listFn          func(ctx context.Context) ([]entity.Container, error)
	getFn           func(ctx context.Context, id string) (*entity.Container, error)
	countAllFn      func(ctx context.Context) (int64, error)
	countByStatusFn func(ctx context.Context, status string) (int64, error)
}

func (m *mockContainerRepo) List(ctx context.Context) ([]entity.Container, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockContainerRepo) Get(ctx context.Context, id string) (*entity.Container, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return nil, nil
}

func (m *mockContainerRepo) CountAll(ctx context.Context) (int64, error) {
	if m.countAllFn != nil {
		return m.countAllFn(ctx)
	}
	return 0, nil
}

func (m *mockContainerRepo) CountByStatus(ctx context.Context, status string) (int64, error) {
	if m.countByStatusFn != nil {
		return m.countByStatusFn(ctx, status)
	}
	return 0, nil
}

var _ containerRepo = (*mockContainerRepo)(nil)

// --- helpers ---

func newUseCaseForTest(repo *mockContainerRepo) *UseCase {
	return &UseCase{repo: repo, l: &mockLogger{}}
}

// --- tests ---

func TestList_Success(t *testing.T) {
	tests := []struct {
		name    string
		containers []entity.Container
	}{
		{
			name: "single container",
			containers: []entity.Container{
				{ID: "abc123", Name: "web-app", Image: "nginx:latest", Status: "running"},
			},
		},
		{
			name: "multiple containers",
			containers: []entity.Container{
				{ID: "abc123", Name: "web-app", Image: "nginx:latest", Status: "running"},
				{ID: "def456", Name: "db", Image: "postgres:16", Status: "running"},
				{ID: "ghi789", Name: "cache", Image: "redis:7", Status: "exited"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockContainerRepo{
				listFn: func(_ context.Context) ([]entity.Container, error) {
					return tt.containers, nil
				},
			}
			uc := newUseCaseForTest(repo)

			result, err := uc.List(context.Background())
			require.NoError(t, err)
			require.Len(t, result, len(tt.containers))
			require.Equal(t, tt.containers, result)
		})
	}
}

func TestList_Empty(t *testing.T) {
	repo := &mockContainerRepo{
		listFn: func(_ context.Context) ([]entity.Container, error) {
			return []entity.Container{}, nil
		},
	}
	uc := newUseCaseForTest(repo)

	result, err := uc.List(context.Background())
	require.NoError(t, err)
	require.Empty(t, result)
	require.NotNil(t, result)
}

func TestList_Error(t *testing.T) {
	wantErr := fmt.Errorf("database connection refused")
	repo := &mockContainerRepo{
		listFn: func(_ context.Context) ([]entity.Container, error) {
			return nil, wantErr
		},
	}
	uc := newUseCaseForTest(repo)

	result, err := uc.List(context.Background())
	require.Error(t, err)
	require.ErrorIs(t, err, wantErr)
	require.Nil(t, result)
}

func TestGet_Success(t *testing.T) {
	want := &entity.Container{
		ID: "abc123", Name: "web-app", Image: "nginx:latest", Status: "running",
	}
	repo := &mockContainerRepo{
		getFn: func(_ context.Context, id string) (*entity.Container, error) {
			require.Equal(t, "abc123", id)
			return want, nil
		},
	}
	uc := newUseCaseForTest(repo)

	result, err := uc.Get(context.Background(), "abc123")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, want, result)
}

func TestGet_NotFound(t *testing.T) {
	tests := []struct {
		name    string
		getFn   func(ctx context.Context, id string) (*entity.Container, error)
	}{
		{
			name: "returns nil",
			getFn: func(_ context.Context, _ string) (*entity.Container, error) {
				return nil, nil
			},
		},
		{
			name: "returns error",
			getFn: func(_ context.Context, _ string) (*entity.Container, error) {
				return nil, fmt.Errorf("container not found: does-not-exist")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockContainerRepo{getFn: tt.getFn}
			uc := newUseCaseForTest(repo)

			result, err := uc.Get(context.Background(), "does-not-exist")
			if err != nil {
				require.Nil(t, result)
			} else {
				require.Nil(t, result, "expected nil container for not-found case")
			}
		})
	}
}

func TestCountAll_Success(t *testing.T) {
	repo := &mockContainerRepo{
		countAllFn: func(_ context.Context) (int64, error) {
			return 42, nil
		},
	}
	uc := newUseCaseForTest(repo)

	count, err := uc.CountAll(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(42), count)
}

func TestCountByStatus_Success(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		wantCount int64
	}{
		{
			name:      "running containers",
			status:    "running",
			wantCount: 5,
		},
		{
			name:      "exited containers",
			status:    "exited",
			wantCount: 3,
		},
		{
			name:      "zero containers",
			status:    "paused",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockContainerRepo{
				countByStatusFn: func(_ context.Context, status string) (int64, error) {
					require.Equal(t, tt.status, status)
					return tt.wantCount, nil
				},
			}
			uc := newUseCaseForTest(repo)

			count, err := uc.CountByStatus(context.Background(), tt.status)
			require.NoError(t, err)
			require.Equal(t, tt.wantCount, count)
		})
	}
}
