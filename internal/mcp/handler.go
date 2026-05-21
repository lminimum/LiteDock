package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lminimum/LiteDock/internal/action"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/usecase/assistant"
)

type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *MCPError) Error() string {
	return e.Message
}

type ToolsListParams struct{}

type ToolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type ToolsCallResult struct {
	Content []ContentBlock `json:"content"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type Handler struct {
	actionRegistry interface {
		List() []action.Action
		Get(name string) (action.Action, bool)
		ExecuteWithConfirmation(ctx context.Context, name string, params map[string]interface{}, token string) (*action.ActionResult, error)
		ExecuteConfirmed(ctx context.Context, name string, params map[string]interface{}) (*action.ActionResult, error)
	}
	tokenService interface {
		ParamsHash(params map[string]string) string
		Validate(tokenStr string, params assistant.ActionConfirmationToken) error
	}
	auditLogger interface {
		Info(format string, v ...interface{})
		Error(err error, format string, v ...interface{})
	}
}

func NewHandler(registry any, tokenService any, auditLogger any) *Handler {
	var reg interface {
		List() []action.Action
		Get(name string) (action.Action, bool)
		ExecuteWithConfirmation(ctx context.Context, name string, params map[string]interface{}, token string) (*action.ActionResult, error)
		ExecuteConfirmed(ctx context.Context, name string, params map[string]interface{}) (*action.ActionResult, error)
	}
	if r, ok := registry.(interface {
		List() []action.Action
		Get(name string) (action.Action, bool)
		ExecuteWithConfirmation(ctx context.Context, name string, params map[string]interface{}, token string) (*action.ActionResult, error)
		ExecuteConfirmed(ctx context.Context, name string, params map[string]interface{}) (*action.ActionResult, error)
	}); ok {
		reg = r
	}

	var ts interface {
		ParamsHash(params map[string]string) string
		Validate(tokenStr string, params assistant.ActionConfirmationToken) error
	}
	if t, ok := tokenService.(interface {
		ParamsHash(params map[string]string) string
		Validate(tokenStr string, params assistant.ActionConfirmationToken) error
	}); ok {
		ts = t
	}

	var al interface {
		Info(format string, v ...interface{})
		Error(err error, format string, v ...interface{})
	}
	if a, ok := auditLogger.(interface {
		Info(format string, v ...interface{})
		Error(err error, format string, v ...interface{})
	}); ok {
		al = a
	}

	return &Handler{
		actionRegistry: reg,
		tokenService:   ts,
		auditLogger:    al,
	}
}

func (h *Handler) Handle(ctx context.Context, req *MCPRequest) *MCPResponse {
	resp := &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "tools/list":
		result, err := h.handleToolsList(ctx)
		if err != nil {
			resp.Error = &MCPError{Code: -32603, Message: err.Error()}
		} else {
			resp.Result = result
		}
	case "tools/call":
		var params ToolsCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &MCPError{Code: -32602, Message: "Invalid params: " + err.Error()}
			return resp
		}
		result, err := h.handleToolsCall(ctx, &params)
		if err != nil {
			var mcpErr *MCPError
			if errors.As(err, &mcpErr) {
				resp.Error = mcpErr
			} else {
				resp.Error = &MCPError{Code: -32603, Message: err.Error()}
			}
		} else {
			resp.Result = result
		}
	default:
		resp.Error = &MCPError{Code: -32601, Message: "Method not found: " + req.Method}
	}

	return resp
}

func (h *Handler) handleToolsList(ctx context.Context) (*ToolsListResult, error) {
	actions := h.actionRegistry.List()
	tools := make([]Tool, 0, len(actions))

	for _, a := range actions {
		properties := make(map[string]interface{})
		required := make([]string, 0)

		for _, p := range a.Params() {
			properties[p.Name] = p.ToolDefForLLM()
			if p.Required {
				required = append(required, p.Name)
			}
		}

		inputSchema := map[string]interface{}{
			"type":       "object",
			"properties": properties,
			"required":   required,
		}

		tools = append(tools, Tool{
			Name:        a.Name(),
			Description: a.Description(),
			InputSchema: inputSchema,
		})
	}

	return &ToolsListResult{Tools: tools}, nil
}

func (h *Handler) handleToolsCall(ctx context.Context, params *ToolsCallParams) (*ToolsCallResult, error) {
	act, found := h.actionRegistry.Get(params.Name)
	if !found {
		return nil, &MCPError{Code: -32602, Message: "unknown action: " + params.Name}
	}

	userID := "anonymous"
	if uid, ok := ctx.Value(action.CtxKeyUserID).(string); ok {
		userID = uid
	}
	sessionID := ""
	if sid, ok := ctx.Value(action.CtxKeySessionID).(string); ok {
		sessionID = sid
	}

	strParams := make(map[string]string)
	for k, v := range params.Arguments {
		if k != "confirmation_token" {
			strParams[k] = fmt.Sprintf("%v", v)
		}
	}

	riskLevel := entity.RiskLevelRead
	if act.Destructive(params.Arguments) {
		riskLevel = entity.RiskLevelDangerous
	}

	var confirmationToken string
	if tok, ok := params.Arguments["confirmation_token"].(string); ok {
		confirmationToken = tok
	}

	if riskLevel == entity.RiskLevelDangerous && confirmationToken == "" {
		h.emitAudit(userID, sessionID, params.Name, strParams, riskLevel, entity.AuditResultRejected, action.ErrConfirmationRequired, false, false)
		return nil, &MCPError{Code: -32000, Message: "confirmation_required"}
	}

	var result *action.ActionResult
	var err error

	if riskLevel == entity.RiskLevelDangerous {
		if h.tokenService == nil {
			h.emitAudit(userID, sessionID, params.Name, strParams, riskLevel, entity.AuditResultFailed, fmt.Errorf("token service not configured"), false, false)
			return nil, &MCPError{Code: -32603, Message: "token service not configured"}
		}

		hash := h.tokenService.ParamsHash(strParams)
		validationParams := assistant.ActionConfirmationToken{
			UserID:     userID,
			SessionID:  sessionID,
			Action:     params.Name,
			ParamsHash: hash,
			RiskLevel:  string(entity.RiskLevelDangerous),
		}

		if vErr := h.tokenService.Validate(confirmationToken, validationParams); vErr != nil {
			tokenExpired := errors.Is(vErr, assistant.ErrTokenExpired)
			h.emitAudit(userID, sessionID, params.Name, strParams, riskLevel, entity.AuditResultRejected, vErr, false, tokenExpired)
			return nil, &MCPError{Code: -32001, Message: "confirmation token is invalid or expired"}
		}

		result, err = h.actionRegistry.ExecuteConfirmed(ctx, params.Name, params.Arguments)
	} else {
		result, err = h.actionRegistry.ExecuteWithConfirmation(ctx, params.Name, params.Arguments, "")
	}

	if err != nil {
		h.emitAudit(userID, sessionID, params.Name, strParams, riskLevel, entity.AuditResultFailed, err, true, false)
		return nil, &MCPError{Code: -32603, Message: "Failed to execute action: " + err.Error()}
	}

	h.emitAudit(userID, sessionID, params.Name, strParams, riskLevel, entity.AuditResultSuccess, nil, true, false)

	dataBytes, _ := json.Marshal(result)
	return &ToolsCallResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: string(dataBytes),
			},
		},
	}, nil
}

func (h *Handler) emitAudit(userID, sessionID, actionName string, params map[string]string, riskLevel entity.RiskLevel, result entity.AuditResult, execErr error, tokenValid, tokenExpired bool) {
	if h.auditLogger == nil {
		return
	}
	event := entity.NewAIAuditEvent(userID, sessionID, entity.AuditSourceMCP, actionName, params, riskLevel, result, execErr)
	event.TokenValid = tokenValid
	event.TokenExpired = tokenExpired
	data, _ := json.Marshal(event)
	h.auditLogger.Info("audit: %s", string(data))
}
