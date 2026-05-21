package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"
)

type AuditResult string

const (
	AuditResultSuccess  AuditResult = "success"
	AuditResultRejected AuditResult = "rejected"
	AuditResultFailed   AuditResult = "failed"
)

type AuditSource string

const (
	AuditSourceREST AuditSource = "rest"
	AuditSourceWS   AuditSource = "ws"
	AuditSourceMCP  AuditSource = "mcp"
)

type AIAuditEvent struct {
	Timestamp    time.Time   `json:"timestamp"`
	UserID       string      `json:"user_id"`
	SessionID    string      `json:"session_id"`
	Source       AuditSource `json:"source"`
	Action       string      `json:"action"`
	ParamsHash   string      `json:"params_hash"`
	RiskLevel    RiskLevel   `json:"risk_level"`
	Result       AuditResult `json:"result"`
	ErrorMsg     string      `json:"error_msg,omitempty"`
	TokenValid   bool        `json:"token_valid"`
	TokenExpired bool        `json:"token_expired"`
}

func (e *AIAuditEvent) Validate() error {
	if e.UserID == "" {
		return errors.New("user_id is required")
	}
	if e.Action == "" {
		return errors.New("action is required")
	}
	if e.Source != AuditSourceREST && e.Source != AuditSourceWS && e.Source != AuditSourceMCP {
		return fmt.Errorf("invalid source: %s", e.Source)
	}
	if e.Result != AuditResultSuccess && e.Result != AuditResultRejected && e.Result != AuditResultFailed {
		return fmt.Errorf("invalid result: %s", e.Result)
	}
	return nil
}

func ParamsHash(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sorted := make(map[string]string, len(params))
	for _, k := range keys {
		sorted[k] = params[k]
	}
	data, _ := json.Marshal(sorted)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func NewAIAuditEvent(
	userID, sessionID string,
	source AuditSource,
	action string,
	params map[string]string,
	riskLevel RiskLevel,
	result AuditResult,
	err error,
) *AIAuditEvent {
	event := &AIAuditEvent{
		Timestamp:  time.Now().UTC(),
		UserID:     userID,
		SessionID:  sessionID,
		Source:     source,
		Action:     action,
		ParamsHash: ParamsHash(params),
		RiskLevel:  riskLevel,
		Result:     result,
		TokenValid: true,
	}
	if err != nil {
		event.ErrorMsg = err.Error()
	}
	return event
}
