package v1

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/lminimum/LiteDock/internal/action"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/usecase/assistant"
	assistant_util "github.com/lminimum/LiteDock/pkg/assistant/util"
	"github.com/lminimum/LiteDock/pkg/logger"
)

// AssistantHandler handles AI assistant requests (NL parsing, fault diagnosis, config recommendations).
type AssistantHandler struct {
	parser        *assistant.NLParserUseCase
	diagnosis     *assistant.FaultDiagnosisUseCase
	recommend     *assistant.ConfigRecommendUseCase
	settingsStore *AISettingsStore
	registry      *action.ActionRegistry
	rateLimiter   *assistant.RateLimiter
	tokenService  *assistant.TokenService
	l             logger.Interface
}

// NewAssistantRoutes registers assistant routes under /v1/assistant.
func NewAssistantRoutes(apiV1Group fiber.Router, parser *assistant.NLParserUseCase, diagnosis *assistant.FaultDiagnosisUseCase, recommend *assistant.ConfigRecommendUseCase, settingsStore *AISettingsStore, registry *action.ActionRegistry, rateLimiter *assistant.RateLimiter, l logger.Interface) {
	h := &AssistantHandler{
		parser:        parser,
		diagnosis:     diagnosis,
		recommend:     recommend,
		settingsStore: settingsStore,
		registry:      registry,
		rateLimiter:   rateLimiter,
		tokenService:  assistant.NewTokenService("", 0),
		l:             l,
	}

	assistantGroup := apiV1Group.Group("/assistant")
	assistantGroup.Post("/parse", h.Parse)
	assistantGroup.Post("/diagnose", h.Diagnose)
	assistantGroup.Post("/recommend", h.Recommend)
	assistantGroup.Post("/stream", h.Stream)
	assistantGroup.Get("/stream/ws", websocket.New(h.StreamWS, websocket.Config{Origins: []string{"*"}}))
	assistantGroup.Post("/execute", h.ExecuteAction)
}

// streamChunk is the SSE event payload sent to the client.
type streamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

// WSEventType represents the type of a WebSocket event envelope.
type WSEventType string

const (
	WSEventContent        WSEventType = "content"
	WSEventActionRequired WSEventType = "action_required"
	WSEventActionResult   WSEventType = "action_result"
	WSEventError          WSEventType = "error"
	WSEventDone           WSEventType = "done"
)

// WSPayloadContent is the payload for content events.
type WSPayloadContent struct {
	Content string `json:"content,omitempty"`
	Done    bool   `json:"done,omitempty"`
}

// WSPayloadActionRequired is the payload for action_required events.
type WSPayloadActionRequired struct {
	Action string            `json:"action,omitempty"`
	Params map[string]string `json:"params,omitempty"`
}

// WSPayloadError is the payload for error events.
type WSPayloadError struct {
	Message string `json:"message,omitempty"`
}

// WSEventEnvelope is the versioned envelope for WebSocket events.
type WSEventEnvelope struct {
	V       int             `json:"v"`
	Type    WSEventType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// sendWSEvent marshals and writes a WebSocket event to the given writer.
func sendWSEvent(w *bufio.Writer, eventType WSEventType, payload interface{}) {
	env := WSEventEnvelope{V: 1, Type: eventType}
	if payload != nil {
		payloadBytes, _ := json.Marshal(payload)
		env.Payload = payloadBytes
	}
	envBytes, _ := json.Marshal(env)
	fmt.Fprintf(w, "data: %s\n\n", envBytes)
	w.Flush()
}

// streamToolCallDeltaFunc is the function sub-object in a streaming delta.tool_calls chunk.
type streamToolCallDeltaFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// streamToolCallDelta is a single entry in the delta.tool_calls array of an OpenAI SSE chunk.
type streamToolCallDelta struct {
	Index    int                     `json:"index"`
	ID       string                  `json:"id,omitempty"`
	Type     string                  `json:"type,omitempty"`
	Function streamToolCallDeltaFunc `json:"function"`
}

// streamToolCall accumulates partial tool call data across streaming chunks.
type streamToolCall struct {
	id   string
	name string
	args string
}

func (h *AssistantHandler) executeToolCallAction(
	actionName string,
	params map[string]string,
	autonomous bool,
) (isDestructive bool, result *action.ActionResult, errMsg string, confirmationToken string, act action.Action) {
	if h.registry == nil {
		return false, nil, "registry not configured", "", nil
	}

	act, found := h.registry.Get(actionName)
	if !found {
		return false, nil, "unknown tool: " + actionName, "", nil
	}

	paramsIface := make(map[string]interface{}, len(params))
	for k, v := range params {
		paramsIface[k] = v
	}

	isDestructive = act.Destructive(paramsIface)

	if autonomous && !isDestructive {
		ctx := context.Background()
		res, err := act.Execute(ctx, paramsIface)
		if err != nil {
			return false, nil, err.Error(), "", act
		}
		return false, res, "", "", act
	}

	riskLevel := entity.RiskLevelRead
	if isDestructive {
		riskLevel = entity.RiskLevelDangerous
	}

	tokenStr := ""
	if h.tokenService != nil {
		hash := h.tokenService.ParamsHash(params)
		tok, err := h.tokenService.Generate(assistant.ActionConfirmationToken{
			Action:     actionName,
			ParamsHash: hash,
			RiskLevel:  string(riskLevel),
		})
		if err == nil {
			tokenStr = tok
		}
	}

	return isDestructive, nil, "", tokenStr, act
}

// processStreamToolCalls converts accumulated tool calls into action_required or action_result events (SSE).
// If autonomous is true, safe (non-destructive) actions are executed immediately and results are streamed as action_result.
// Dangerous actions always require confirmation.
func (h *AssistantHandler) processStreamToolCalls(ctx context.Context, w *bufio.Writer, toolCalls map[int]*streamToolCall, autonomous bool, messages []assistant.ChatMessage) {
	if len(toolCalls) == 0 {
		return
	}

	for i := 0; i < len(toolCalls); i++ {
		tc, ok := toolCalls[i]
		if !ok {
			continue
		}

		actionName := tc.name
		params, parseErr := parseToolCallArgs(tc.args)

		if parseErr != "" {
			sendWSEvent(w, WSEventError, WSPayloadError{Message: parseErr})
			continue
		}

		isDestructive, result, errMsg, tokenStr, act := h.executeToolCallAction(actionName, params, autonomous)

		if errMsg != "" && result == nil {
			sendWSEvent(w, WSEventError, WSPayloadError{Message: errMsg})
			continue
		}

		if result != nil {
			result = h.maybeEnrichLogActionResult(ctx, messages, actionName, params, result)
			env := WSEventEnvelope{V: 1, Type: WSEventActionResult, Payload: mustMarshal(map[string]interface{}{
				"action":   actionName,
				"params":   params,
				"result":   result,
				"message":  result.Message,
				"data":     result.Data,
				"executed": true,
			})}
			envBytes, _ := json.Marshal(env)
			fmt.Fprintf(w, "data: %s\n\n", envBytes)
			w.Flush()
			continue
		}

		paramsIface := make(map[string]interface{}, len(params))
		for k, v := range params {
			paramsIface[k] = v
		}

		riskLevel := entity.RiskLevelRead
		if isDestructive {
			riskLevel = entity.RiskLevelDangerous
		}

		intent := entity.ActionIntent{
			Action:               actionName,
			Params:               params,
			RiskLevel:            riskLevel,
			RequiresConfirmation: true,
			ConfirmationMessage:  "",
		}
		if act != nil {
			intent.ConfirmationMessage = act.ConfirmationMessage(paramsIface)
		}
		intent.ConfirmationToken = tokenStr

		intentBytes, _ := json.Marshal(intent)
		env := WSEventEnvelope{V: 1, Type: WSEventActionRequired, Payload: intentBytes}
		envBytes, _ := json.Marshal(env)
		fmt.Fprintf(w, "data: %s\n\n", envBytes)
		w.Flush()
	}
}

// parseToolCallArgs parses JSON arguments from a tool call's arguments string.
// Returns params map and empty error string on success, or empty params and error message on failure.
func parseToolCallArgs(args string) (map[string]string, string) {
	if args == "" {
		return make(map[string]string), ""
	}
	var rawParams map[string]interface{}
	if err := json.Unmarshal([]byte(args), &rawParams); err != nil {
		return nil, "invalid tool arguments"
	}
	params := make(map[string]string)
	for k, v := range rawParams {
		params[k] = fmt.Sprintf("%v", v)
	}
	return params, ""
}

// processWSToolCalls sends accumulated tool calls as action_result or action_required events over WebSocket.
// If autonomous is true, safe (non-destructive) actions are executed immediately and results are sent as action_result.
// Dangerous actions always require confirmation.
func (h *AssistantHandler) processWSToolCalls(ctx context.Context, c *websocket.Conn, toolCalls map[int]*streamToolCall, autonomous bool, messages []assistant.ChatMessage) {
	if len(toolCalls) == 0 {
		return
	}

	for i := 0; i < len(toolCalls); i++ {
		tc, ok := toolCalls[i]
		if !ok {
			continue
		}

		actionName := tc.name
		params, parseErr := parseToolCallArgs(tc.args)

		if parseErr != "" {
			c.WriteJSON(WSEventEnvelope{V: 1, Type: WSEventError, Payload: mustMarshal(WSPayloadError{Message: parseErr})})
			continue
		}

		isDestructive, result, errMsg, tokenStr, act := h.executeToolCallAction(actionName, params, autonomous)

		if errMsg != "" && result == nil {
			c.WriteJSON(WSEventEnvelope{V: 1, Type: WSEventError, Payload: mustMarshal(WSPayloadError{Message: errMsg})})
			continue
		}

		if result != nil {
			result = h.maybeEnrichLogActionResult(ctx, messages, actionName, params, result)
			c.WriteJSON(WSEventEnvelope{V: 1, Type: WSEventActionResult, Payload: mustMarshal(map[string]interface{}{
				"action":   actionName,
				"params":   params,
				"result":   result,
				"message":  result.Message,
				"data":     result.Data,
				"executed": true,
			})})
			continue
		}

		paramsIface := make(map[string]interface{}, len(params))
		for k, v := range params {
			paramsIface[k] = v
		}

		riskLevel := entity.RiskLevelRead
		if isDestructive {
			riskLevel = entity.RiskLevelDangerous
		}

		confirmationMsg := ""
		if act != nil {
			confirmationMsg = act.ConfirmationMessage(paramsIface)
		}

		intent := entity.ActionIntent{
			Action:               actionName,
			Params:               params,
			RiskLevel:            riskLevel,
			RequiresConfirmation: true,
			ConfirmationMessage:  confirmationMsg,
			ConfirmationToken:    tokenStr,
		}

		c.WriteJSON(WSEventEnvelope{V: 1, Type: WSEventActionRequired, Payload: mustMarshal(intent)})
	}
}

func mustMarshal(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func generateAssistantToolDefs(rawTools []map[string]interface{}) []assistant.ToolDef {
	tools := make([]assistant.ToolDef, 0, len(rawTools))
	for _, raw := range rawTools {
		toolBytes, err := json.Marshal(raw)
		if err != nil {
			continue
		}

		var tool assistant.ToolDef
		if err := json.Unmarshal(toolBytes, &tool); err != nil {
			continue
		}
		tools = append(tools, tool)
	}
	return tools
}

// emitAudit logs a structured audit event for an action execution attempt.
// It never blocks execution — logging failures are silently swallowed.
func (h *AssistantHandler) emitAudit(userID, sessionID, actionName string, params map[string]string, riskLevel entity.RiskLevel, result entity.AuditResult, execErr error, tokenValid, tokenExpired bool) {
	event := entity.NewAIAuditEvent(userID, sessionID, entity.AuditSourceREST, actionName, params, riskLevel, result, execErr)
	event.TokenValid = tokenValid
	event.TokenExpired = tokenExpired
	data, _ := json.Marshal(event)
	h.l.Info("audit: %s", string(data))
}

// StreamRequest is the request body for POST /v1/assistant/stream.
type StreamRequest struct {
	Messages   []assistant.ChatMessage `json:"messages"`
	Tools      []assistant.ToolDef     `json:"tools,omitempty"`
	Autonomous bool                    `json:"autonomous,omitempty"`
}

// ExecuteRequest represents a request to execute a confirmed action.
type ExecuteRequest struct {
	Action            string            `json:"action"`
	Params            map[string]string `json:"params"`
	ConfirmationToken string            `json:"confirmation_token"`
}

// Stream handles POST /v1/assistant/stream — SSE streaming chat completion.
func (h *AssistantHandler) Stream(c *fiber.Ctx) error {
	var req StreamRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if len(req.Messages) == 0 {
		return errorResponse(c, fiber.StatusBadRequest, "messages is empty")
	}

	settings := h.settingsStore.Get()
	client := assistant.NewLLMClient(settings.APIEndpoint, settings.APIKey, settings.ModelName)

	messages := req.Messages
	if req.Autonomous {
		messages = prependAgentSystemPrompt(messages)
	}
	tools := req.Tools
	if len(tools) == 0 && h.registry != nil {
		tools = generateAssistantToolDefs(h.registry.GenerateToolDefs())
	}
	resp, err := client.StreamChatCompletion(c.Context(), messages, tools)
	if err != nil {
		h.l.Error(err, "AssistantHandler - Stream - StreamChatCompletion")
		return errorResponse(c, fiber.StatusBadGateway, "LLM stream request failed: "+err.Error())
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "close")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 65536), 65536)

		toolCalls := make(map[int]*streamToolCall)

		for scanner.Scan() {
			line := strings.TrimSuffix(scanner.Text(), "\r")
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				h.processStreamToolCalls(context.Background(), w, toolCalls, req.Autonomous, messages)
				chunk, _ := json.Marshal(streamChunk{Content: "", Done: true})
				fmt.Fprintf(w, "data: %s\n\n", chunk)
				w.Flush()
				return
			}

			var openAIChunk struct {
				Choices []struct {
					Delta struct {
						Content   string                `json:"content"`
						ToolCalls []streamToolCallDelta `json:"tool_calls"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(data), &openAIChunk); err != nil {
				continue
			}

			if len(openAIChunk.Choices) == 0 {
				continue
			}

			choice := openAIChunk.Choices[0]

			for _, tc := range choice.Delta.ToolCalls {
				existing, exists := toolCalls[tc.Index]
				if !exists {
					existing = &streamToolCall{}
					toolCalls[tc.Index] = existing
				}
				if tc.ID != "" {
					existing.id = tc.ID
				}
				if tc.Function.Name != "" {
					existing.name = tc.Function.Name
				}
				existing.args += tc.Function.Arguments
			}

			if choice.FinishReason != nil && *choice.FinishReason == "tool_calls" {
				h.processStreamToolCalls(context.Background(), w, toolCalls, req.Autonomous, messages)
				toolCalls = make(map[int]*streamToolCall)
			}

			content := choice.Delta.Content
			if content == "" {
				continue
			}

			chunk, _ := json.Marshal(streamChunk{Content: content, Done: false})
			fmt.Fprintf(w, "data: %s\n\n", chunk)
			w.Flush()
		}

		if scanner.Err() != nil {
			h.l.Error(scanner.Err(), "AssistantHandler - Stream - scanner")
		}

		h.processStreamToolCalls(context.Background(), w, toolCalls, req.Autonomous, messages)

		chunk, _ := json.Marshal(streamChunk{Content: "", Done: true})
		fmt.Fprintf(w, "data: %s\n\n", chunk)
		w.Flush()
	})

	return nil
}

// StreamWS handles GET /v1/assistant/stream/ws — WebSocket streaming chat completion.
func (h *AssistantHandler) StreamWS(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		_, msgBytes, err := c.ReadMessage()
		if err != nil {
			cancel()
			break
		}

		rateLimitKey := "ws:" + c.RemoteAddr().String()
		if uid, ok := c.Locals("userID").(string); ok && uid != "" {
			rateLimitKey = "user:" + uid
		}

		if h.rateLimiter != nil && !h.rateLimiter.Allow(rateLimitKey) {
			c.WriteJSON(map[string]interface{}{"error": assistant.MapErrorToUserMessage(assistant.ErrRateLimited)})
			continue
		}

		var req struct {
			Messages   []assistant.ChatMessage `json:"messages"`
			Tools      []assistant.ToolDef     `json:"tools,omitempty"`
			Autonomous bool                    `json:"autonomous,omitempty"`
		}
		if err := json.Unmarshal(msgBytes, &req); err != nil || len(req.Messages) == 0 {
			c.WriteJSON(map[string]interface{}{"error": "invalid request: messages required"})
			continue
		}

		if len(req.Messages) > 50 {
			c.WriteJSON(map[string]interface{}{"error": "too many messages, max 50"})
			continue
		}

		valid := true
		for i := range req.Messages {
			if len(req.Messages[i].Content) > 2000 {
				c.WriteJSON(map[string]interface{}{"error": "message content too long, max 2000 characters per message"})
				valid = false
				break
			}
		}
		if !valid {
			continue
		}

		settings := h.settingsStore.Get()
		client := assistant.NewLLMClient(settings.APIEndpoint, settings.APIKey, settings.ModelName)

		messages := req.Messages
		if req.Autonomous {
			messages = prependAgentSystemPrompt(messages)
		}
		tools := req.Tools
		if len(tools) == 0 && h.registry != nil {
			tools = generateAssistantToolDefs(h.registry.GenerateToolDefs())
		}
		resp, err := client.StreamChatCompletion(ctx, messages, tools)
		if err != nil {
			c.WriteJSON(map[string]interface{}{"error": "LLM request failed: " + assistant.MapErrorToUserMessage(err)})
			h.l.Error(err, "AssistantHandler - StreamWS - StreamChatCompletion")
			continue
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 65536), 65536)

		toolCalls := make(map[int]*streamToolCall)
		doneSent := false

		for scanner.Scan() {
			line := strings.TrimSuffix(scanner.Text(), "\r")
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				h.processWSToolCalls(ctx, c, toolCalls, req.Autonomous, messages)
				c.WriteJSON(map[string]interface{}{"done": true})
				doneSent = true
				break
			}

			var openAIChunk struct {
				Choices []struct {
					Delta struct {
						Content   string                `json:"content"`
						ToolCalls []streamToolCallDelta `json:"tool_calls"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(data), &openAIChunk); err != nil {
				continue
			}
			if len(openAIChunk.Choices) == 0 {
				continue
			}

			choice := openAIChunk.Choices[0]

			for _, tc := range choice.Delta.ToolCalls {
				existing, exists := toolCalls[tc.Index]
				if !exists {
					existing = &streamToolCall{}
					toolCalls[tc.Index] = existing
				}
				if tc.ID != "" {
					existing.id = tc.ID
				}
				if tc.Function.Name != "" {
					existing.name = tc.Function.Name
				}
				existing.args += tc.Function.Arguments
			}

			if choice.FinishReason != nil && *choice.FinishReason == "tool_calls" {
				h.processWSToolCalls(ctx, c, toolCalls, req.Autonomous, messages)
				toolCalls = make(map[int]*streamToolCall)
			}

			if choice.Delta.Content != "" {
				c.WriteJSON(map[string]interface{}{"content": choice.Delta.Content, "done": false})
			}
		}

		resp.Body.Close()

		if scanner.Err() != nil {
			h.l.Error(scanner.Err(), "AssistantHandler - StreamWS - scanner")
		}

		h.processWSToolCalls(ctx, c, toolCalls, req.Autonomous, messages)

		if !doneSent {
			c.WriteJSON(map[string]interface{}{"done": true})
		}
	}
}

// Parse handles POST /v1/assistant/parse — natural language intent parsing.
func (h *AssistantHandler) Parse(c *fiber.Ctx) error {
	var req entity.ParseRequest
	if err := c.BodyParser(&req); err != nil {
		h.l.Error(err, "AssistantHandler - Parse")
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if len(strings.TrimSpace(req.Text)) == 0 {
		return errorResponse(c, fiber.StatusBadRequest, "input is empty")
	}

	if len(req.Text) > 500 {
		return errorResponse(c, fiber.StatusBadRequest, "input too long, max 500 characters")
	}

	req.Text = html.EscapeString(req.Text)
	req.Text = assistant_util.StripShellChars(req.Text)

	if err := action.SanitizeInput(req.Text); err != nil {
		h.l.Warn("AssistantHandler - Parse - sanitize: %v", err)
		return errorResponse(c, fiber.StatusBadRequest, "Input contains potentially malicious content")
	}

	resp, err := h.parser.Parse(c.Context(), req.Text)
	if err != nil {
		h.l.Error(err, "AssistantHandler - Parse - parser.Parse")
		return errorResponse(c, fiber.StatusBadRequest, assistant.MapErrorToUserMessage(err))
	}

	return successResponse(c, resp)
}

// ExecuteAction handles POST /v1/assistant/execute — execute a confirmed action.
func (h *AssistantHandler) ExecuteAction(c *fiber.Ctx) error {
	var req ExecuteRequest
	if err := c.BodyParser(&req); err != nil {
		h.l.Error(fmt.Errorf("AssistantHandler - ExecuteAction - BodyParser: %w", err), "")
		return structuredErrorResponse(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}

	if req.Action == "" {
		return structuredErrorResponse(c, fiber.StatusBadRequest, "bad_request", "action is required")
	}

	userID, _ := c.Locals("userID").(string)
	if userID == "" {
		return structuredErrorResponse(c, fiber.StatusUnauthorized, "unauthorized", "authentication required")
	}

	sessionID, _ := c.Locals("sessionID").(string)

	if h.rateLimiter != nil && !h.rateLimiter.Allow("user:"+userID) {
		return structuredErrorResponse(c, fiber.StatusTooManyRequests, "rate_limited", "too many requests, please try again later")
	}

	act, found := h.registry.Get(req.Action)
	if !found {
		return structuredErrorResponse(c, fiber.StatusBadRequest, "bad_request", "unknown action: "+req.Action)
	}

	params := make(map[string]interface{}, len(req.Params))
	for k, v := range req.Params {
		params[k] = v
	}

	riskLevel := entity.RiskLevelRead
	if act.Destructive(params) {
		riskLevel = entity.RiskLevelDangerous
	}

	ctx := c.UserContext()
	if userID != "" {
		ctx = context.WithValue(ctx, action.CtxKeyUserID, userID)
	}
	if sessionID != "" {
		ctx = context.WithValue(ctx, action.CtxKeySessionID, sessionID)
	}

	result, err := h.registry.ExecuteWithConfirmation(ctx, req.Action, params, req.ConfirmationToken)
	if err != nil {
		if errors.Is(err, action.ErrConfirmationRequired) && strings.Contains(err.Error(), "no token service configured") {
			isDestructive := act.Destructive(params)
			if isDestructive {
				if req.ConfirmationToken == "" {
					h.emitAudit(userID, sessionID, req.Action, req.Params, riskLevel, entity.AuditResultRejected, action.ErrConfirmationRequired, false, false)
					return structuredErrorResponse(c, fiber.StatusForbidden, "confirmation_required", "a confirmation token is required for this action")
				}

				if h.tokenService == nil {
					h.emitAudit(userID, sessionID, req.Action, req.Params, riskLevel, entity.AuditResultFailed, fmt.Errorf("token service not configured"), false, false)
					return structuredErrorResponse(c, fiber.StatusInternalServerError, "internal_error", "token service not configured")
				}

				hash := h.tokenService.ParamsHash(req.Params)
				validationParams := assistant.ActionConfirmationToken{
					UserID:     userID,
					SessionID:  sessionID,
					Action:     req.Action,
					ParamsHash: hash,
					RiskLevel:  string(entity.RiskLevelDangerous),
				}

				if vErr := h.tokenService.Validate(req.ConfirmationToken, validationParams); vErr != nil {
					h.l.Error(fmt.Errorf("AssistantHandler - ExecuteAction - tokenService.Validate: %w", vErr), "")
					tokenExpired := errors.Is(vErr, action.ErrTokenExpired)
					h.emitAudit(userID, sessionID, req.Action, req.Params, riskLevel, entity.AuditResultRejected, vErr, false, tokenExpired)
					return structuredErrorResponse(c, fiber.StatusForbidden, "confirmation_token_invalid", "confirmation token is invalid or expired")
				}
			}

			result, err = h.registry.ExecuteConfirmed(ctx, req.Action, params)
			if err != nil {
				h.l.Error(fmt.Errorf("AssistantHandler - ExecuteAction - registry.ExecuteConfirmed: %w", err), "")
				h.emitAudit(userID, sessionID, req.Action, req.Params, riskLevel, entity.AuditResultFailed, err, true, false)
				return structuredErrorResponse(c, fiber.StatusBadRequest, "execution_failed", "Failed to execute action: "+err.Error())
			}

			h.emitAudit(userID, sessionID, req.Action, req.Params, riskLevel, entity.AuditResultSuccess, nil, true, false)
			return successResponse(c, result)
		}

		if errors.Is(err, action.ErrConfirmationRequired) {
			h.emitAudit(userID, sessionID, req.Action, req.Params, riskLevel, entity.AuditResultRejected, err, false, false)
			return structuredErrorResponse(c, fiber.StatusForbidden, "confirmation_required", err.Error())
		}

		if errors.Is(err, action.ErrInvalidToken) || errors.Is(err, action.ErrTokenExpired) {
			tokenExpired := errors.Is(err, action.ErrTokenExpired)
			h.emitAudit(userID, sessionID, req.Action, req.Params, riskLevel, entity.AuditResultRejected, err, false, tokenExpired)
			return structuredErrorResponse(c, fiber.StatusForbidden, "confirmation_token_invalid", "confirmation token is invalid or expired")
		}

		h.l.Error(fmt.Errorf("AssistantHandler - ExecuteAction - registry.ExecuteWithConfirmation: %w", err), "")
		h.emitAudit(userID, sessionID, req.Action, req.Params, riskLevel, entity.AuditResultFailed, err, true, false)
		return structuredErrorResponse(c, fiber.StatusBadRequest, "execution_failed", "Failed to execute action: "+err.Error())
	}

	h.emitAudit(userID, sessionID, req.Action, req.Params, riskLevel, entity.AuditResultSuccess, nil, true, false)
	return successResponse(c, result)
}

// Diagnose handles POST /v1/assistant/diagnose — container fault diagnosis.
func (h *AssistantHandler) Diagnose(c *fiber.Ctx) error {
	var req entity.DiagnoseRequest
	if err := c.BodyParser(&req); err != nil {
		h.l.Error(err, "AssistantHandler - Diagnose")
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if len(strings.TrimSpace(req.ContainerID)) == 0 {
		return errorResponse(c, fiber.StatusBadRequest, "container_id is empty")
	}

	if len(req.ContainerID) > 500 {
		return errorResponse(c, fiber.StatusBadRequest, "container_id too long, max 500 characters")
	}

	if len(strings.TrimSpace(req.ContainerName)) == 0 {
		return errorResponse(c, fiber.StatusBadRequest, "container_name is empty")
	}

	if len(req.ContainerName) > 500 {
		return errorResponse(c, fiber.StatusBadRequest, "container_name too long, max 500 characters")
	}

	req.ContainerID = html.EscapeString(req.ContainerID)
	req.ContainerID = assistant_util.StripShellChars(req.ContainerID)
	req.ContainerName = html.EscapeString(req.ContainerName)
	req.ContainerName = assistant_util.StripShellChars(req.ContainerName)

	resp, err := h.diagnosis.Diagnose(c.Context(), req.ContainerID, req.ExitCode)
	if err != nil {
		h.l.Error(err, "AssistantHandler - Diagnose - diagnosis.Diagnose")
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to diagnose")
	}

	return successResponse(c, resp)
}

// Recommend handles POST /v1/assistant/recommend — configuration recommendations.
func (h *AssistantHandler) Recommend(c *fiber.Ctx) error {
	var req entity.RecommendRequest
	if err := c.BodyParser(&req); err != nil {
		h.l.Error(err, "AssistantHandler - Recommend")
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if len(strings.TrimSpace(req.Scenario)) == 0 {
		return errorResponse(c, fiber.StatusBadRequest, "scenario is empty")
	}

	if len(req.Scenario) > 500 {
		return errorResponse(c, fiber.StatusBadRequest, "scenario too long, max 500 characters")
	}

	req.Scenario = html.EscapeString(req.Scenario)
	req.Scenario = assistant_util.StripShellChars(req.Scenario)

	resp, err := h.recommend.Recommend(c.Context(), req.Scenario)
	if err != nil {
		h.l.Error(err, "AssistantHandler - Recommend - recommend.Recommend")
		return errorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return successResponse(c, resp)
}
