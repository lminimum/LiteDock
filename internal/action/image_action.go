package action

import (
	"context"
	"fmt"

	dockerImage "github.com/docker/docker/api/types/image"
	"github.com/lminimum/LiteDock/internal/entity"
)

// imageUseCase defines the image use case methods needed by ImageAction.
type imageUseCase interface {
	List(ctx context.Context, machineID string) ([]entity.Image, error)
	Prune(ctx context.Context, machineID string) (*dockerImage.PruneReport, error)
}

// imageOperations lists the supported image sub-actions.
var imageOperations = []string{
	"list_images",
	"prune_images",
}

// ImageAction exposes image management operations as AI-callable actions.
type ImageAction struct {
	uc imageUseCase
}

// NewImageAction creates a new ImageAction wrapping the given use case.
func NewImageAction(uc imageUseCase) *ImageAction {
	return &ImageAction{uc: uc}
}

// Name returns the action name.
func (a *ImageAction) Name() string {
	return "image"
}

// Description returns a human-readable description of the action.
func (a *ImageAction) Description() string {
	return "Manage Docker images: list images and prune unused images"
}

// Params returns the parameter definitions for the image action.
func (a *ImageAction) Params() []ParamDef {
	return []ParamDef{
		{
			Name:        "operation",
			Type:        "string",
			Required:    true,
			Description: "The image operation to perform: list_images, prune_images",
		},
		{
			Name:        "machine_id",
			Type:        "string",
			Required:    true,
			Description: "The ID of the target machine (use 'local' for local Docker)",
		},
	}
}

// Validate checks that the provided parameters are valid.
func (a *ImageAction) Validate(params map[string]interface{}) error {
	op, _ := params["operation"].(string)
	if op == "" {
		return fmt.Errorf("operation is required")
	}

	valid := false
	for _, validOp := range imageOperations {
		if op == validOp {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unknown operation: %s, supported: %v", op, imageOperations)
	}

	machineID, _ := params["machine_id"].(string)
	if machineID == "" {
		return fmt.Errorf("machine_id is required")
	}

	return nil
}

// Destructive returns true if the image operation is destructive (e.g., prune_images).
func (a *ImageAction) Destructive(params map[string]interface{}) bool {
	op, _ := params["operation"].(string)
	return op == "prune_images"
}

// ConfirmationMessage returns a human-readable message about what the action will do.
func (a *ImageAction) ConfirmationMessage(params map[string]interface{}) string {
	op, _ := params["operation"].(string)
	switch op {
	case "prune_images":
		return "This will permanently remove all unused Docker images. This action cannot be undone."
	default:
		return fmt.Sprintf("Execute %s on images.", op)
	}
}

// Execute runs the requested image operation.
func (a *ImageAction) Execute(ctx context.Context, params map[string]interface{}) (*ActionResult, error) {
	op, _ := params["operation"].(string)
	machineID, _ := params["machine_id"].(string)

	switch op {
	case "list_images":
		images, err := a.uc.List(ctx, machineID)
		if err != nil {
			return nil, fmt.Errorf("ImageAction.Execute - list_images: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: "Images listed successfully",
			Data: map[string]interface{}{
				"images": images,
				"count":  len(images),
			},
		}, nil

	case "prune_images":
		report, err := a.uc.Prune(ctx, machineID)
		if err != nil {
			return nil, fmt.Errorf("ImageAction.Execute - prune_images: %w", err)
		}
		return &ActionResult{
			Success: true,
			Message: "Unused images pruned successfully",
			Data: map[string]interface{}{
				"space_reclaimed": report.SpaceReclaimed,
				"images_deleted":  len(report.ImagesDeleted),
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown image operation: %s", op)
	}
}
