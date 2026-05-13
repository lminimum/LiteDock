package action

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ActionRegistry is a central registry for all AI-callable operations.
type ActionRegistry struct {
	mu      sync.RWMutex
	actions map[string]Action
}

// NewActionRegistry creates a new empty registry.
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[string]Action),
	}
}

// Register adds an action to the registry.
func (r *ActionRegistry) Register(a Action) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := a.Name()
	if _, exists := r.actions[name]; exists {
		return fmt.Errorf("%w: %s", ErrActionAlreadyRegistered, name)
	}
	r.actions[name] = a
	return nil
}

// Get retrieves an action by name.
func (r *ActionRegistry) Get(name string) (Action, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, ok := r.actions[name]
	return a, ok
}

// List returns all registered actions.
func (r *ActionRegistry) List() []Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Action, 0, len(r.actions))
	for _, a := range r.actions {
		result = append(result, a)
	}
	return result
}

// Execute runs an action by name with the given params.
// If the action is destructive and confirmed is false, it returns ErrDestructiveAction.
func (r *ActionRegistry) Execute(ctx context.Context, name string, params map[string]interface{}) (*ActionResult, error) {
	a, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownAction, name)
	}

	if err := a.Validate(params); err != nil {
		return nil, fmt.Errorf("%w for %s: %w", ErrValidationFailed, name, err)
	}

	if a.Destructive(params) {
		return nil, fmt.Errorf("%w: %s", ErrDestructiveAction, name)
	}

	start := time.Now()
	result, err := a.Execute(ctx, params)
	if result != nil {
		result.Duration = time.Since(start).Round(time.Millisecond).String()
	}
	return result, err
}

// ExecuteConfirmed runs an action by name with the given params, bypassing the destructive check.
// This should only be called after the user has explicitly confirmed the action.
func (r *ActionRegistry) ExecuteConfirmed(ctx context.Context, name string, params map[string]interface{}) (*ActionResult, error) {
	a, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownAction, name)
	}

	if err := a.Validate(params); err != nil {
		return nil, fmt.Errorf("%w for %s: %w", ErrValidationFailed, name, err)
	}

	start := time.Now()
	result, err := a.Execute(ctx, params)
	if result != nil {
		result.Duration = time.Since(start).Round(time.Millisecond).String()
	}
	return result, err
}

// GenerateToolDefs generates OpenAI-compatible tool definitions from all registered actions.
func (r *ActionRegistry) GenerateToolDefs() []map[string]interface{} {
	actions := r.List()
	tools := make([]map[string]interface{}, 0, len(actions))

	for _, a := range actions {
		properties := make(map[string]interface{})
		required := make([]string, 0)

		for _, p := range a.Params() {
			properties[p.Name] = p.ToolDefForLLM()
			if p.Required {
				required = append(required, p.Name)
			}
		}

		tools = append(tools, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        a.Name(),
				"description": a.Description(),
				"parameters": map[string]interface{}{
					"type":       "object",
					"properties": properties,
					"required":   required,
				},
			},
		})
	}

	return tools
}
