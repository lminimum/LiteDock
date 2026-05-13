package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lminimum/LiteDock/internal/action"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/assistant/engine"
	tknzr "github.com/lminimum/LiteDock/pkg/assistant/tokenizer"
	"github.com/lminimum/LiteDock/pkg/logger"
)

const (
	_maxInputLength = 500

	// _systemPrompt instructs the LLM on how to respond and when to use tools.
	_systemPrompt = "You are LiteDock AI Assistant. You help users manage Docker containers, images, networks, and volumes. When users request an action, use the available tools. Otherwise, answer conversationally."
)

// ActionRequest represents a parsed tool call from the LLM.
type ActionRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// stopWords are common words filtered from parameter extraction.
var _stopWords = map[string]bool{
	"container":  true,
	"containers": true,
	"image":      true,
	"images":     true,
	"app":        true,
	"容器":         true,
	"镜像":         true,
	"应用":         true,
}

// LLMClientInterface defines the method needed for LLM-based parsing.
type LLMClientInterface interface {
	ChatCompletion(ctx context.Context, messages []ChatMessage, tools []ToolDef) (*ChatResponse, error)
}

// NLParserUseCase parses user natural language input into structured intents and parameters.
// It tries the LLM first (if configured) and falls back to the TF-IDF engine.
type NLParserUseCase struct {
	engine         *engine.Engine
	tokenizer      engine.Tokenizer
	llmClient      LLMClientInterface
	actionRegistry *action.ActionRegistry
	logger         logger.Interface
	rateLimiter    *RateLimiter
}

// NewNLParserUseCase creates a new NLParserUseCase with the given engine, tokenizer, and logger.
// Only TF-IDF engine will be used; call NewNLParser for LLM support.
func NewNLParserUseCase(engine *engine.Engine, tokenizer engine.Tokenizer, logger logger.Interface) *NLParserUseCase {
	return &NLParserUseCase{
		engine:      engine,
		tokenizer:   tokenizer,
		logger:      logger,
		rateLimiter: NewRateLimiter(),
	}
}

// SetActionRegistry sets the action registry on the parser for checking destructive actions.
func (uc *NLParserUseCase) SetActionRegistry(reg *action.ActionRegistry) {
	uc.actionRegistry = reg
}

// NewNLParser creates a new NLParserUseCase with LLM client and action registry.
// The TF-IDF engine and tokenizer are kept as fallback when LLM is unavailable.
func NewNLParser(llmClient LLMClientInterface, actionRegistry *action.ActionRegistry,
	engine *engine.Engine, tokenizer engine.Tokenizer, logger logger.Interface) *NLParserUseCase {
	return &NLParserUseCase{
		engine:         engine,
		tokenizer:      tokenizer,
		llmClient:      llmClient,
		actionRegistry: actionRegistry,
		logger:         logger,
		rateLimiter:    NewRateLimiter(),
	}
}

// Parse analyzes user input text and returns a structured ParseResponse.
// It first attempts to use the LLM (if configured). If the LLM is unavailable or
// returns an error, it falls back to the TF-IDF engine.
//
// When the LLM returns plain text, it is wrapped as a chat response.
// When the LLM returns a tool_call, it is parsed into an action request.
func (uc *NLParserUseCase) Parse(ctx context.Context, text string) (entity.ParseResponse, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return entity.ParseResponse{}, fmt.Errorf("请输入指令")
	}

	if len(text) > _maxInputLength {
		return entity.ParseResponse{}, fmt.Errorf("指令过长，最多%d个字符", _maxInputLength)
	}

	if uc.llmClient != nil && uc.actionRegistry != nil {
		if uc.rateLimiter != nil && !uc.rateLimiter.Allow("global") {
			uc.logger.Warn("NLParserUseCase - Parse - rate limited")
			return entity.ParseResponse{}, fmt.Errorf("%w: %s", ErrRateLimited, MsgRateLimited)
		}

		resp, err := uc.parseWithLLM(ctx, text)
		if err == nil {
			return resp, nil
		}
		uc.logger.Warn(fmt.Sprintf("NLParserUseCase - Parse - LLM failed, falling back to TF-IDF: %v", err))
	}

	return uc.parseWithTFIDF(text)
}

func (uc *NLParserUseCase) parseWithLLM(ctx context.Context, text string) (entity.ParseResponse, error) {
	messages := []ChatMessage{
		{Role: "system", Content: _systemPrompt},
		{Role: "user", Content: text},
	}

	toolDefs := uc.actionRegistry.GenerateToolDefs()
	tools := convertToolDefs(toolDefs)

	start := time.Now()
	requestID := generateRequestID()

	llmResp, err := uc.llmClient.ChatCompletion(ctx, messages, tools)
	duration := time.Since(start)

	if err != nil {
		uc.logger.Error(fmt.Sprintf("LLM API error: request_id=%s duration=%s status=error error=%v", requestID, duration, err))
		return entity.ParseResponse{}, fmt.Errorf("NLParserUseCase - parseWithLLM - ChatCompletion: %w", err)
	}

	uc.logger.Info(fmt.Sprintf("LLM API success: request_id=%s duration=%s", requestID, duration))

	if len(llmResp.ToolCalls) > 0 {
		tc := llmResp.ToolCalls[0]
		args := make(map[string]interface{})
		if tc.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				uc.logger.Warn(fmt.Sprintf("NLParserUseCase - parseWithLLM - unmarshal arguments: %v", err))
			}
		}

		strParams := make(map[string]string, len(args))
		for k, v := range args {
			switch val := v.(type) {
			case string:
				strParams[k] = val
			default:
				if b, err := json.Marshal(val); err == nil {
					strParams[k] = string(b)
				}
			}
		}

		// Check if this action requires confirmation
		resp := entity.ParseResponse{
			Intent:      tc.Function.Name,
			Action:      tc.Function.Name,
			Description: fmt.Sprintf("Executing: %s", tc.Function.Name),
			Params:      strParams,
		}

		if actionObj, ok := uc.actionRegistry.Get(tc.Function.Name); ok {
			if actionObj.Destructive(args) {
				resp.RequiresConfirmation = true
				resp.ConfirmationMessage = actionObj.ConfirmationMessage(args)
				resp.ActionName = tc.Function.Name
				resp.ActionParams = strParams
			}
		}

		return resp, nil
	}

	return entity.ParseResponse{
		Intent:      "chat",
		Description: llmResp.Content,
		Params:      make(map[string]string),
	}, nil
}

func (uc *NLParserUseCase) parseWithTFIDF(text string) (entity.ParseResponse, error) {
	rule, score, err := uc.engine.Match(text)
	if err != nil {
		return entity.ParseResponse{}, fmt.Errorf("NLParserUseCase - Parse - engine.Match: %w", err)
	}

	if score == 0 || rule.Name == "" {
		return entity.ParseResponse{
			Intent:      "unknown",
			Description: "未识别您的指令",
			Params:      make(map[string]string),
		}, nil
	}

	params := uc.extractParams(text, rule.Action)

	return entity.ParseResponse{
		Intent:      rule.Intent,
		Action:      rule.Action,
		Description: rule.Description,
		Params:      params,
	}, nil
}

// extractParams finds potential parameter values in the input text after the action keyword.
// It tokenizes the remaining text, filters out stop words, and maps the first valid token
// as the container_name parameter.
func (uc *NLParserUseCase) extractParams(text string, action string) map[string]string {
	params := make(map[string]string)

	textLower := strings.ToLower(text)
	actionLower := strings.ToLower(action)
	idx := strings.Index(textLower, actionLower)
	if idx < 0 {
		return params
	}

	remaining := strings.TrimSpace(text[idx+len(action):])
	if remaining == "" {
		return params
	}

	tokens, err := uc.tokenizer.Tokenize(remaining)
	if err != nil {
		uc.logger.Warn(fmt.Sprintf("NLParserUseCase - extractParams - tokenize remaining text: %v", err))

		return params
	}

	for _, token := range tokens {
		lower := strings.ToLower(token)
		if _stopWords[lower] {
			continue
		}

		params["container_name"] = lower

		break
	}

	return params
}

// NewNLParserTokenizer creates a new Tokenizer instance for NL parsing.
// The returned tokenizer should be closed with Close() when no longer needed.
func NewNLParserTokenizer() (*tknzr.Tokenizer, error) {
	return tknzr.NewTokenizer()
}

func convertToolDefs(rawTools []map[string]interface{}) []ToolDef {
	tools := make([]ToolDef, 0, len(rawTools))
	for _, raw := range rawTools {
		tool := ToolDef{Type: "function"}
		if fn, ok := raw["function"].(map[string]interface{}); ok {
			tool.Function.Name, _ = fn["name"].(string)
			tool.Function.Description, _ = fn["description"].(string)
			tool.Function.Parameters = fn["parameters"]
		}
		tools = append(tools, tool)
	}
	return tools
}

func generateRequestID() string {
	return time.Now().Format("20060102150405.000000")
}
