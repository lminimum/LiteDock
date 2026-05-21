package action

import (
	"context"
	"fmt"

	"github.com/lminimum/LiteDock/internal/entity"
)

// volumeUseCase defines the volume use case methods needed by VolumeAction.
type volumeUseCase interface {
	ListVolumes(ctx context.Context, machineID string) ([]entity.Volume, error)
	CreateVolume(ctx context.Context, machineID, name, driver string) (*entity.Volume, error)
	DeleteVolume(ctx context.Context, machineID, volumeName string) error
	InspectVolume(ctx context.Context, machineID, volumeName string) (*entity.Volume, error)
}

// volumeOperations lists the supported volume sub-actions.
var volumeOperations = []string{
	"list_volumes",
	"inspect_volume",
	"create_volume",
	"delete_volume",
}

// VolumeAction exposes volume management operations as AI-callable actions.
type VolumeAction struct {
	uc volumeUseCase
}

// NewVolumeAction creates a new VolumeAction wrapping the given use case.
func NewVolumeAction(uc volumeUseCase) *VolumeAction {
	return &VolumeAction{uc: uc}
}

// Name returns the action name.
func (a *VolumeAction) Name() string {
	return "volume"
}

// Description returns a human-readable description of the action.
func (a *VolumeAction) Description() string {
	return "Manage Docker volumes: list, inspect, create, and delete"
}

// Params returns the parameter definitions for the volume action.
func (a *VolumeAction) Params() []ParamDef {
	return []ParamDef{
		{
			Name:        "operation",
			Type:        "string",
			Required:    true,
			Description: "The volume operation to perform: list_volumes, inspect_volume, create_volume, delete_volume",
		},
		{
			Name:        "machine_id",
			Type:        "string",
			Required:    true,
			Description: "The ID of the target machine (use 'local' for local Docker)",
		},
		{
			Name:        "volume_name",
			Type:        "string",
			Required:    true,
			Description: "The volume name (required for inspect, create, delete)",
		},
		{
			Name:        "driver",
			Type:        "string",
			Required:    false,
			Description: "The volume driver (default: local, only for create)",
			Default:     "local",
		},
	}
}

// Validate checks that the provided parameters are valid.
func (a *VolumeAction) Validate(params map[string]interface{}) error {
	op, _ := params["operation"].(string)
	if op == "" {
		return fmt.Errorf("operation is required")
	}

	valid := false
	for _, validOp := range volumeOperations {
		if op == validOp {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unknown operation: %s, supported: %v", op, volumeOperations)
	}

	machineID, _ := params["machine_id"].(string)
	if machineID == "" {
		return fmt.Errorf("machine_id is required")
	}

	volumeName, _ := params["volume_name"].(string)
	if volumeName == "" && (op == "inspect_volume" || op == "create_volume" || op == "delete_volume") {
		return fmt.Errorf("volume_name is required for %s", op)
	}

	return nil
}

// Destructive returns true if the volume operation requires confirmation.
func (a *VolumeAction) Destructive(params map[string]interface{}) bool {
	op, _ := params["operation"].(string)
	switch op {
	case "delete_volume":
		return true
	default:
		return false
	}
}

// ConfirmationMessage returns a human-readable message about what the action will do.
func (a *VolumeAction) ConfirmationMessage(params map[string]interface{}) string {
	op, _ := params["operation"].(string)
	volumeName, _ := params["volume_name"].(string)
	if volumeName == "" {
		volumeName = "unknown"
	}
	switch op {
	case "delete_volume":
		return fmt.Sprintf("This will permanently delete volume '%s'.", volumeName)
	default:
		return fmt.Sprintf("Execute %s on volume '%s'.", op, volumeName)
	}
}

// Execute runs the requested volume operation.
func (a *VolumeAction) Execute(ctx context.Context, params map[string]interface{}) (*ActionResult, error) {
	op, _ := params["operation"].(string)
	machineID, _ := params["machine_id"].(string)
	volumeName, _ := params["volume_name"].(string)
	driver, _ := params["driver"].(string)
	if driver == "" {
		driver = "local"
	}

	switch op {
	case "list_volumes":
		volumes, err := a.uc.ListVolumes(ctx, machineID)
		if err != nil {
			return nil, fmt.Errorf("VolumeAction.Execute - list_volumes: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: "Volumes listed successfully",
			Data: map[string]interface{}{
				"volumes": volumes,
			},
		}, nil

	case "inspect_volume":
		volume, err := a.uc.InspectVolume(ctx, machineID, volumeName)
		if err != nil {
			return nil, fmt.Errorf("VolumeAction.Execute - inspect_volume: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Volume '%s' inspected successfully", volumeName),
			Data: map[string]interface{}{
				"volume": volume,
			},
		}, nil

	case "create_volume":
		volume, err := a.uc.CreateVolume(ctx, machineID, volumeName, driver)
		if err != nil {
			return nil, fmt.Errorf("VolumeAction.Execute - create_volume: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Volume '%s' created successfully", volumeName),
			Data: map[string]interface{}{
				"volume": volume,
			},
		}, nil

	case "delete_volume":
		if err := a.uc.DeleteVolume(ctx, machineID, volumeName); err != nil {
			return nil, fmt.Errorf("VolumeAction.Execute - delete_volume: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Volume '%s' deleted successfully", volumeName),
		}, nil

	default:
		return nil, fmt.Errorf("unknown volume operation: %s", op)
	}
}
