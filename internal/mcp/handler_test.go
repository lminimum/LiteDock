package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/lminimum/LiteDock/internal/action"
	"github.com/lminimum/LiteDock/internal/usecase/assistant"
	"github.com/stretchr/testify/require"
)

type mockAction struct {
	name          string
	description   string
	params        []action.ParamDef
	destructiveFn func(map[string]interface{}) bool
	executeFn     func(context.Context, map[string]interface{}) (*action.ActionResult, error)
}

func (m *mockAction) Name() string {
	return m.name
}

func (m *mockAction) Description() string {
	return m.description
}

func (m *mockAction) Params() []action.ParamDef {
	return m.params
}

func (m *mockAction) Validate(params map[string]interface{}) error {
	return nil
}

func (m *mockAction) Destructive(params map[string]interface{}) bool {
	if m.destructiveFn != nil {
		return m.destructiveFn(params)
	}
	return false
}

func (m *mockAction) ConfirmationMessage(params map[string]interface{}) string {
	return "confirm"
}

func (m *mockAction) Execute(ctx context.Context, params map[string]interface{}) (*action.ActionResult, error) {
	if m.executeFn != nil {
		return m.executeFn(ctx, params)
	}
	return &action.ActionResult{Success: true, Message: "executed"}, nil
}

type mockActionRegistry struct {
	actions map[string]action.Action
}

func (m *mockActionRegistry) List() []action.Action {
	list := make([]action.Action, 0, len(m.actions))
	for _, a := range m.actions {
		list = append(list, a)
	}
	return list
}

func (m *mockActionRegistry) Get(name string) (action.Action, bool) {
	a, ok := m.actions[name]
	return a, ok
}

func (m *mockActionRegistry) ExecuteWithConfirmation(ctx context.Context, name string, params map[string]interface{}, token string) (*action.ActionResult, error) {
	a, ok := m.Get(name)
	if !ok {
		return nil, errors.New("unknown action")
	}
	return a.Execute(ctx, params)
}

func (m *mockActionRegistry) ExecuteConfirmed(ctx context.Context, name string, params map[string]interface{}) (*action.ActionResult, error) {
	a, ok := m.Get(name)
	if !ok {
		return nil, errors.New("unknown action")
	}
	return a.Execute(ctx, params)
}

type mockTokenService struct {
	validateFn func(tokenStr string, params assistant.ActionConfirmationToken) error
}

func (m *mockTokenService) ParamsHash(params map[string]string) string {
	return "hash"
}

func (m *mockTokenService) Validate(tokenStr string, params assistant.ActionConfirmationToken) error {
	if m.validateFn != nil {
		return m.validateFn(tokenStr, params)
	}
	return nil
}

type mockAuditLogger struct {
	infos []string
}

func (m *mockAuditLogger) Info(format string, v ...interface{}) {
	m.infos = append(m.infos, format)
}

func (m *mockAuditLogger) Error(err error, format string, v ...interface{}) {}

func TestToolsList(t *testing.T) {
	reg := &mockActionRegistry{
		actions: map[string]action.Action{
			"test_action": &mockAction{
				name:        "test_action",
				description: "A test action",
				params: []action.ParamDef{
					{
						Name:        "param1",
						Type:        "string",
						Required:    true,
						Description: "A param",
					},
				},
			},
		},
	}

	h := NewHandler(reg, nil, nil)

	req := &MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/list",
		ID:      1,
	}

	resp := h.Handle(context.Background(), req)
	require.Nil(t, resp.Error)

	result, ok := resp.Result.(*ToolsListResult)
	require.True(t, ok)
	require.Len(t, result.Tools, 1)
	require.Equal(t, "test_action", result.Tools[0].Name)
	require.Equal(t, "A test action", result.Tools[0].Description)
}

func TestToolsCallSafe(t *testing.T) {
	reg := &mockActionRegistry{
		actions: map[string]action.Action{
			"safe_action": &mockAction{
				name:        "safe_action",
				description: "A safe action",
			},
		},
	}

	h := NewHandler(reg, nil, nil)

	argsBytes, _ := json.Marshal(ToolsCallParams{
		Name: "safe_action",
		Arguments: map[string]interface{}{
			"arg1": "val1",
		},
	})

	req := &MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  argsBytes,
		ID:      1,
	}

	resp := h.Handle(context.Background(), req)
	require.Nil(t, resp.Error)

	result, ok := resp.Result.(*ToolsCallResult)
	require.True(t, ok)
	require.Len(t, result.Content, 1)
	require.Equal(t, "text", result.Content[0].Type)
	require.Contains(t, result.Content[0].Text, "executed")
}

func TestToolsCallDangerousNoToken(t *testing.T) {
	reg := &mockActionRegistry{
		actions: map[string]action.Action{
			"dangerous_action": &mockAction{
				name:        "dangerous_action",
				description: "A dangerous action",
				destructiveFn: func(m map[string]interface{}) bool {
					return true
				},
			},
		},
	}

	logger := &mockAuditLogger{}
	h := NewHandler(reg, nil, logger)

	argsBytes, _ := json.Marshal(ToolsCallParams{
		Name: "dangerous_action",
		Arguments: map[string]interface{}{
			"arg1": "val1",
		},
	})

	req := &MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  argsBytes,
		ID:      1,
	}

	resp := h.Handle(context.Background(), req)
	require.NotNil(t, resp.Error)
	require.Equal(t, -32000, resp.Error.Code)
	require.Equal(t, "confirmation_required", resp.Error.Message)
	require.Len(t, logger.infos, 1)
}

func TestToolsCallDangerousWithValidToken(t *testing.T) {
	reg := &mockActionRegistry{
		actions: map[string]action.Action{
			"dangerous_action": &mockAction{
				name:        "dangerous_action",
				description: "A dangerous action",
				destructiveFn: func(m map[string]interface{}) bool {
					return true
				},
			},
		},
	}

	ts := &mockTokenService{
		validateFn: func(tokenStr string, params assistant.ActionConfirmationToken) error {
			if tokenStr == "valid_token" {
				return nil
			}
			return errors.New("invalid token")
		},
	}

	logger := &mockAuditLogger{}
	h := NewHandler(reg, ts, logger)

	argsBytes, _ := json.Marshal(ToolsCallParams{
		Name: "dangerous_action",
		Arguments: map[string]interface{}{
			"confirmation_token": "valid_token",
			"arg1":               "val1",
		},
	})

	req := &MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  argsBytes,
		ID:      1,
	}

	resp := h.Handle(context.Background(), req)
	require.Nil(t, resp.Error)

	result, ok := resp.Result.(*ToolsCallResult)
	require.True(t, ok)
	require.Len(t, result.Content, 1)
	require.Contains(t, result.Content[0].Text, "executed")
	require.Len(t, logger.infos, 1)
}

func TestToolsCallDangerousWithInvalidToken(t *testing.T) {
	reg := &mockActionRegistry{
		actions: map[string]action.Action{
			"dangerous_action": &mockAction{
				name:        "dangerous_action",
				description: "A dangerous action",
				destructiveFn: func(m map[string]interface{}) bool {
					return true
				},
			},
		},
	}

	ts := &mockTokenService{
		validateFn: func(tokenStr string, params assistant.ActionConfirmationToken) error {
			return assistant.ErrTokenMismatch
		},
	}

	logger := &mockAuditLogger{}
	h := NewHandler(reg, ts, logger)

	argsBytes, _ := json.Marshal(ToolsCallParams{
		Name: "dangerous_action",
		Arguments: map[string]interface{}{
			"confirmation_token": "invalid_token",
			"arg1":               "val1",
		},
	})

	req := &MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  argsBytes,
		ID:      1,
	}

	resp := h.Handle(context.Background(), req)
	require.NotNil(t, resp.Error)
	require.Equal(t, -32001, resp.Error.Code)
	require.Contains(t, resp.Error.Message, "invalid or expired")
	require.Len(t, logger.infos, 1)
}

func TestToolsCallUnknownTool(t *testing.T) {
	reg := &mockActionRegistry{
		actions: map[string]action.Action{},
	}

	h := NewHandler(reg, nil, nil)

	argsBytes, _ := json.Marshal(ToolsCallParams{
		Name: "nonexistent_action",
		Arguments: map[string]interface{}{
			"arg1": "val1",
		},
	})

	req := &MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  argsBytes,
		ID:      1,
	}

	resp := h.Handle(context.Background(), req)
	require.NotNil(t, resp.Error)
	require.Equal(t, -32602, resp.Error.Code)
	require.Contains(t, resp.Error.Message, "unknown action")
}

func TestToolsCallInvalidParamsJSON(t *testing.T) {
	reg := &mockActionRegistry{
		actions: map[string]action.Action{},
	}

	h := NewHandler(reg, nil, nil)

	req := &MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  json.RawMessage(`{invalid json}`),
		ID:      1,
	}

	resp := h.Handle(context.Background(), req)
	require.NotNil(t, resp.Error)
	require.Equal(t, -32602, resp.Error.Code)
	require.Contains(t, resp.Error.Message, "Invalid params")
}

func TestToolsCallDangerousWithExpiredToken(t *testing.T) {
	reg := &mockActionRegistry{
		actions: map[string]action.Action{
			"dangerous_action": &mockAction{
				name:        "dangerous_action",
				description: "A dangerous action",
				destructiveFn: func(m map[string]interface{}) bool {
					return true
				},
			},
		},
	}

	ts := &mockTokenService{
		validateFn: func(tokenStr string, params assistant.ActionConfirmationToken) error {
			return assistant.ErrTokenExpired
		},
	}

	logger := &mockAuditLogger{}
	h := NewHandler(reg, ts, logger)

	argsBytes, _ := json.Marshal(ToolsCallParams{
		Name: "dangerous_action",
		Arguments: map[string]interface{}{
			"confirmation_token": "expired_token",
			"arg1":               "val1",
		},
	})

	req := &MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  argsBytes,
		ID:      1,
	}

	resp := h.Handle(context.Background(), req)
	require.NotNil(t, resp.Error)
	require.Equal(t, -32001, resp.Error.Code)
	require.Contains(t, resp.Error.Message, "invalid or expired")
	require.Len(t, logger.infos, 1)
}
