package entity

import (
	"errors"
	"testing"
	"time"
)

func TestAIAuditEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   AIAuditEvent
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid event with all fields",
			event: AIAuditEvent{
				Timestamp:  time.Now().UTC(),
				UserID:     "user-123",
				SessionID:  "session-456",
				Source:     AuditSourceREST,
				Action:     "container.delete",
				ParamsHash: "abc123",
				RiskLevel:  RiskLevelDangerous,
				Result:     AuditResultSuccess,
				TokenValid: true,
			},
			wantErr: false,
		},
		{
			name: "valid event with ws source",
			event: AIAuditEvent{
				Timestamp:  time.Now().UTC(),
				UserID:     "user-123",
				SessionID:  "session-456",
				Source:     AuditSourceWS,
				Action:     "container.list",
				ParamsHash: "",
				RiskLevel:  RiskLevelRead,
				Result:     AuditResultSuccess,
				TokenValid: true,
			},
			wantErr: false,
		},
		{
			name: "valid event with mcp source",
			event: AIAuditEvent{
				Timestamp:  time.Now().UTC(),
				UserID:     "user-123",
				SessionID:  "session-456",
				Source:     AuditSourceMCP,
				Action:     "image.pull",
				ParamsHash: "def456",
				RiskLevel:  RiskLevelModify,
				Result:     AuditResultRejected,
				TokenValid: false,
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			event: AIAuditEvent{
				Timestamp:  time.Now().UTC(),
				UserID:     "",
				SessionID:  "session-456",
				Source:     AuditSourceREST,
				Action:     "container.delete",
				ParamsHash: "abc123",
				RiskLevel:  RiskLevelDangerous,
				Result:     AuditResultSuccess,
				TokenValid: true,
			},
			wantErr: true,
			errMsg:  "user_id is required",
		},
		{
			name: "missing action",
			event: AIAuditEvent{
				Timestamp:  time.Now().UTC(),
				UserID:     "user-123",
				SessionID:  "session-456",
				Source:     AuditSourceREST,
				Action:     "",
				ParamsHash: "abc123",
				RiskLevel:  RiskLevelDangerous,
				Result:     AuditResultSuccess,
				TokenValid: true,
			},
			wantErr: true,
			errMsg:  "action is required",
		},
		{
			name: "invalid source",
			event: AIAuditEvent{
				Timestamp:  time.Now().UTC(),
				UserID:     "user-123",
				SessionID:  "session-456",
				Source:     "invalid",
				Action:     "container.delete",
				ParamsHash: "abc123",
				RiskLevel:  RiskLevelDangerous,
				Result:     AuditResultSuccess,
				TokenValid: true,
			},
			wantErr: true,
			errMsg:  "invalid source: invalid",
		},
		{
			name: "invalid result",
			event: AIAuditEvent{
				Timestamp:  time.Now().UTC(),
				UserID:     "user-123",
				SessionID:  "session-456",
				Source:     AuditSourceREST,
				Action:     "container.delete",
				ParamsHash: "abc123",
				RiskLevel:  RiskLevelDangerous,
				Result:     "invalid",
				TokenValid: true,
			},
			wantErr: true,
			errMsg:  "invalid result: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestParamsHash(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
	}{
		{
			name:   "nil params",
			params: nil,
		},
		{
			name:   "empty params",
			params: map[string]string{},
		},
		{
			name:   "single param",
			params: map[string]string{"key": "value"},
		},
		{
			name:   "multiple params sorted",
			params: map[string]string{"zebra": "z", "alpha": "a", "beta": "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParamsHash(tt.params)
			if tt.params == nil || len(tt.params) == 0 {
				if got != "" {
					t.Errorf("ParamsHash() = %v, want empty string", got)
				}
				return
			}
			if got == "" {
				t.Errorf("ParamsHash() = empty, want non-empty hash")
			}
		})
	}
}

func TestParamsHash_Deterministic(t *testing.T) {
	params := map[string]string{"action": "delete", "container_id": "abc123", "force": "true"}
	hash1 := ParamsHash(params)
	hash2 := ParamsHash(params)
	if hash1 != hash2 {
		t.Errorf("ParamsHash() not deterministic: got %v and %v", hash1, hash2)
	}
}

func TestParamsHash_OrderIndependent(t *testing.T) {
	params1 := map[string]string{"a": "1", "b": "2", "c": "3"}
	params2 := map[string]string{"c": "3", "a": "1", "b": "2"}
	hash1 := ParamsHash(params1)
	hash2 := ParamsHash(params2)
	if hash1 != hash2 {
		t.Errorf("ParamsHash() order dependent: got %v and %v", hash1, hash2)
	}
}

func TestNewAIAuditEvent(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		sessionID string
		source    AuditSource
		action    string
		params    map[string]string
		riskLevel RiskLevel
		result    AuditResult
		err       error
		wantHash  bool
		wantErr   bool
	}{
		{
			name:      "success case",
			userID:    "user-123",
			sessionID: "session-456",
			source:    AuditSourceREST,
			action:    "container.delete",
			params:    map[string]string{"id": "abc"},
			riskLevel: RiskLevelDangerous,
			result:    AuditResultSuccess,
			err:       nil,
			wantHash:  true,
			wantErr:   false,
		},
		{
			name:      "error case sets error_msg",
			userID:    "user-123",
			sessionID: "session-456",
			source:    AuditSourceWS,
			action:    "container.start",
			params:    nil,
			riskLevel: RiskLevelModify,
			result:    AuditResultFailed,
			err:       errors.New("connection refused"),
			wantHash:  true,
			wantErr:   false,
		},
		{
			name:      "missing user_id fails validation",
			userID:    "",
			sessionID: "session-456",
			source:    AuditSourceREST,
			action:    "container.delete",
			params:    map[string]string{"id": "abc"},
			riskLevel: RiskLevelDangerous,
			result:    AuditResultSuccess,
			err:       nil,
			wantHash:  true,
			wantErr:   true,
		},
		{
			name:      "missing action fails validation",
			userID:    "user-123",
			sessionID: "session-456",
			source:    AuditSourceMCP,
			action:    "",
			params:    map[string]string{"id": "abc"},
			riskLevel: RiskLevelRead,
			result:    AuditResultSuccess,
			err:       nil,
			wantHash:  true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewAIAuditEvent(
				tt.userID, tt.sessionID, tt.source, tt.action,
				tt.params, tt.riskLevel, tt.result, tt.err,
			)
			if event.Timestamp.IsZero() {
				t.Errorf("NewAIAuditEvent() Timestamp not set")
			}
			if tt.wantHash && event.ParamsHash == "" && len(tt.params) > 0 {
				t.Errorf("NewAIAuditEvent() ParamsHash not computed")
			}
			if !tt.wantHash && event.ParamsHash != "" {
				t.Errorf("NewAIAuditEvent() ParamsHash should be empty, got %v", event.ParamsHash)
			}
			if tt.err != nil && event.ErrorMsg == "" {
				t.Errorf("NewAIAuditEvent() ErrorMsg not set")
			}
			if tt.err == nil && event.ErrorMsg != "" {
				t.Errorf("NewAIAuditEvent() ErrorMsg should be empty, got %v", event.ErrorMsg)
			}
			err := event.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
