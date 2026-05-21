package action

import (
	"context"
	"fmt"

	"github.com/lminimum/LiteDock/internal/entity"
)

// composeUseCase defines the compose use case methods needed by ComposeAction.
type composeUseCase interface {
	List(ctx context.Context, machineID string) ([]entity.ComposeFile, error)
	Get(ctx context.Context, machineID, projectName string) (*entity.ComposeFile, error)
	Create(ctx context.Context, machineID, name, yamlContent, filePath string) (*entity.ComposeFile, error)
	Update(ctx context.Context, machineID, projectName, yamlContent string) error
	Delete(ctx context.Context, machineID, projectName string) error
	Up(ctx context.Context, machineID, projectName string) error
	Down(ctx context.Context, machineID, projectName string, volumes bool) error
	Build(ctx context.Context, machineID, projectName string) error
	Start(ctx context.Context, machineID, projectName string) error
	Stop(ctx context.Context, machineID, projectName string) error
	Restart(ctx context.Context, machineID, projectName string) error
	Logs(ctx context.Context, machineID, projectName string) (string, error)
	Ps(ctx context.Context, machineID, projectName string) ([]entity.ComposeService, error)
}

// composeOperations lists the supported compose sub-actions.
var composeOperations = []string{
	"list_compose",
	"compose_ps",
	"compose_logs",
	"compose_up",
	"compose_down",
	"compose_start",
	"compose_stop",
	"compose_restart",
	"compose_build",
}

// ComposeAction exposes compose management operations as AI-callable actions.
type ComposeAction struct {
	uc composeUseCase
}

// NewComposeAction creates a new ComposeAction wrapping the given use case.
func NewComposeAction(uc composeUseCase) *ComposeAction {
	return &ComposeAction{uc: uc}
}

// Name returns the action name.
func (a *ComposeAction) Name() string {
	return "compose"
}

// Description returns a human-readable description of the action.
func (a *ComposeAction) Description() string {
	return "Manage Docker Compose projects: list, start, stop, restart, view logs and status"
}

// Params returns the parameter definitions for the compose action.
func (a *ComposeAction) Params() []ParamDef {
	return []ParamDef{
		{
			Name:        "operation",
			Type:        "string",
			Required:    true,
			Description: "The compose operation to perform: list_compose, compose_ps, compose_logs, compose_up, compose_down, compose_start, compose_stop, compose_restart, compose_build",
		},
		{
			Name:        "machine_id",
			Type:        "string",
			Required:    true,
			Description: "The ID of the target machine (use 'local' for local Docker)",
		},
		{
			Name:        "project_name",
			Type:        "string",
			Required:    true,
			Description: "The name of the compose project",
		},
		{
			Name:        "yaml_content",
			Type:        "string",
			Required:    false,
			Description: "The compose file YAML content (for compose_up with create/update)",
		},
		{
			Name:        "file_path",
			Type:        "string",
			Required:    false,
			Description: "Path to the compose file (for compose_up, alternative to yaml_content)",
		},
		{
			Name:        "volumes",
			Type:        "boolean",
			Required:    false,
			Description: "Remove volumes when stopping (for compose_down, default: false)",
			Default:     "false",
		},
		{
			Name:        "tail",
			Type:        "string",
			Required:    false,
			Description: "Number of log lines to return (only for compose_logs, default: 100)",
			Default:     "100",
		},
	}
}

// Validate checks that the provided parameters are valid.
func (a *ComposeAction) Validate(params map[string]interface{}) error {
	op, _ := params["operation"].(string)
	if op == "" {
		return fmt.Errorf("operation is required")
	}

	valid := false
	for _, validOp := range composeOperations {
		if op == validOp {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unknown operation: %s, supported: %v", op, composeOperations)
	}

	machineID, _ := params["machine_id"].(string)
	if machineID == "" {
		return fmt.Errorf("machine_id is required")
	}

	projectName, _ := params["project_name"].(string)
	if projectName == "" {
		return fmt.Errorf("project_name is required")
	}

	return nil
}

// Destructive returns true if the compose operation requires confirmation.
func (a *ComposeAction) Destructive(params map[string]interface{}) bool {
	op, _ := params["operation"].(string)
	switch op {
	case "compose_up", "compose_down", "compose_start", "compose_stop", "compose_restart", "compose_build":
		return true
	default:
		return false
	}
}

// ConfirmationMessage returns a human-readable message about what the action will do.
func (a *ComposeAction) ConfirmationMessage(params map[string]interface{}) string {
	op, _ := params["operation"].(string)
	projectName, _ := params["project_name"].(string)
	if projectName == "" {
		projectName = "unknown"
	}
	switch op {
	case "compose_up":
		return fmt.Sprintf("This will start compose project '%s'.", projectName)
	case "compose_down":
		return fmt.Sprintf("This will stop and remove compose project '%s'.", projectName)
	case "compose_start":
		return fmt.Sprintf("This will start compose project '%s'.", projectName)
	case "compose_stop":
		return fmt.Sprintf("This will stop compose project '%s'.", projectName)
	case "compose_restart":
		return fmt.Sprintf("This will restart compose project '%s'.", projectName)
	case "compose_build":
		return fmt.Sprintf("This will build compose project '%s'.", projectName)
	default:
		return fmt.Sprintf("Execute %s on compose project '%s'.", op, projectName)
	}
}

// Execute runs the requested compose operation.
func (a *ComposeAction) Execute(ctx context.Context, params map[string]interface{}) (*ActionResult, error) {
	op, _ := params["operation"].(string)
	machineID, _ := params["machine_id"].(string)
	projectName, _ := params["project_name"].(string)

	switch op {
	case "list_compose":
		files, err := a.uc.List(ctx, machineID)
		if err != nil {
			return nil, fmt.Errorf("ComposeAction.Execute - list_compose: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: "Compose projects listed successfully",
			Data: map[string]interface{}{
				"projects": files,
			},
		}, nil

	case "compose_ps":
		services, err := a.uc.Ps(ctx, machineID, projectName)
		if err != nil {
			return nil, fmt.Errorf("ComposeAction.Execute - compose_ps: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: "Compose project status retrieved successfully",
			Data: map[string]interface{}{
				"services": services,
			},
		}, nil

	case "compose_logs":
		tail, _ := params["tail"].(string)
		if tail == "" {
			tail = "100"
		}
		logs, err := a.uc.Logs(ctx, machineID, projectName)
		if err != nil {
			return nil, fmt.Errorf("ComposeAction.Execute - compose_logs: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: "Compose project logs retrieved successfully",
			Data: map[string]interface{}{
				"logs": logs,
				"tail": tail,
			},
		}, nil

	case "compose_up":
		yamlContent, _ := params["yaml_content"].(string)
		filePath, _ := params["file_path"].(string)

		// If yaml_content is provided, create or update the compose file
		if yamlContent != "" {
			_, err := a.uc.Get(ctx, machineID, projectName)
			if err != nil {
				// Project doesn't exist, create it
				_, err = a.uc.Create(ctx, machineID, projectName, yamlContent, filePath)
				if err != nil {
					return nil, fmt.Errorf("ComposeAction.Execute - compose_up: %w", err)
				}
			} else {
				// Project exists, update it
				err = a.uc.Update(ctx, machineID, projectName, yamlContent)
				if err != nil {
					return nil, fmt.Errorf("ComposeAction.Execute - compose_up: %w", err)
				}
			}
		}

		if err := a.uc.Up(ctx, machineID, projectName); err != nil {
			return nil, fmt.Errorf("ComposeAction.Execute - compose_up: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Compose project '%s' started successfully", projectName),
		}, nil

	case "compose_down":
		volumes, _ := params["volumes"].(bool)
		if err := a.uc.Down(ctx, machineID, projectName, volumes); err != nil {
			return nil, fmt.Errorf("ComposeAction.Execute - compose_down: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Compose project '%s' stopped successfully", projectName),
		}, nil

	case "compose_start":
		if err := a.uc.Start(ctx, machineID, projectName); err != nil {
			return nil, fmt.Errorf("ComposeAction.Execute - compose_start: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Compose project '%s' started successfully", projectName),
		}, nil

	case "compose_stop":
		if err := a.uc.Stop(ctx, machineID, projectName); err != nil {
			return nil, fmt.Errorf("ComposeAction.Execute - compose_stop: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Compose project '%s' stopped successfully", projectName),
		}, nil

	case "compose_restart":
		if err := a.uc.Restart(ctx, machineID, projectName); err != nil {
			return nil, fmt.Errorf("ComposeAction.Execute - compose_restart: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Compose project '%s' restarted successfully", projectName),
		}, nil

	case "compose_build":
		if err := a.uc.Build(ctx, machineID, projectName); err != nil {
			return nil, fmt.Errorf("ComposeAction.Execute - compose_build: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Compose project '%s' built successfully", projectName),
		}, nil

	default:
		return nil, fmt.Errorf("unknown compose operation: %s", op)
	}
}
