package action

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// ParamDef defines a single parameter for an action.
type ParamDef struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
}

// ActionResult is returned after executing an action.
type ActionResult struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
	Duration string      `json:"duration,omitempty"`
}

// Action is implemented by use cases to expose AI-callable operations.
type Action interface {
	Name() string
	Description() string
	Params() []ParamDef
	Validate(params map[string]interface{}) error
	Execute(ctx context.Context, params map[string]interface{}) (*ActionResult, error)

	// Destructive returns true if this operation is destructive (data loss risk).
	Destructive(params map[string]interface{}) bool
	// ConfirmationMessage returns a human-readable description of what the action will do,
	// shown to the user before confirmation.
	ConfirmationMessage(params map[string]interface{}) string
}

// ToolDefForLLM generates an OpenAI tool definition property entry for a single parameter.
func (p ParamDef) ToolDefForLLM() map[string]interface{} {
	return map[string]interface{}{
		"type":        p.Type,
		"description": p.Description,
	}
}

// ErrActionAlreadyRegistered is returned when registering a duplicate action.
var ErrActionAlreadyRegistered = fmt.Errorf("action already registered")

// ErrUnknownAction is returned when executing an unregistered action.
var ErrUnknownAction = fmt.Errorf("unknown action")

// ErrValidationFailed is returned when action parameter validation fails.
var ErrValidationFailed = fmt.Errorf("validation failed")

// ErrDestructiveAction is returned when attempting to execute a destructive action without confirmation.
var ErrDestructiveAction = fmt.Errorf("destructive action requires confirmation")

// ErrMaliciousInput is returned when user input contains potentially malicious content.
var ErrMaliciousInput = fmt.Errorf("input contains potentially malicious content")

// ErrConfirmationRequired is returned when a destructive action is attempted without a confirmation token.
var ErrConfirmationRequired = fmt.Errorf("confirmation token required for this action")

// ErrInvalidToken is returned when the confirmation token is invalid or tampered with.
var ErrInvalidToken = fmt.Errorf("invalid confirmation token")

// ErrTokenExpired is returned when the confirmation token has expired.
var ErrTokenExpired = fmt.Errorf("confirmation token expired")

var (
	// Patterns that are rejected as potentially malicious.
	_maliciousPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>`),
		regexp.MustCompile(`(?i)<[^>]*on\w+\s*=`),
		regexp.MustCompile(`(?i)(\bDROP\b|\bALTER\b|\bTRUNCATE\b|\bDELETE\b)\s+(TABLE|DATABASE|SCHEMA|FROM)`),
		regexp.MustCompile(`(?i)\bUNION\s+SELECT\b`),
		regexp.MustCompile(`(?i)(\bEXEC\b|\bEXECUTE\b)\s*(\(|\s)`),
		regexp.MustCompile(`(?i)\bLOAD\s+(DATA|FILE)\b`),
		regexp.MustCompile(`(?i)\bINTO\s+(OUTFILE|DUMPFILE)\b`),
	}
)

// SanitizeInput checks user input for potentially malicious content.
// Returns ErrMaliciousInput if dangerous patterns are detected.
func SanitizeInput(input string) error {
	for _, pattern := range _maliciousPatterns {
		if pattern.MatchString(input) {
			return fmt.Errorf("%w: matched pattern %s", ErrMaliciousInput, pattern.String())
		}
	}
	// Check for prompt injection double-encoding tricks
	if strings.Contains(input, "\\\"}}") || strings.Contains(input, "\\\"}]") {
		return fmt.Errorf("%w: detected prompt injection encoding", ErrMaliciousInput)
	}
	return nil
}
