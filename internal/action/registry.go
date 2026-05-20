package action

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type ctxKey string

const (
	ctxKeyUserID    ctxKey = "user_id"
	ctxKeySessionID ctxKey = "session_id"
)

const (
	RiskLevelDangerous = "dangerous"
	RiskLevelSafe      = "safe"
)

const defaultTokenTTL = 2 * time.Minute

// TokenParams holds the parameters for token validation.
type TokenParams struct {
	UserID     string
	SessionID  string
	Action     string
	ParamsHash string
	RiskLevel  string
}

// TokenValidator abstracts the confirmation token validation logic.
type TokenValidator interface {
	Validate(tokenStr string, params TokenParams) error
}

// ActionRegistry is a central registry for all AI-callable operations.
type ActionRegistry struct {
	mu           sync.RWMutex
	actions      map[string]Action
	tokenService TokenValidator
}

// NewActionRegistry creates a new empty registry without token validation.
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[string]Action),
	}
}

// NewActionRegistryWithToken creates a new registry with confirmation token support.
func NewActionRegistryWithToken(ts TokenValidator) *ActionRegistry {
	return &ActionRegistry{
		actions:      make(map[string]Action),
		tokenService: ts,
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
// If the action is destructive, it returns ErrDestructiveAction.
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

// ExecuteReadOnly runs only non-destructive actions.
// Destructive actions are blocked with ErrConfirmationRequired.
func (r *ActionRegistry) ExecuteReadOnly(ctx context.Context, name string, params map[string]interface{}) (*ActionResult, error) {
	a, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownAction, name)
	}

	if err := a.Validate(params); err != nil {
		return nil, fmt.Errorf("%w for %s: %w", ErrValidationFailed, name, err)
	}

	if a.Destructive(params) {
		return nil, fmt.Errorf("%w: %s", ErrConfirmationRequired, name)
	}

	start := time.Now()
	result, err := a.Execute(ctx, params)
	if result != nil {
		result.Duration = time.Since(start).Round(time.Millisecond).String()
	}
	return result, err
}

// ExecuteWithConfirmation runs an action with a confirmation token.
// For non-destructive actions, the token is ignored.
// For destructive actions, a valid token matching action+params+user is required.
func (r *ActionRegistry) ExecuteWithConfirmation(ctx context.Context, name string, params map[string]interface{}, token string) (*ActionResult, error) {
	a, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownAction, name)
	}

	if err := a.Validate(params); err != nil {
		return nil, fmt.Errorf("%w for %s: %w", ErrValidationFailed, name, err)
	}

	isDestructive := a.Destructive(params)

	if isDestructive {
		if r.tokenService == nil {
			return nil, fmt.Errorf("%w: no token service configured", ErrConfirmationRequired)
		}

		if token == "" {
			return nil, fmt.Errorf("%w: %s", ErrConfirmationRequired, name)
		}

		userID := ctxValueOr(ctx, ctxKeyUserID, "anonymous")
		sessionID := ctxValueOr(ctx, ctxKeySessionID, "")

		validationParams := TokenParams{
			UserID:     userID,
			SessionID:  sessionID,
			Action:     name,
			ParamsHash: ComputeParamsHash(params),
			RiskLevel:  RiskLevelDangerous,
		}

		if err := r.tokenService.Validate(token, validationParams); err != nil {
			return nil, err
		}
	}

	start := time.Now()
	result, err := a.Execute(ctx, params)
	if result != nil {
		result.Duration = time.Since(start).Round(time.Millisecond).String()
	}
	return result, err
}

// Deprecated: ExecuteConfirmed runs an action bypassing the destructive check.
// Use ExecuteWithConfirmation instead, which validates a confirmation token.
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

func ctxValueOr(ctx context.Context, key ctxKey, fallback string) string {
	if v, ok := ctx.Value(key).(string); ok {
		return v
	}
	return fallback
}

// ComputeParamsHash produces a deterministic SHA-256 hex digest of the given params.
func ComputeParamsHash(params map[string]interface{}) string {
	if params == nil {
		hash := sha256.Sum256([]byte(""))
		return fmt.Sprintf("%x", hash)
	}

	strParams := make(map[string]string, len(params))
	for k, v := range params {
		strParams[k] = fmt.Sprintf("%v", v)
	}

	keys := make([]string, 0, len(strParams))
	for k := range strParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(strParams[k])
	}

	hash := sha256.Sum256([]byte(sb.String()))
	return fmt.Sprintf("%x", hash)
}
