package action

import (
	"context"
	"errors"
	"testing"

	"github.com/lminimum/LiteDock/internal/entity"
	dockerImage "github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/require"
)

// --- mockImageUseCase ---

type mockImageUseCase struct {
	listFn  func(ctx context.Context, machineID string) ([]entity.Image, error)
	pruneFn func(ctx context.Context, machineID string) (*dockerImage.PruneReport, error)
}

func (m *mockImageUseCase) List(ctx context.Context, machineID string) ([]entity.Image, error) {
	if m.listFn != nil {
		return m.listFn(ctx, machineID)
	}
	return nil, nil
}

func (m *mockImageUseCase) Prune(ctx context.Context, machineID string) (*dockerImage.PruneReport, error) {
	if m.pruneFn != nil {
		return m.pruneFn(ctx, machineID)
	}
	return nil, nil
}

var _ imageUseCase = (*mockImageUseCase)(nil)

// --- helpers ---

func newImageActionForTest(uc *mockImageUseCase) *ImageAction {
	return NewImageAction(uc)
}

func TestImageAction_Name(t *testing.T) {
	a := NewImageAction(&mockImageUseCase{})
	require.Equal(t, "image", a.Name())
}

func TestImageAction_Description(t *testing.T) {
	a := NewImageAction(&mockImageUseCase{})
	require.Contains(t, a.Description(), "image")
	require.Contains(t, a.Description(), "list")
	require.Contains(t, a.Description(), "prune")
}

func TestImageAction_Params(t *testing.T) {
	a := NewImageAction(&mockImageUseCase{})
	params := a.Params()

	require.Len(t, params, 2)

	names := make(map[string]bool)
	for _, p := range params {
		names[p.Name] = true
	}
	require.True(t, names["operation"])
	require.True(t, names["machine_id"])
}

func TestImageAction_Validate_Success(t *testing.T) {
	a := NewImageAction(&mockImageUseCase{})

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name: "list_images",
			params: map[string]interface{}{
				"operation":  "list_images",
				"machine_id": "local",
			},
		},
		{
			name: "prune_images",
			params: map[string]interface{}{
				"operation":  "prune_images",
				"machine_id": "local",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := a.Validate(tt.params)
			require.NoError(t, err)
		})
	}
}

func TestImageAction_Validate_Error(t *testing.T) {
	a := NewImageAction(&mockImageUseCase{})

	tests := []struct {
		name   string
		params map[string]interface{}
		errMsg string
	}{
		{
			name:   "missing operation",
			params: map[string]interface{}{},
			errMsg: "operation is required",
		},
		{
			name: "empty operation",
			params: map[string]interface{}{
				"operation": "",
			},
			errMsg: "operation is required",
		},
		{
			name: "unknown operation",
			params: map[string]interface{}{
				"operation": "build_image",
			},
			errMsg: "unknown operation: build_image",
		},
		{
			name: "missing machine_id",
			params: map[string]interface{}{
				"operation": "list_images",
			},
			errMsg: "machine_id is required",
		},
		{
			name: "empty machine_id",
			params: map[string]interface{}{
				"operation":  "list_images",
				"machine_id": "",
			},
			errMsg: "machine_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := a.Validate(tt.params)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestImageAction_Execute_List(t *testing.T) {
	expectedImages := []entity.Image{
		{ID: "img1", RepoTags: []string{"nginx:latest"}},
		{ID: "img2", RepoTags: []string{"redis:7"}},
	}

	var capturedMachineID string
	mock := &mockImageUseCase{
		listFn: func(_ context.Context, machineID string) ([]entity.Image, error) {
			capturedMachineID = machineID
			return expectedImages, nil
		},
	}
	a := newImageActionForTest(mock)

	result, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":  "list_images",
		"machine_id": "local",
	})
	require.NoError(t, err)
	require.True(t, result.Success)
	require.NotNil(t, result.Data)

	data, ok := result.Data.(map[string]interface{})
	require.True(t, ok)

	images, ok := data["images"].([]entity.Image)
	require.True(t, ok)
	require.Len(t, images, 2)

	count, ok := data["count"].(int)
	require.True(t, ok)
	require.Equal(t, 2, count)

	require.Equal(t, "local", capturedMachineID)
}

func TestImageAction_Execute_List_Error(t *testing.T) {
	wantErr := errors.New("cannot connect to Docker daemon")
	mock := &mockImageUseCase{
		listFn: func(_ context.Context, _ string) ([]entity.Image, error) {
			return nil, wantErr
		},
	}
	a := newImageActionForTest(mock)

	_, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":  "list_images",
		"machine_id": "local",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, wantErr)
}

func TestImageAction_Execute_List_Empty(t *testing.T) {
	mock := &mockImageUseCase{
		listFn: func(_ context.Context, _ string) ([]entity.Image, error) {
			return []entity.Image{}, nil
		},
	}
	a := newImageActionForTest(mock)

	result, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":  "list_images",
		"machine_id": "local",
	})
	require.NoError(t, err)
	require.True(t, result.Success)

	data := result.Data.(map[string]interface{})
	require.Equal(t, 0, data["count"])
}

func TestImageAction_Execute_Prune(t *testing.T) {
	var capturedMachineID string
	mock := &mockImageUseCase{
		pruneFn: func(_ context.Context, machineID string) (*dockerImage.PruneReport, error) {
			capturedMachineID = machineID
			return &dockerImage.PruneReport{
				ImagesDeleted:  []dockerImage.DeleteResponse{{Deleted: "img1"}},
				SpaceReclaimed: 1048576,
			}, nil
		},
	}
	a := newImageActionForTest(mock)

	result, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":  "prune_images",
		"machine_id": "local",
	})
	require.NoError(t, err)
	require.True(t, result.Success)
	require.NotNil(t, result.Data)

	data, ok := result.Data.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, uint64(1048576), data["space_reclaimed"])
	require.Equal(t, 1, data["images_deleted"])
	require.Equal(t, "local", capturedMachineID)
}

func TestImageAction_Execute_Prune_NoImages(t *testing.T) {
	mock := &mockImageUseCase{
		pruneFn: func(_ context.Context, _ string) (*dockerImage.PruneReport, error) {
			return &dockerImage.PruneReport{
				ImagesDeleted:  []dockerImage.DeleteResponse{},
				SpaceReclaimed: 0,
			}, nil
		},
	}
	a := newImageActionForTest(mock)

	result, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":  "prune_images",
		"machine_id": "local",
	})
	require.NoError(t, err)
	require.True(t, result.Success)

	data := result.Data.(map[string]interface{})
	require.Equal(t, uint64(0), data["space_reclaimed"])
	require.Equal(t, 0, data["images_deleted"])
}

func TestImageAction_Execute_Prune_Error(t *testing.T) {
	wantErr := errors.New("prune failed")
	mock := &mockImageUseCase{
		pruneFn: func(_ context.Context, _ string) (*dockerImage.PruneReport, error) {
			return nil, wantErr
		},
	}
	a := newImageActionForTest(mock)

	_, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":  "prune_images",
		"machine_id": "local",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, wantErr)
}

func TestImageAction_Execute_UnknownOperation(t *testing.T) {
	a := NewImageAction(&mockImageUseCase{})
	_, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":  "nonexistent_op",
		"machine_id": "local",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown image operation")
}

func TestImageAction_RegistryIntegration(t *testing.T) {
	reg := NewActionRegistry()
	a := NewImageAction(&mockImageUseCase{})

	err := reg.Register(a)
	require.NoError(t, err)

	got, ok := reg.Get("image")
	require.True(t, ok)
	require.Equal(t, "image", got.Name())
}

func TestImageAction_RegisterBothWithRegistry(t *testing.T) {
	reg := NewActionRegistry()

	containerAction := NewContainerAction(&mockContainerUseCase{})
	imageAction := NewImageAction(&mockImageUseCase{})

	require.NoError(t, reg.Register(containerAction))
	require.NoError(t, reg.Register(imageAction))

	// Verify both are registered
	_, ok := reg.Get("container")
	require.True(t, ok)

	_, ok = reg.Get("image")
	require.True(t, ok)

	// Verify tool definitions include both
	tools := reg.GenerateToolDefs()
	require.Len(t, tools, 2)
}
