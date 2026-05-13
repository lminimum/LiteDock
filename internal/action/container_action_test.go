package action

import (
	"context"
	"errors"
	"testing"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/stretchr/testify/require"
)

// --- mockContainerUseCase ---

type mockContainerUseCase struct {
	startFn   func(ctx context.Context, machineID, containerID string) error
	stopFn    func(ctx context.Context, machineID, containerID string) error
	restartFn func(ctx context.Context, machineID, containerID string) error
	logsFn    func(ctx context.Context, machineID, containerID, tail string) (string, error)
	listFn    func(ctx context.Context) ([]entity.Container, error)
}

func (m *mockContainerUseCase) Start(ctx context.Context, machineID, containerID string) error {
	if m.startFn != nil {
		return m.startFn(ctx, machineID, containerID)
	}
	return nil
}

func (m *mockContainerUseCase) Stop(ctx context.Context, machineID, containerID string) error {
	if m.stopFn != nil {
		return m.stopFn(ctx, machineID, containerID)
	}
	return nil
}

func (m *mockContainerUseCase) Restart(ctx context.Context, machineID, containerID string) error {
	if m.restartFn != nil {
		return m.restartFn(ctx, machineID, containerID)
	}
	return nil
}

func (m *mockContainerUseCase) Logs(ctx context.Context, machineID, containerID, tail string) (string, error) {
	if m.logsFn != nil {
		return m.logsFn(ctx, machineID, containerID, tail)
	}
	return "", nil
}

func (m *mockContainerUseCase) List(ctx context.Context) ([]entity.Container, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

var _ containerUseCase = (*mockContainerUseCase)(nil)

// --- helpers ---

func newContainerActionForTest(uc *mockContainerUseCase) *ContainerAction {
	return NewContainerAction(uc)
}

func TestContainerAction_Name(t *testing.T) {
	a := NewContainerAction(&mockContainerUseCase{})
	require.Equal(t, "container", a.Name())
}

func TestContainerAction_Description(t *testing.T) {
	a := NewContainerAction(&mockContainerUseCase{})
	require.Contains(t, a.Description(), "container")
	require.Contains(t, a.Description(), "start")
	require.Contains(t, a.Description(), "stop")
	require.Contains(t, a.Description(), "restart")
	require.Contains(t, a.Description(), "logs")
}

func TestContainerAction_Params(t *testing.T) {
	a := NewContainerAction(&mockContainerUseCase{})
	params := a.Params()

	require.Len(t, params, 4)

	names := make(map[string]bool)
	for _, p := range params {
		names[p.Name] = true
	}
	require.True(t, names["operation"])
	require.True(t, names["machine_id"])
	require.True(t, names["container_id"])
	require.True(t, names["tail"])
}

func TestContainerAction_Validate_Success(t *testing.T) {
	a := NewContainerAction(&mockContainerUseCase{})

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name: "start_container",
			params: map[string]interface{}{
				"operation":    "start_container",
				"machine_id":   "local",
				"container_id": "abc123",
			},
		},
		{
			name: "stop_container",
			params: map[string]interface{}{
				"operation":    "stop_container",
				"machine_id":   "local",
				"container_id": "abc123",
			},
		},
		{
			name: "restart_container",
			params: map[string]interface{}{
				"operation":    "restart_container",
				"machine_id":   "local",
				"container_id": "abc123",
			},
		},
		{
			name: "get_container_logs",
			params: map[string]interface{}{
				"operation":    "get_container_logs",
				"machine_id":   "local",
				"container_id": "abc123",
				"tail":         "50",
			},
		},
		{
			name: "get_container_logs without tail",
			params: map[string]interface{}{
				"operation":    "get_container_logs",
				"machine_id":   "local",
				"container_id": "abc123",
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

func TestContainerAction_Validate_Error(t *testing.T) {
	a := NewContainerAction(&mockContainerUseCase{})

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
				"operation": "invalid_op",
			},
			errMsg: "unknown operation: invalid_op",
		},
		{
			name: "missing machine_id",
			params: map[string]interface{}{
				"operation":    "start_container",
				"container_id": "abc123",
			},
			errMsg: "machine_id is required",
		},
		{
			name: "empty machine_id",
			params: map[string]interface{}{
				"operation":    "start_container",
				"machine_id":   "",
				"container_id": "abc123",
			},
			errMsg: "machine_id is required",
		},
		{
			name: "missing container_id",
			params: map[string]interface{}{
				"operation":  "start_container",
				"machine_id": "local",
			},
			errMsg: "container_id is required",
		},
		{
			name: "empty container_id",
			params: map[string]interface{}{
				"operation":    "start_container",
				"machine_id":   "local",
				"container_id": "",
			},
			errMsg: "container_id is required",
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

func TestContainerAction_Execute_Start(t *testing.T) {
	var capturedMachineID, capturedContainerID string

	mock := &mockContainerUseCase{
		startFn: func(_ context.Context, machineID, containerID string) error {
			capturedMachineID = machineID
			capturedContainerID = containerID
			return nil
		},
	}
	a := newContainerActionForTest(mock)

	result, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":    "start_container",
		"machine_id":   "local",
		"container_id": "my-container",
	})
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Contains(t, result.Message, "my-container")
	require.Contains(t, result.Message, "started")
	require.Equal(t, "local", capturedMachineID)
	require.Equal(t, "my-container", capturedContainerID)
}

func TestContainerAction_Execute_Start_Error(t *testing.T) {
	wantErr := errors.New("docker daemon unavailable")
	mock := &mockContainerUseCase{
		startFn: func(_ context.Context, _, _ string) error {
			return wantErr
		},
	}
	a := newContainerActionForTest(mock)

	_, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":    "start_container",
		"machine_id":   "local",
		"container_id": "my-container",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, wantErr)
}

func TestContainerAction_Execute_Stop(t *testing.T) {
	var capturedMachineID, capturedContainerID string

	mock := &mockContainerUseCase{
		stopFn: func(_ context.Context, machineID, containerID string) error {
			capturedMachineID = machineID
			capturedContainerID = containerID
			return nil
		},
	}
	a := newContainerActionForTest(mock)

	result, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":    "stop_container",
		"machine_id":   "local",
		"container_id": "my-container",
	})
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Contains(t, result.Message, "my-container")
	require.Contains(t, result.Message, "stopped")
	require.Equal(t, "local", capturedMachineID)
	require.Equal(t, "my-container", capturedContainerID)
}

func TestContainerAction_Execute_Stop_Error(t *testing.T) {
	wantErr := errors.New("container not running")
	mock := &mockContainerUseCase{
		stopFn: func(_ context.Context, _, _ string) error {
			return wantErr
		},
	}
	a := newContainerActionForTest(mock)

	_, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":    "stop_container",
		"machine_id":   "local",
		"container_id": "my-container",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, wantErr)
}

func TestContainerAction_Execute_Restart(t *testing.T) {
	var capturedMachineID, capturedContainerID string

	mock := &mockContainerUseCase{
		restartFn: func(_ context.Context, machineID, containerID string) error {
			capturedMachineID = machineID
			capturedContainerID = containerID
			return nil
		},
	}
	a := newContainerActionForTest(mock)

	result, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":    "restart_container",
		"machine_id":   "local",
		"container_id": "my-container",
	})
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Contains(t, result.Message, "my-container")
	require.Contains(t, result.Message, "restarted")
	require.Equal(t, "local", capturedMachineID)
	require.Equal(t, "my-container", capturedContainerID)
}

func TestContainerAction_Execute_Restart_Error(t *testing.T) {
	wantErr := errors.New("failed to restart")
	mock := &mockContainerUseCase{
		restartFn: func(_ context.Context, _, _ string) error {
			return wantErr
		},
	}
	a := newContainerActionForTest(mock)

	_, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":    "restart_container",
		"machine_id":   "local",
		"container_id": "my-container",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, wantErr)
}

func TestContainerAction_Execute_Logs(t *testing.T) {
	var capturedMachineID, capturedContainerID, capturedTail string

	mock := &mockContainerUseCase{
		logsFn: func(_ context.Context, machineID, containerID, tail string) (string, error) {
			capturedMachineID = machineID
			capturedContainerID = containerID
			capturedTail = tail
			return "line1\nline2\nline3\n", nil
		},
	}
	a := newContainerActionForTest(mock)

	result, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":    "get_container_logs",
		"machine_id":   "local",
		"container_id": "web-1",
		"tail":         "50",
	})
	require.NoError(t, err)
	require.True(t, result.Success)
	require.NotNil(t, result.Data)

	data, ok := result.Data.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "line1\nline2\nline3\n", data["logs"])
	require.Equal(t, "local", capturedMachineID)
	require.Equal(t, "web-1", capturedContainerID)
	require.Equal(t, "50", capturedTail)
}

func TestContainerAction_Execute_Logs_DefaultTail(t *testing.T) {
	var capturedTail string

	mock := &mockContainerUseCase{
		logsFn: func(_ context.Context, _, _, tail string) (string, error) {
			capturedTail = tail
			return "log output", nil
		},
	}
	a := newContainerActionForTest(mock)

	result, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":    "get_container_logs",
		"machine_id":   "local",
		"container_id": "web-1",
	})
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Equal(t, "100", capturedTail, "should default to 100 when tail not provided")
}

func TestContainerAction_Execute_Logs_Error(t *testing.T) {
	wantErr := errors.New("container not found")
	mock := &mockContainerUseCase{
		logsFn: func(_ context.Context, _, _, _ string) (string, error) {
			return "", wantErr
		},
	}
	a := newContainerActionForTest(mock)

	_, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":    "get_container_logs",
		"machine_id":   "local",
		"container_id": "nonexistent",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, wantErr)
}

func TestContainerAction_Execute_UnknownOperation(t *testing.T) {
	a := NewContainerAction(&mockContainerUseCase{})
	_, err := a.Execute(context.Background(), map[string]interface{}{
		"operation":    "nonexistent_op",
		"machine_id":   "local",
		"container_id": "abc123",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown container operation")
}

func TestContainerAction_RegistryIntegration(t *testing.T) {
	reg := NewActionRegistry()
	a := NewContainerAction(&mockContainerUseCase{})

	err := reg.Register(a)
	require.NoError(t, err)

	got, ok := reg.Get("container")
	require.True(t, ok)
	require.Equal(t, "container", got.Name())
}
