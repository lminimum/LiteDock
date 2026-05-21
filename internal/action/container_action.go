package action

import (
	"context"
	"fmt"

	"github.com/lminimum/LiteDock/internal/entity"
)

// containerUseCase defines the container use case methods needed by ContainerAction.
type containerUseCase interface {
	Start(ctx context.Context, machineID, containerID string) error
	Stop(ctx context.Context, machineID, containerID string) error
	Restart(ctx context.Context, machineID, containerID string) error
	Logs(ctx context.Context, machineID, containerID, tail string) (string, error)
	List(ctx context.Context) ([]entity.Container, error)
}

// containerOperations lists the supported container sub-actions.
var containerOperations = []string{
	"start_container",
	"stop_container",
	"restart_container",
	"get_container_logs",
}

// ContainerAction exposes container management operations as AI-callable actions.
type ContainerAction struct {
	uc containerUseCase
}

// NewContainerAction creates a new ContainerAction wrapping the given use case.
func NewContainerAction(uc containerUseCase) *ContainerAction {
	return &ContainerAction{uc: uc}
}

// Name returns the action name.
func (a *ContainerAction) Name() string {
	return "container"
}

// Description returns a human-readable description of the action.
func (a *ContainerAction) Description() string {
	return "Manage Docker containers: start, stop, restart, and view logs"
}

// Params returns the parameter definitions for the container action.
func (a *ContainerAction) Params() []ParamDef {
	return []ParamDef{
		{
			Name:        "operation",
			Type:        "string",
			Required:    true,
			Description: "The container operation to perform: start_container, stop_container, restart_container, get_container_logs",
		},
		{
			Name:        "machine_id",
			Type:        "string",
			Required:    true,
			Description: "The ID of the target machine (use 'local' for local Docker)",
		},
		{
			Name:        "container_id",
			Type:        "string",
			Required:    true,
			Description: "The container ID or name",
		},
		{
			Name:        "tail",
			Type:        "string",
			Required:    false,
			Description: "Number of log lines to return (only for get_container_logs, default: 100)",
			Default:     "100",
		},
	}
}

// Validate checks that the provided parameters are valid.
func (a *ContainerAction) Validate(params map[string]interface{}) error {
	op, _ := params["operation"].(string)
	if op == "" {
		return fmt.Errorf("operation is required")
	}

	valid := false
	for _, validOp := range containerOperations {
		if op == validOp {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unknown operation: %s, supported: %v", op, containerOperations)
	}

	machineID, _ := params["machine_id"].(string)
	if machineID == "" {
		return fmt.Errorf("machine_id is required")
	}

	containerID, _ := params["container_id"].(string)
	if containerID == "" {
		return fmt.Errorf("container_id is required")
	}

	return nil
}

// Destructive returns true if the container operation requires confirmation.
// Stop and restart can cause data loss if not graceful.
func (a *ContainerAction) Destructive(params map[string]interface{}) bool {
	op, _ := params["operation"].(string)
	switch op {
	case "stop_container", "restart_container":
		return true
	default:
		return false
	}
}

// ConfirmationMessage returns a human-readable message about what the action will do.
func (a *ContainerAction) ConfirmationMessage(params map[string]interface{}) string {
	op, _ := params["operation"].(string)
	containerID, _ := params["container_id"].(string)
	if containerID == "" {
		containerID = "unknown"
	}
	switch op {
	case "stop_container":
		return fmt.Sprintf("This will stop container '%s'.", containerID)
	case "restart_container":
		return fmt.Sprintf("This will restart container '%s'.", containerID)
	default:
		return fmt.Sprintf("Execute %s on container '%s'.", op, containerID)
	}
}

// Execute runs the requested container operation.
func (a *ContainerAction) Execute(ctx context.Context, params map[string]interface{}) (*ActionResult, error) {
	op, _ := params["operation"].(string)
	machineID, _ := params["machine_id"].(string)
	containerID, _ := params["container_id"].(string)

	switch op {
	case "start_container":
		if err := a.uc.Start(ctx, machineID, containerID); err != nil {
			return nil, fmt.Errorf("ContainerAction.Execute - start_container: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Container %s started successfully", containerID),
		}, nil

	case "stop_container":
		if err := a.uc.Stop(ctx, machineID, containerID); err != nil {
			return nil, fmt.Errorf("ContainerAction.Execute - stop_container: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Container %s stopped successfully", containerID),
		}, nil

	case "restart_container":
		if err := a.uc.Restart(ctx, machineID, containerID); err != nil {
			return nil, fmt.Errorf("ContainerAction.Execute - restart_container: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Container %s restarted successfully", containerID),
		}, nil

	case "get_container_logs":
		tail, _ := params["tail"].(string)
		if tail == "" {
			tail = "100"
		}
		logs, err := a.uc.Logs(ctx, machineID, containerID, tail)
		if err != nil {
			return nil, fmt.Errorf("ContainerAction.Execute - get_container_logs: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: "Container logs retrieved successfully",
			Data: map[string]interface{}{
				"logs": logs,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown container operation: %s", op)
	}
}
