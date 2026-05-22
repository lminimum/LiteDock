package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
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

// _chineseSynonyms maps Chinese/English action keywords to their normalized action names.
var _chineseSynonyms = map[string]string{
	"停止":       "stop_container",
	"关掉":       "stop_container",
	"停掉":       "stop_container",
	"关闭":       "stop_container",
	"shutdown": "stop_container",
	"stop":     "stop_container",
	"重启":       "restart_container",
	"重新启动":     "restart_container",
	"restart":  "restart_container",
	"reload":   "restart_container",
	"删除":       "delete_container",
	"移除":       "delete_container",
	"销毁":       "delete_container",
	"delete":   "delete_container",
	"remove":   "delete_container",
	"destroy":  "delete_container",
	"查看日志":     "get_container_logs",
	"日志":       "get_container_logs",
	"logs":     "get_container_logs",
	"tail":     "get_container_logs",
	"输出":       "get_container_logs",
	"查看":       "get_container_logs",
	"列出":       "list_containers",
	"列表":       "list_containers",
	"查看列表":     "list_containers",
	"list":     "list_containers",
	"ls":       "list_containers",
	"show":     "list_containers",
	"启动":       "start_container",
	"开启":       "start_container",
	"运行":       "start_container",
	"开始":       "start_container",
	"start":    "start_container",
	"创建":       "create_container",
	"新建":       "create_container",
	"新增":       "create_container",
	"create":   "create_container",
	"new":      "create_container",
	"拉取":       "pull_image",
	"下载镜像":     "pull_image",
	"pull":     "pull_image",
	"fetch":    "pull_image",
	"网络":       "network_inspect",
	"网络详情":     "network_inspect",
	"network":  "network_inspect",
	"net":      "network_inspect",
	"卷":        "volume_operations",
	"存储卷":      "volume_operations",
	"volume":   "volume_operations",
	"存储":       "volume_operations",
	"状态":       "container_stats",
	"资源":       "container_stats",
	"统计":       "container_stats",
	"stats":    "container_stats",
	"资源使用":     "container_stats",
	"检查":       "container_inspect",
	"详情":       "container_inspect",
	"inspect":  "container_inspect",
	"详情信息":     "container_inspect",
}

// _logTailPatterns are regex patterns for extracting log tail count.
var _logTailPatterns = []*regexp.Regexp{
	regexp.MustCompile(`最后\s*(\d+)\s*(?:行|条|个)?`),
	regexp.MustCompile(`(\d+)\s*行\s*(?:日志|log|logs)?`),
	regexp.MustCompile(`last\s+(\d+)\s*(?:lines?|log|logs)?`),
	regexp.MustCompile(`tail\s+[a-zA-Z0-9_-]+\s+(\d+)`),
	regexp.MustCompile(`tail\s+(\d+)`),
	regexp.MustCompile(`最近\s*(\d+)\s*(?:行)?(?:日志)?`),
	regexp.MustCompile(`(\d+)\s*(?:lines?|log|logs)`),
}

// _containerIDPatterns are regex patterns for extracting container IDs (partial hex strings).
var _containerIDPatterns = []*regexp.Regexp{
	regexp.MustCompile(`([0-9a-f]{6,12})`),
}

// _machineIDPatterns are regex patterns for extracting machine IDs.
var _machineIDPatterns = []*regexp.Regexp{
	regexp.MustCompile(`machine[-_]?uuid[:\s]+([a-zA-Z0-9_-]+)`),
}

// _containerNamePatterns are regex patterns for extracting container names from Chinese queries.
// These are applied AFTER finding the action keyword in the text.
var _containerNamePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?:停止|关闭|停掉|重启|删除|查看|列出|启动|查看日志|stop|start|restart|delete|list|logs?|tail)\s+([a-zA-Z0-9_-]+)`),
	regexp.MustCompile(`(?:的\s*(?:日志|状态|详情|容器))?([a-zA-Z0-9_-]+)`),
}

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

// _destructiveKeywords are keywords that indicate a potentially destructive operation.
// Used as a fail-safe when no rule matches but the user intent is clearly modifying.
var _destructiveKeywords = []string{
	"删除", "删除所有", "delete", "prune", "remove", "destroy", "drop",
	"停止", "关掉", "重启", "restart",
}

// containsDestructiveKeyword checks if the input text contains any destructive keyword.
func containsDestructiveKeyword(text string) bool {
	lower := strings.ToLower(text)
	for _, kw := range _destructiveKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// normalizeAction maps a TF-IDF rule action to a registered action name.
// If ruleAction matches a registered action directly, it is returned as-is.
// Otherwise, it tries to find a registered action by extracting suffixes
// (e.g., "stop_container" → "container") and returns the normalized name
// with the original ruleAction as the "operation" parameter.
// It also handles Chinese synonyms by looking up _chineseSynonyms map.
func (uc *NLParserUseCase) normalizeAction(ruleAction string) (actionName string, extraParams map[string]string) {
	if uc.actionRegistry == nil {
		if normalized, ok := _chineseSynonyms[strings.ToLower(ruleAction)]; ok {
			return normalized, nil
		}
		return ruleAction, nil
	}

	if _, ok := uc.actionRegistry.Get(ruleAction); ok {
		return ruleAction, nil
	}

	if normalized, ok := _chineseSynonyms[strings.ToLower(ruleAction)]; ok {
		if _, exists := uc.actionRegistry.Get(normalized); exists {
			return normalized, nil
		}
	}

	parts := strings.Split(ruleAction, "_")
	for i := 1; i < len(parts); i++ {
		candidate := strings.Join(parts[i:], "_")
		if a, ok := uc.actionRegistry.Get(candidate); ok {
			for _, p := range a.Params() {
				if p.Name == "operation" {
					return candidate, map[string]string{"operation": ruleAction}
				}
			}
		}
		singular := strings.TrimSuffix(candidate, "s")
		if singular != candidate {
			if a, ok := uc.actionRegistry.Get(singular); ok {
				for _, p := range a.Params() {
					if p.Name == "operation" {
						return singular, map[string]string{"operation": ruleAction}
					}
				}
			}
		}
	}

	return ruleAction, nil
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

// Close stops the internal rate limiter.
func (uc *NLParserUseCase) Close() {
	if uc.rateLimiter != nil {
		uc.rateLimiter.Close()
	}
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
		if containsDestructiveKeyword(text) {
			uc.logger.Warn(fmt.Sprintf("NLParserUseCase - parseWithTFIDF - destructive keyword in unmatched input: %s", text))
			return entity.ParseResponse{
				Intent:               "unknown",
				Description:          "该操作需要确认",
				Params:               make(map[string]string),
				RequiresConfirmation: true,
			}, nil
		}
		return entity.ParseResponse{
			Intent:      "unknown",
			Description: "未识别您的指令",
			Params:      make(map[string]string),
		}, nil
	}

	actionName, extraParams := uc.normalizeAction(rule.Action)
	params := uc.extractParams(text, rule.Action)
	for k, v := range extraParams {
		params[k] = v
	}

	resp := entity.ParseResponse{
		Intent:      rule.Intent,
		Action:      actionName,
		Description: rule.Description,
		Params:      params,
	}

	if uc.actionRegistry != nil {
		if a, ok := uc.actionRegistry.Get(actionName); ok {
			ifParams := make(map[string]interface{}, len(params))
			for k, v := range params {
				ifParams[k] = v
			}
			if a.Destructive(ifParams) {
				resp.RequiresConfirmation = true
				resp.ConfirmationMessage = a.ConfirmationMessage(ifParams)
				resp.ActionName = actionName
				resp.ActionParams = params
				uc.logger.Warn(fmt.Sprintf("NLParserUseCase - parseWithTFIDF - destructive action requires confirmation: %s", actionName))
			}
		}
	}

	return resp, nil
}

// extractParams finds potential parameter values in the input text after the action keyword.
// For compound actions like "start_container", it extracts the verb prefix ("start")
// to locate the action keyword in the user's text.
// It also handles Chinese patterns for container name extraction and log tail count.
func (uc *NLParserUseCase) extractParams(text string, action string) map[string]string {
	params := make(map[string]string)

	textLower := strings.ToLower(text)
	actionLower := strings.ToLower(action)

	searchKey := actionLower
	if idx := strings.Index(actionLower, "_"); idx > 0 {
		searchKey = actionLower[:idx]
	}

	idx := strings.Index(textLower, searchKey)
	if idx < 0 {
		chineseKey := uc.findChineseKeywordForAction(actionLower)
		if chineseKey != "" {
			searchKey = chineseKey
			idx = strings.Index(textLower, searchKey)
		}
		if idx < 0 {
			uc.extractContainerNameFromRemaining(textLower, params)
			uc.extractLogTail(textLower, params)
			uc.extractContainerID(textLower, params)
			uc.extractMachineID(textLower, params)
			return params
		}
	}

	remaining := strings.TrimSpace(text[idx+len(searchKey):])

	uc.extractLogTail(remaining, params)
	uc.extractContainerID(remaining, params)
	uc.extractMachineID(remaining, params)
	uc.extractContainerNameFromRemaining(remaining, params)

	if remaining == "" {
		return params
	}

	if _, exists := params["container_name"]; !exists {
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
	}

	return params
}

// findChineseKeywordForAction finds a Chinese keyword that maps to the given action prefix.
// It prefers Chinese keywords over English ones for better NL parsing.
func (uc *NLParserUseCase) findChineseKeywordForAction(action string) string {
	actionLower := strings.ToLower(action)
	prefix := actionLower
	if idx := strings.Index(actionLower, "_"); idx > 0 {
		prefix = actionLower[:idx]
	}

	var chineseKey, englishKey string
	for key, value := range _chineseSynonyms {
		if value == prefix || value == actionLower || strings.HasPrefix(value, prefix+"_") {
			isChinese := len(key) > 0 && []rune(key)[0] > 127
			if isChinese && chineseKey == "" {
				chineseKey = key
			} else if !isChinese && englishKey == "" {
				englishKey = key
			}
		}
	}
	if chineseKey != "" {
		return chineseKey
	}
	return englishKey
}

// extractLogTail extracts log tail count from text using regex patterns.
func (uc *NLParserUseCase) extractLogTail(text string, params map[string]string) {
	for _, pattern := range _logTailPatterns {
		if matches := pattern.FindStringSubmatch(text); len(matches) > 1 {
			params["tail"] = matches[1]
			return
		}
	}
}

// extractContainerNameFromRemaining extracts container name from remaining text using Chinese/English patterns.
func (uc *NLParserUseCase) extractContainerNameFromRemaining(remaining string, params map[string]string) {
	if len(remaining) == 0 {
		return
	}

	tokens, err := uc.tokenizer.Tokenize(remaining)
	if err != nil || len(tokens) == 0 {
		return
	}

	firstToken := strings.ToLower(tokens[0])
	if _stopWords[firstToken] {
		return
	}

	for _, pattern := range _containerNamePatterns {
		if matches := pattern.FindStringSubmatch(remaining); len(matches) > 1 {
			if _, exists := params["container_name"]; !exists {
				params["container_name"] = matches[1]
			}
			return
		}
	}
}

// extractContainerID extracts container ID (partial hex string) from text.
func (uc *NLParserUseCase) extractContainerID(text string, params map[string]string) {
	for _, pattern := range _containerIDPatterns {
		if matches := pattern.FindStringSubmatch(text); len(matches) > 1 {
			if _, exists := params["container_id"]; !exists {
				params["container_id"] = matches[1]
			}
			return
		}
	}
}

// extractMachineID extracts machine ID from patterns like "machine-uuid xxx".
func (uc *NLParserUseCase) extractMachineID(text string, params map[string]string) {
	for _, pattern := range _machineIDPatterns {
		if matches := pattern.FindStringSubmatch(text); len(matches) > 1 {
			if _, exists := params["machine_id"]; !exists {
				params["machine_id"] = matches[1]
			}
			return
		}
	}
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
