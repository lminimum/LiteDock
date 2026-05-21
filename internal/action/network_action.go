package action

import (
	"context"
	"fmt"

	"github.com/lminimum/LiteDock/internal/entity"
)

// networkUseCase defines the network use case methods needed by NetworkAction.
type networkUseCase interface {
	ListNetworks(ctx context.Context, machineID string) ([]entity.Network, error)
	CreateNetwork(ctx context.Context, machineID, name, driver string) (*entity.Network, error)
	DeleteNetwork(ctx context.Context, machineID, networkName string) error
	InspectNetwork(ctx context.Context, machineID, networkName string) (*entity.Network, error)
}

// networkOperations lists the supported network sub-actions.
var networkOperations = []string{
	"list_networks",
	"inspect_network",
	"create_network",
	"delete_network",
}

// builtInNetworks lists networks that cannot be deleted.
var builtInNetworks = []string{
	"bridge",
	"host",
	"none",
}

// NetworkAction exposes network management operations as AI-callable actions.
type NetworkAction struct {
	uc networkUseCase
}

// NewNetworkAction creates a new NetworkAction wrapping the given use case.
func NewNetworkAction(uc networkUseCase) *NetworkAction {
	return &NetworkAction{uc: uc}
}

// Name returns the action name.
func (a *NetworkAction) Name() string {
	return "network"
}

// Description returns a human-readable description of the action.
func (a *NetworkAction) Description() string {
	return "Manage Docker networks: list, inspect, create, and delete networks"
}

// Params returns the parameter definitions for the network action.
func (a *NetworkAction) Params() []ParamDef {
	return []ParamDef{
		{
			Name:        "operation",
			Type:        "string",
			Required:    true,
			Description: "The network operation to perform: list_networks, inspect_network, create_network, delete_network",
		},
		{
			Name:        "machine_id",
			Type:        "string",
			Required:    true,
			Description: "The ID of the target machine (use 'local' for local Docker)",
		},
		{
			Name:        "network_id",
			Type:        "string",
			Required:    false,
			Description: "The network ID or name (required for inspect_network and delete_network)",
		},
		{
			Name:        "network_name",
			Type:        "string",
			Required:    false,
			Description: "The name for the new network (required for create_network)",
		},
		{
			Name:        "driver",
			Type:        "string",
			Required:    false,
			Description: "The network driver (optional for create_network, default: bridge)",
			Default:     "bridge",
		},
	}
}

// Validate checks that the provided parameters are valid.
func (a *NetworkAction) Validate(params map[string]interface{}) error {
	op, _ := params["operation"].(string)
	if op == "" {
		return fmt.Errorf("operation is required")
	}

	valid := false
	for _, validOp := range networkOperations {
		if op == validOp {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unknown operation: %s, supported: %v", op, networkOperations)
	}

	machineID, _ := params["machine_id"].(string)
	if machineID == "" {
		return fmt.Errorf("machine_id is required")
	}

	switch op {
	case "inspect_network", "delete_network":
		networkID, _ := params["network_id"].(string)
		if networkID == "" {
			return fmt.Errorf("network_id is required for %s", op)
		}
		if op == "delete_network" {
			for _, builtIn := range builtInNetworks {
				if networkID == builtIn {
					return fmt.Errorf("deletion of built-in network '%s' is not allowed", networkID)
				}
			}
		}
	case "create_network":
		networkName, _ := params["network_name"].(string)
		if networkName == "" {
			return fmt.Errorf("network_name is required for create_network")
		}
	}

	return nil
}

// Destructive returns true if the network operation requires confirmation.
func (a *NetworkAction) Destructive(params map[string]interface{}) bool {
	op, _ := params["operation"].(string)
	switch op {
	case "delete_network":
		return true
	default:
		return false
	}
}

// ConfirmationMessage returns a human-readable message about what the action will do.
func (a *NetworkAction) ConfirmationMessage(params map[string]interface{}) string {
	op, _ := params["operation"].(string)
	networkID, _ := params["network_id"].(string)
	if networkID == "" {
		networkID = "unknown"
	}
	switch op {
	case "delete_network":
		return fmt.Sprintf("This will delete network '%s'.", networkID)
	case "create_network":
		networkName, _ := params["network_name"].(string)
		driver, _ := params["driver"].(string)
		if driver == "" {
			driver = "bridge"
		}
		return fmt.Sprintf("This will create network '%s' with driver '%s'.", networkName, driver)
	default:
		return fmt.Sprintf("Execute %s on network '%s'.", op, networkID)
	}
}

// Execute runs the requested network operation.
func (a *NetworkAction) Execute(ctx context.Context, params map[string]interface{}) (*ActionResult, error) {
	op, _ := params["operation"].(string)
	machineID, _ := params["machine_id"].(string)

	switch op {
	case "list_networks":
		networks, err := a.uc.ListNetworks(ctx, machineID)
		if err != nil {
			return nil, fmt.Errorf("NetworkAction.Execute - list_networks: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Listed %d networks", len(networks)),
			Data: map[string]interface{}{
				"networks": networks,
			},
		}, nil

	case "inspect_network":
		networkID, _ := params["network_id"].(string)
		network, err := a.uc.InspectNetwork(ctx, machineID, networkID)
		if err != nil {
			return nil, fmt.Errorf("NetworkAction.Execute - inspect_network: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Network '%s' inspected successfully", networkID),
			Data: map[string]interface{}{
				"network": network,
			},
		}, nil

	case "create_network":
		networkName, _ := params["network_name"].(string)
		driver, _ := params["driver"].(string)
		if driver == "" {
			driver = "bridge"
		}
		network, err := a.uc.CreateNetwork(ctx, machineID, networkName, driver)
		if err != nil {
			return nil, fmt.Errorf("NetworkAction.Execute - create_network: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Network '%s' created successfully", networkName),
			Data: map[string]interface{}{
				"network": network,
			},
		}, nil

	case "delete_network":
		networkID, _ := params["network_id"].(string)
		if err := a.uc.DeleteNetwork(ctx, machineID, networkID); err != nil {
			return nil, fmt.Errorf("NetworkAction.Execute - delete_network: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: fmt.Sprintf("Network '%s' deleted successfully", networkID),
		}, nil

	default:
		return nil, fmt.Errorf("unknown network operation: %s", op)
	}
}
