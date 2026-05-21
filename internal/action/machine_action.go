package action

import (
	"context"
	"fmt"

	"github.com/lminimum/LiteDock/internal/entity"
)

// machineUseCase defines the machine use case methods needed by MachineAction.
type machineUseCase interface {
	List(ctx context.Context) ([]entity.RemoteMachine, error)
	GetByID(ctx context.Context, id string) (*entity.RemoteMachine, error)
	TestConnection(ctx context.Context, machineID string) error
	ListContainers(ctx context.Context, machineID string) ([]entity.Container, error)
}

// machineOperations lists the supported machine sub-actions.
var machineOperations = []string{
	"list_machines",
	"inspect_machine",
	"test_machine_connection",
	"list_machine_containers",
}

// MachineAction exposes machine management operations as AI-callable actions.
type MachineAction struct {
	uc machineUseCase
}

// NewMachineAction creates a new MachineAction wrapping the given use case.
func NewMachineAction(uc machineUseCase) *MachineAction {
	return &MachineAction{uc: uc}
}

// Name returns the action name.
func (a *MachineAction) Name() string {
	return "machine"
}

// Description returns a human-readable description of the action.
func (a *MachineAction) Description() string {
	return "Manage remote machines: list, inspect, test connection, and list containers"
}

// Params returns the parameter definitions for the machine action.
func (a *MachineAction) Params() []ParamDef {
	return []ParamDef{
		{
			Name:        "operation",
			Type:        "string",
			Required:    true,
			Description: "The machine operation to perform: list_machines, inspect_machine, test_machine_connection, list_machine_containers",
		},
		{
			Name:        "machine_id",
			Type:        "string",
			Required:    true,
			Description: "The ID of the target machine",
		},
		{
			Name:        "machine_name",
			Type:        "string",
			Required:    false,
			Description: "The name of the machine (optional, for inspect by name)",
		},
	}
}

// Validate checks that the provided parameters are valid.
func (a *MachineAction) Validate(params map[string]interface{}) error {
	op, _ := params["operation"].(string)
	if op == "" {
		return fmt.Errorf("operation is required")
	}

	valid := false
	for _, validOp := range machineOperations {
		if op == validOp {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unknown operation: %s, supported: %v", op, machineOperations)
	}

	machineID, _ := params["machine_id"].(string)
	if machineID == "" {
		return fmt.Errorf("machine_id is required")
	}

	return nil
}

// Destructive returns false as all machine operations are read-only.
func (a *MachineAction) Destructive(params map[string]interface{}) bool {
	return false
}

// ConfirmationMessage returns a human-readable message about what the action will do.
func (a *MachineAction) ConfirmationMessage(params map[string]interface{}) string {
	op, _ := params["operation"].(string)
	machineID, _ := params["machine_id"].(string)
	if machineID == "" {
		machineID = "unknown"
	}
	return fmt.Sprintf("Execute %s on machine '%s'.", op, machineID)
}

// Execute runs the requested machine operation.
func (a *MachineAction) Execute(ctx context.Context, params map[string]interface{}) (*ActionResult, error) {
	op, _ := params["operation"].(string)
	machineID, _ := params["machine_id"].(string)

	switch op {
	case "list_machines":
		machines, err := a.uc.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("MachineAction.Execute - list_machines: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: "Machines retrieved successfully",
			Data: map[string]interface{}{
				"machines": machines,
			},
		}, nil

	case "inspect_machine":
		machine, err := a.uc.GetByID(ctx, machineID)
		if err != nil {
			return nil, fmt.Errorf("MachineAction.Execute - inspect_machine: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: "Machine inspected successfully",
			Data: map[string]interface{}{
				"machine": machine,
			},
		}, nil

	case "test_machine_connection":
		if err := a.uc.TestConnection(ctx, machineID); err != nil {
			return &ActionResult{
				Success: false,
				Message: "Connection failed",
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			}, nil
		}
		return &ActionResult{
			Success: true,
			Message: "Connection successful",
		}, nil

	case "list_machine_containers":
		containers, err := a.uc.ListContainers(ctx, machineID)
		if err != nil {
			return nil, fmt.Errorf("MachineAction.Execute - list_machine_containers: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: "Containers retrieved successfully",
			Data: map[string]interface{}{
				"containers": containers,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown machine operation: %s", op)
	}
}
