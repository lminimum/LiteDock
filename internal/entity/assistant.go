package entity

// ParseRequest represents a request to parse user intent from text.
type ParseRequest struct {
	Text string `json:"text"`
}

// ParseResponse represents the parsed intent and parameters.
type ParseResponse struct {
	Intent       string            `json:"intent"`
	Params       map[string]string `json:"params"`
	Action       string            `json:"action"`
	Description  string            `json:"description"`
	RequiresConfirmation bool      `json:"requires_confirmation,omitempty"`
	ConfirmationMessage string      `json:"confirmation_message,omitempty"`
	ActionName   string            `json:"action_name,omitempty"`
	ActionParams map[string]string `json:"action_params,omitempty"`
}

// DiagnoseRequest represents a request to diagnose container issues.
type DiagnoseRequest struct {
	ContainerID   string `json:"container_id"`
	ContainerName string `json:"container_name"`
	ExitCode      int    `json:"exit_code"`
}

// DiagnoseResponse represents the diagnosis result for a container.
type DiagnoseResponse struct {
	Diagnosis    string   `json:"diagnosis"`
	Cause        string   `json:"cause"`
	Remediation  []string `json:"remediation"`
	ExitCode     int      `json:"exit_code"`
}

// RecommendRequest represents a request to get configuration recommendations.
type RecommendRequest struct {
	Scenario string `json:"scenario"`
}

// RecommendResponse represents configuration recommendations for a scenario.
type RecommendResponse struct {
	Scenario string                `json:"scenario"`
	Configs  []ConfigRecommendation `json:"configs"`
}

type ConfigRecommendation struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Reason string `json:"reason"`
}

type RiskLevel string

const (
	RiskLevelRead      RiskLevel = "read"
	RiskLevelModify    RiskLevel = "modify"
	RiskLevelDangerous RiskLevel = "dangerous"
)

type ActionIntent struct {
	Action                       string            `json:"action"`
	Params                       map[string]string `json:"params"`
	RiskLevel                    RiskLevel         `json:"risk_level"`
	RequiresConfirmation         bool              `json:"requires_confirmation"`
	RequiresSecondConfirmation   bool              `json:"requires_second_confirmation,omitempty"`
	ConfirmationMessage          string            `json:"confirmation_message,omitempty"`
	ConfirmationToken            string            `json:"confirmation_token,omitempty"`
}