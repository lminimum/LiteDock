package v1

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lminimum/LiteDock/internal/action"
	"github.com/lminimum/LiteDock/internal/usecase/assistant"
)

const (
	_agentModeSystemPrompt   = "You are LiteDock AI Assistant in autonomous mode. Prefer using tools to inspect real Docker state instead of suggesting shell commands. When a user asks about container or compose logs, fetch the logs and analyze them yourself. Reply with concise findings, likely causes, and next steps. Do not instruct the user to run docker logs or other shell commands manually."
	_logAnalysisSystemPrompt = "You are LiteDock's log analysis assistant. Analyze the provided Docker logs and answer in the same language as the conversation. Return a Markdown report with short headings, bullet points, and bold labels so it renders clearly in the UI. Focus on errors, warnings, probable causes, and concrete next steps. Do not suggest shell commands or reprint the entire log unless needed."
	_maxLogAnalysisChars     = 12000
)

func prependAgentSystemPrompt(messages []assistant.ChatMessage) []assistant.ChatMessage {
	out := make([]assistant.ChatMessage, 0, len(messages)+1)
	out = append(out, assistant.ChatMessage{Role: "system", Content: _agentModeSystemPrompt})
	out = append(out, messages...)
	return out
}

func (h *AssistantHandler) maybeEnrichLogActionResult(
	ctx context.Context,
	messages []assistant.ChatMessage,
	actionName string,
	params map[string]string,
	result *action.ActionResult,
) *action.ActionResult {
	if result == nil || !shouldAnalyzeLogAction(params) {
		return result
	}

	summary := h.summarizeLogAction(ctx, messages, actionName, params, result)
	if strings.TrimSpace(summary) == "" {
		return result
	}

	enriched := *result
	enriched.Message = summary

	if dataMap, ok := result.Data.(map[string]interface{}); ok {
		cloned := make(map[string]interface{}, len(dataMap)+1)
		for k, v := range dataMap {
			cloned[k] = v
		}
		cloned["analysis"] = summary
		enriched.Data = cloned
	}

	return &enriched
}

func (h *AssistantHandler) summarizeLogAction(
	ctx context.Context,
	messages []assistant.ChatMessage,
	actionName string,
	params map[string]string,
	result *action.ActionResult,
) string {
	logs := extractLogText(result)
	if strings.TrimSpace(logs) == "" {
		return "No logs were returned."
	}

	if h.settingsStore == nil {
		return ""
	}

	settings := h.settingsStore.Get()
	if strings.TrimSpace(settings.APIEndpoint) == "" || strings.TrimSpace(settings.ModelName) == "" {
		return ""
	}

	client := assistant.NewLLMClient(settings.APIEndpoint, settings.APIKey, settings.ModelName)
	analysisMessages := make([]assistant.ChatMessage, 0, len(messages)+2)
	analysisMessages = append(analysisMessages, messages...)
	analysisMessages = append(analysisMessages, assistant.ChatMessage{
		Role:    "user",
		Content: buildLogAnalysisPrompt(actionName, params, logs),
	})

	analysisCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := client.ChatCompletion(analysisCtx, append([]assistant.ChatMessage{{Role: "system", Content: _logAnalysisSystemPrompt}}, analysisMessages...), nil)
	if err != nil || resp == nil {
		return ""
	}

	return strings.TrimSpace(resp.Content)
}

func shouldAnalyzeLogAction(params map[string]string) bool {
	op := strings.TrimSpace(params["operation"])
	return strings.HasSuffix(op, "_logs") || strings.Contains(op, "logs")
}

func extractLogText(result *action.ActionResult) string {
	if result == nil {
		return ""
	}

	dataMap, ok := result.Data.(map[string]interface{})
	if !ok {
		return ""
	}

	raw, ok := dataMap["logs"]
	if !ok {
		return ""
	}

	switch logs := raw.(type) {
	case string:
		return logs
	case []byte:
		return string(logs)
	default:
		return fmt.Sprintf("%v", logs)
	}
}

func buildLogAnalysisPrompt(actionName string, params map[string]string, logs string) string {
	trimmedLogs := truncateLogText(strings.TrimSpace(logs))

	var b strings.Builder
	b.WriteString("Analyze the following Docker logs and answer directly to the user.\n")
	b.WriteString("Return the answer as a concise Markdown report. Use headings and bullet points.\n")
	b.WriteString("Preferred structure:\n")
	b.WriteString("### Summary\n")
	b.WriteString("### Key errors and warnings\n")
	b.WriteString("### Likely cause\n")
	b.WriteString("### Next steps\n")
	b.WriteString("Action: ")
	b.WriteString(actionName)
	b.WriteString("\n")

	if len(params) > 0 {
		b.WriteString("Parameters:\n")
		for k, v := range params {
			b.WriteString("- ")
			b.WriteString(k)
			b.WriteString(": ")
			b.WriteString(v)
			b.WriteString("\n")
		}
	}

	b.WriteString("\nFocus on errors, warnings, probable root cause, and concrete next steps. Do not suggest shell commands. Keep the answer concise and readable.\n\n")
	b.WriteString("Logs:\n```text\n")
	b.WriteString(trimmedLogs)
	b.WriteString("\n```")

	return b.String()
}

func truncateLogText(logs string) string {
	if len(logs) <= _maxLogAnalysisChars {
		return logs
	}

	start := len(logs) - _maxLogAnalysisChars
	if start < 0 {
		start = 0
	}

	return "[truncated logs]\n" + logs[start:]
}
