package action

import (
	"context"
	"testing"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/stretchr/testify/require"
)

// mockNetworkUseCase is a mock implementation of networkUseCase for testing.
type mockNetworkUseCase struct {
	listNetworksFn   func(ctx context.Context, machineID string) ([]entity.Network, error)
	createNetworkFn  func(ctx context.Context, machineID, name, driver string) (*entity.Network, error)
	deleteNetworkFn  func(ctx context.Context, machineID, networkName string) error
	inspectNetworkFn func(ctx context.Context, machineID, networkName string) (*entity.Network, error)
}

func (m *mockNetworkUseCase) ListNetworks(ctx context.Context, machineID string) ([]entity.Network, error) {
	if m.listNetworksFn != nil {
		return m.listNetworksFn(ctx, machineID)
	}
	return nil, nil
}

func (m *mockNetworkUseCase) CreateNetwork(ctx context.Context, machineID, name, driver string) (*entity.Network, error) {
	if m.createNetworkFn != nil {
		return m.createNetworkFn(ctx, machineID, name, driver)
	}
	return nil, nil
}

func (m *mockNetworkUseCase) DeleteNetwork(ctx context.Context, machineID, networkName string) error {
	if m.deleteNetworkFn != nil {
		return m.deleteNetworkFn(ctx, machineID, networkName)
	}
	return nil
}

func (m *mockNetworkUseCase) InspectNetwork(ctx context.Context, machineID, networkName string) (*entity.Network, error) {
	if m.inspectNetworkFn != nil {
		return m.inspectNetworkFn(ctx, machineID, networkName)
	}
	return nil, nil
}

var _ networkUseCase = (*mockNetworkUseCase)(nil)

// mockVolumeUseCase is a mock implementation of volumeUseCase for testing.
type mockVolumeUseCase struct {
	listVolumesFn   func(ctx context.Context, machineID string) ([]entity.Volume, error)
	createVolumeFn  func(ctx context.Context, machineID, name, driver string) (*entity.Volume, error)
	deleteVolumeFn  func(ctx context.Context, machineID, volumeName string) error
	inspectVolumeFn func(ctx context.Context, machineID, volumeName string) (*entity.Volume, error)
}

func (m *mockVolumeUseCase) ListVolumes(ctx context.Context, machineID string) ([]entity.Volume, error) {
	if m.listVolumesFn != nil {
		return m.listVolumesFn(ctx, machineID)
	}
	return nil, nil
}

func (m *mockVolumeUseCase) CreateVolume(ctx context.Context, machineID, name, driver string) (*entity.Volume, error) {
	if m.createVolumeFn != nil {
		return m.createVolumeFn(ctx, machineID, name, driver)
	}
	return nil, nil
}

func (m *mockVolumeUseCase) DeleteVolume(ctx context.Context, machineID, volumeName string) error {
	if m.deleteVolumeFn != nil {
		return m.deleteVolumeFn(ctx, machineID, volumeName)
	}
	return nil
}

func (m *mockVolumeUseCase) InspectVolume(ctx context.Context, machineID, volumeName string) (*entity.Volume, error) {
	if m.inspectVolumeFn != nil {
		return m.inspectVolumeFn(ctx, machineID, volumeName)
	}
	return nil, nil
}

var _ volumeUseCase = (*mockVolumeUseCase)(nil)

// mockComposeUseCase is a mock implementation of composeUseCase for testing.
type mockComposeUseCase struct {
	listFn       func(ctx context.Context, machineID string) ([]entity.ComposeFile, error)
	getFn        func(ctx context.Context, machineID, projectName string) (*entity.ComposeFile, error)
	createFn     func(ctx context.Context, machineID, name, yamlContent, filePath string) (*entity.ComposeFile, error)
	updateFn     func(ctx context.Context, machineID, projectName, yamlContent string) error
	deleteFn     func(ctx context.Context, machineID, projectName string) error
	upFn         func(ctx context.Context, machineID, projectName string) error
	downFn       func(ctx context.Context, machineID, projectName string, volumes bool) error
	buildFn      func(ctx context.Context, machineID, projectName string) error
	startFn      func(ctx context.Context, machineID, projectName string) error
	stopFn       func(ctx context.Context, machineID, projectName string) error
	restartFn    func(ctx context.Context, machineID, projectName string) error
	logsFn       func(ctx context.Context, machineID, projectName string) (string, error)
	psFn         func(ctx context.Context, machineID, projectName string) ([]entity.ComposeService, error)
}

func (m *mockComposeUseCase) List(ctx context.Context, machineID string) ([]entity.ComposeFile, error) {
	if m.listFn != nil {
		return m.listFn(ctx, machineID)
	}
	return nil, nil
}

func (m *mockComposeUseCase) Get(ctx context.Context, machineID, projectName string) (*entity.ComposeFile, error) {
	if m.getFn != nil {
		return m.getFn(ctx, machineID, projectName)
	}
	return nil, nil
}

func (m *mockComposeUseCase) Create(ctx context.Context, machineID, name, yamlContent, filePath string) (*entity.ComposeFile, error) {
	if m.createFn != nil {
		return m.createFn(ctx, machineID, name, yamlContent, filePath)
	}
	return nil, nil
}

func (m *mockComposeUseCase) Update(ctx context.Context, machineID, projectName, yamlContent string) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, machineID, projectName, yamlContent)
	}
	return nil
}

func (m *mockComposeUseCase) Delete(ctx context.Context, machineID, projectName string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, machineID, projectName)
	}
	return nil
}

func (m *mockComposeUseCase) Up(ctx context.Context, machineID, projectName string) error {
	if m.upFn != nil {
		return m.upFn(ctx, machineID, projectName)
	}
	return nil
}

func (m *mockComposeUseCase) Down(ctx context.Context, machineID, projectName string, volumes bool) error {
	if m.downFn != nil {
		return m.downFn(ctx, machineID, projectName, volumes)
	}
	return nil
}

func (m *mockComposeUseCase) Build(ctx context.Context, machineID, projectName string) error {
	if m.buildFn != nil {
		return m.buildFn(ctx, machineID, projectName)
	}
	return nil
}

func (m *mockComposeUseCase) Start(ctx context.Context, machineID, projectName string) error {
	if m.startFn != nil {
		return m.startFn(ctx, machineID, projectName)
	}
	return nil
}

func (m *mockComposeUseCase) Stop(ctx context.Context, machineID, projectName string) error {
	if m.stopFn != nil {
		return m.stopFn(ctx, machineID, projectName)
	}
	return nil
}

func (m *mockComposeUseCase) Restart(ctx context.Context, machineID, projectName string) error {
	if m.restartFn != nil {
		return m.restartFn(ctx, machineID, projectName)
	}
	return nil
}

func (m *mockComposeUseCase) Logs(ctx context.Context, machineID, projectName string) (string, error) {
	if m.logsFn != nil {
		return m.logsFn(ctx, machineID, projectName)
	}
	return "", nil
}

func (m *mockComposeUseCase) Ps(ctx context.Context, machineID, projectName string) ([]entity.ComposeService, error) {
	if m.psFn != nil {
		return m.psFn(ctx, machineID, projectName)
	}
	return nil, nil
}

var _ composeUseCase = (*mockComposeUseCase)(nil)

// mockMachineUseCase is a mock implementation of machineUseCase for testing.
type mockMachineUseCase struct {
	listFn             func(ctx context.Context) ([]entity.RemoteMachine, error)
	getByIDFn         func(ctx context.Context, id string) (*entity.RemoteMachine, error)
	testConnectionFn   func(ctx context.Context, machineID string) error
	listContainersFn   func(ctx context.Context, machineID string) ([]entity.Container, error)
}

func (m *mockMachineUseCase) List(ctx context.Context) ([]entity.RemoteMachine, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockMachineUseCase) GetByID(ctx context.Context, id string) (*entity.RemoteMachine, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockMachineUseCase) TestConnection(ctx context.Context, machineID string) error {
	if m.testConnectionFn != nil {
		return m.testConnectionFn(ctx, machineID)
	}
	return nil
}

func (m *mockMachineUseCase) ListContainers(ctx context.Context, machineID string) ([]entity.Container, error) {
	if m.listContainersFn != nil {
		return m.listContainersFn(ctx, machineID)
	}
	return nil, nil
}

var _ machineUseCase = (*mockMachineUseCase)(nil)

// TestNewActions_GenerateToolDefs_ContainsAllFour verifies that all 4 new actions
// (network, volume, compose, machine) are properly registered and can generate tool definitions.
func TestNewActions_GenerateToolDefs_ContainsAllFour(t *testing.T) {
	// Create mock use cases
	networkUC := &mockNetworkUseCase{}
	volumeUC := &mockVolumeUseCase{}
	composeUC := &mockComposeUseCase{}
	machineUC := &mockMachineUseCase{}

	// Instantiate the 4 new actions
	networkAction := NewNetworkAction(networkUC)
	volumeAction := NewVolumeAction(volumeUC)
	composeAction := NewComposeAction(composeUC)
	machineAction := NewMachineAction(machineUC)

	// Verify actions implement the Action interface
	var _ Action = networkAction
	var _ Action = volumeAction
	var _ Action = composeAction
	var _ Action = machineAction

	// Register with a test registry
	reg := NewActionRegistry()
	require.NoError(t, reg.Register(networkAction))
	require.NoError(t, reg.Register(volumeAction))
	require.NoError(t, reg.Register(composeAction))
	require.NoError(t, reg.Register(machineAction))

	// Generate tool definitions
	toolDefs := reg.GenerateToolDefs()

	// Build a map of action names for easy verification
	actionNames := make(map[string]bool)
	for _, tool := range toolDefs {
		funcMap, ok := tool["function"].(map[string]interface{})
		require.True(t, ok, "tool should have 'function' key")
		name, ok := funcMap["name"].(string)
		require.True(t, ok, "function should have 'name' key")
		actionNames[name] = true
	}

	// Verify all 4 new action names are present
	require.True(t, actionNames["network"], "tool defs should contain 'network' action")
	require.True(t, actionNames["volume"], "tool defs should contain 'volume' action")
	require.True(t, actionNames["compose"], "tool defs should contain 'compose' action")
	require.True(t, actionNames["machine"], "tool defs should contain 'machine' action")

	// Verify total count (should be exactly 4)
	require.Len(t, toolDefs, 4, "should have exactly 4 tool definitions")
}

// TestNewActions_Params verifies that each new action has proper parameter definitions.
func TestNewActions_Params(t *testing.T) {
	networkUC := &mockNetworkUseCase{}
	volumeUC := &mockVolumeUseCase{}
	composeUC := &mockComposeUseCase{}
	machineUC := &mockMachineUseCase{}

	networkAction := NewNetworkAction(networkUC)
	volumeAction := NewVolumeAction(volumeUC)
	composeAction := NewComposeAction(composeUC)
	machineAction := NewMachineAction(machineUC)

	// Verify each action has params
	require.NotEmpty(t, networkAction.Params(), "network action should have params")
	require.NotEmpty(t, volumeAction.Params(), "volume action should have params")
	require.NotEmpty(t, composeAction.Params(), "compose action should have params")
	require.NotEmpty(t, machineAction.Params(), "machine action should have params")

	// Verify operation param is required for all
	for _, action := range []Action{networkAction, volumeAction, composeAction, machineAction} {
		params := action.Params()
		hasOperation := false
		for _, p := range params {
			if p.Name == "operation" && p.Required {
				hasOperation = true
				break
			}
		}
		require.True(t, hasOperation, "%s should have required 'operation' param", action.Name())
	}
}

// TestNewActions_Name verifies that each new action has the correct name.
func TestNewActions_Name(t *testing.T) {
	networkUC := &mockNetworkUseCase{}
	volumeUC := &mockVolumeUseCase{}
	composeUC := &mockComposeUseCase{}
	machineUC := &mockMachineUseCase{}

	require.Equal(t, "network", NewNetworkAction(networkUC).Name())
	require.Equal(t, "volume", NewVolumeAction(volumeUC).Name())
	require.Equal(t, "compose", NewComposeAction(composeUC).Name())
	require.Equal(t, "machine", NewMachineAction(machineUC).Name())
}

// TestNewActions_Description verifies that each new action has a non-empty description.
func TestNewActions_Description(t *testing.T) {
	networkUC := &mockNetworkUseCase{}
	volumeUC := &mockVolumeUseCase{}
	composeUC := &mockComposeUseCase{}
	machineUC := &mockMachineUseCase{}

	networkAction := NewNetworkAction(networkUC)
	volumeAction := NewVolumeAction(volumeUC)
	composeAction := NewComposeAction(composeUC)
	machineAction := NewMachineAction(machineUC)

	require.NotEmpty(t, networkAction.Description())
	require.NotEmpty(t, volumeAction.Description())
	require.NotEmpty(t, composeAction.Description())
	require.NotEmpty(t, machineAction.Description())
}

// TestNewActions_Destructive verifies destructive action detection.
func TestNewActions_Destructive(t *testing.T) {
	networkUC := &mockNetworkUseCase{}
	volumeUC := &mockVolumeUseCase{}
	composeUC := &mockComposeUseCase{}
	machineUC := &mockMachineUseCase{}

	networkAction := NewNetworkAction(networkUC)
	volumeAction := NewVolumeAction(volumeUC)
	composeAction := NewComposeAction(composeUC)
	machineAction := NewMachineAction(machineUC)

	// delete_network should be destructive
	require.True(t, networkAction.Destructive(map[string]interface{}{"operation": "delete_network"}))
	// list_networks should not be destructive
	require.False(t, networkAction.Destructive(map[string]interface{}{"operation": "list_networks"}))

	// delete_volume should be destructive
	require.True(t, volumeAction.Destructive(map[string]interface{}{"operation": "delete_volume"}))
	// list_volumes should not be destructive
	require.False(t, volumeAction.Destructive(map[string]interface{}{"operation": "list_volumes"}))

	// compose operations that change state should be destructive
	require.True(t, composeAction.Destructive(map[string]interface{}{"operation": "compose_up"}))
	require.True(t, composeAction.Destructive(map[string]interface{}{"operation": "compose_down"}))
	require.True(t, composeAction.Destructive(map[string]interface{}{"operation": "compose_start"}))
	require.True(t, composeAction.Destructive(map[string]interface{}{"operation": "compose_stop"}))
	// compose_ps should not be destructive
	require.False(t, composeAction.Destructive(map[string]interface{}{"operation": "compose_ps"}))

	// machine operations are all read-only
	require.False(t, machineAction.Destructive(map[string]interface{}{"operation": "list_machines"}))
	require.False(t, machineAction.Destructive(map[string]interface{}{"operation": "inspect_machine"}))
}

// TestNewActions_Validate verifies parameter validation.
func TestNewActions_Validate(t *testing.T) {
	networkUC := &mockNetworkUseCase{}
	volumeUC := &mockVolumeUseCase{}
	composeUC := &mockComposeUseCase{}
	machineUC := &mockMachineUseCase{}

	networkAction := NewNetworkAction(networkUC)
	volumeAction := NewVolumeAction(volumeUC)
	composeAction := NewComposeAction(composeUC)
	machineAction := NewMachineAction(machineUC)

	// Empty params should fail validation (operation is required)
	require.Error(t, networkAction.Validate(map[string]interface{}{}))
	require.Error(t, volumeAction.Validate(map[string]interface{}{}))
	require.Error(t, composeAction.Validate(map[string]interface{}{}))
	require.Error(t, machineAction.Validate(map[string]interface{}{}))

	// Valid params should pass
	require.NoError(t, networkAction.Validate(map[string]interface{}{
		"operation": "list_networks",
		"machine_id": "local",
	}))
	require.NoError(t, volumeAction.Validate(map[string]interface{}{
		"operation": "list_volumes",
		"machine_id": "local",
	}))
	require.NoError(t, composeAction.Validate(map[string]interface{}{
		"operation": "list_compose",
		"machine_id": "local",
		"project_name": "myproject",
	}))
	require.NoError(t, machineAction.Validate(map[string]interface{}{
		"operation": "list_machines",
		"machine_id": "local",
	}))
}

// TestNewActions_ConfirmationMessage verifies confirmation messages are generated.
func TestNewActions_ConfirmationMessage(t *testing.T) {
	networkUC := &mockNetworkUseCase{}
	volumeUC := &mockVolumeUseCase{}
	composeUC := &mockComposeUseCase{}
	machineUC := &mockMachineUseCase{}

	networkAction := NewNetworkAction(networkUC)
	volumeAction := NewVolumeAction(volumeUC)
	composeAction := NewComposeAction(composeUC)
	machineAction := NewMachineAction(machineUC)

	// Each action should return a non-empty confirmation message
	require.NotEmpty(t, networkAction.ConfirmationMessage(map[string]interface{}{
		"operation": "delete_network",
		"network_id": "test-network",
	}))
	require.NotEmpty(t, volumeAction.ConfirmationMessage(map[string]interface{}{
		"operation": "delete_volume",
		"volume_name": "test-volume",
	}))
	require.NotEmpty(t, composeAction.ConfirmationMessage(map[string]interface{}{
		"operation": "compose_up",
		"project_name": "test-project",
	}))
	require.NotEmpty(t, machineAction.ConfirmationMessage(map[string]interface{}{
		"operation": "list_machines",
	}))
}
