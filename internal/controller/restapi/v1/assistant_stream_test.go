package v1

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/action"
	"github.com/lminimum/LiteDock/internal/usecase/assistant"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/stretchr/testify/require"
)

type testLogger struct{}

func (t *testLogger) Debug(_ interface{}, _ ...interface{}) {}
func (t *testLogger) Info(_ string, _ ...interface{})       {}
func (t *testLogger) Warn(_ string, _ ...interface{})       {}
func (t *testLogger) Error(_ interface{}, _ ...interface{}) {}
func (t *testLogger) Fatal(_ interface{}, _ ...interface{}) {}

var _ logger.Interface = (*testLogger)(nil)

func newMockLLMStreamServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/v1/chat/completions", r.URL.Path)

		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		require.Equal(t, true, reqBody["stream"])

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`data: {"choices":[{"delta":{"content":"Hel"}}]}`,
			`data: {"choices":[{"delta":{"content":"lo"}}]}`,
			`data: {"choices":[{"delta":{"content":" World"}}]}`,
			`data: {"choices":[{"delta":{"content":"!"}}]}`,
			`data: [DONE]`,
		}
		for _, chunk := range chunks {
			fmt.Fprintf(w, "%s\n\n", chunk)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
}

func TestStreamHandler_Success(t *testing.T) {
	mockLLM := newMockLLMStreamServer(t)
	defer mockLLM.Close()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	store := NewAISettingsStore(mockLLM.URL, "test-key", "test-model")

	messages := []assistant.ChatMessage{
		{Role: "user", Content: "Hello"},
	}

	handler := &AssistantHandler{
		settingsStore: store,
		l:             &testLogger{},
	}

	app.Post("/stream", handler.Stream)

	body, _ := json.Marshal(map[string]interface{}{
		"messages": messages,
	})

	req := httptest.NewRequest("POST", "/stream", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	var fullContent string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var chunk streamChunk
		err := json.Unmarshal([]byte(data), &chunk)
		require.NoError(t, err, "invalid chunk: %s", data)

		if chunk.Done {
			break
		}
		fullContent += chunk.Content
	}
	require.NoError(t, scanner.Err())
	require.Equal(t, "Hello World!", fullContent)
}

func TestStreamHandler_EmptyMessages(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	store := NewAISettingsStore("http://localhost", "", "test-model")

	handler := &AssistantHandler{
		settingsStore: store,
		l:             &testLogger{},
	}

	app.Post("/stream", handler.Stream)

	body, _ := json.Marshal(map[string]interface{}{
		"messages": []interface{}{},
	})

	req := httptest.NewRequest("POST", "/stream", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	require.Contains(t, string(respBody), "messages is empty")
}

func TestStreamHandler_LLMError(t *testing.T) {
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "internal error"},
		})
	}))
	defer errorServer.Close()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	store := NewAISettingsStore(errorServer.URL, "", "test-model")

	handler := &AssistantHandler{
		settingsStore: store,
		l:             &testLogger{},
	}

	app.Post("/stream", handler.Stream)

	messages := []assistant.ChatMessage{
		{Role: "user", Content: "Hi"},
	}
	body, _ := json.Marshal(map[string]interface{}{
		"messages": messages,
	})

	req := httptest.NewRequest("POST", "/stream", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadGateway, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	require.Contains(t, string(respBody), "LLM stream request failed")
}

type testAction struct {
	name        string
	destructive bool
}

func (a *testAction) Name() string        { return a.name }
func (a *testAction) Description() string { return "test action " + a.name }
func (a *testAction) Params() []action.ParamDef {
	return []action.ParamDef{
		{Name: "container_id", Type: "string", Required: true, Description: "container ID"},
	}
}
func (a *testAction) Validate(_ map[string]interface{}) error       { return nil }
func (a *testAction) Destructive(_ map[string]interface{}) bool     { return a.destructive }
func (a *testAction) ConfirmationMessage(_ map[string]interface{}) string { return "confirm " + a.name }
func (a *testAction) Execute(_ context.Context, _ map[string]interface{}) (*action.ActionResult, error) {
	return &action.ActionResult{Success: true}, nil
}

var _ action.Action = (*testAction)(nil)

func newMockLLMToolCallServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_abc","type":"function","function":{"name":"container","arguments":""}}]}}]}`,
			`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"cont"}}]}}]}`,
			`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"ainer_id\":\"web\"}"}}]}}]}`,
			`data: {"choices":[{"finish_reason":"tool_calls"}]}`,
			`data: [DONE]`,
		}
		for _, chunk := range chunks {
			fmt.Fprintf(w, "%s\n\n", chunk)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
}

func TestStreamHandler_ToolCall_ActionRequired(t *testing.T) {
	mockLLM := newMockLLMToolCallServer(t)
	defer mockLLM.Close()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	store := NewAISettingsStore(mockLLM.URL, "test-key", "test-model")

	reg := action.NewActionRegistry()
	require.NoError(t, reg.Register(&testAction{name: "container"}))

	handler := &AssistantHandler{
		settingsStore: store,
		registry:      reg,
		tokenService:  assistant.NewTokenService("", 0),
		l:             &testLogger{},
	}

	app.Post("/stream", handler.Stream)

	body, _ := json.Marshal(map[string]interface{}{
		"messages": []assistant.ChatMessage{{Role: "user", Content: "stop web container"}},
	})

	req := httptest.NewRequest("POST", "/stream", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	foundActionRequired := false
	foundDone := false
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var env WSEventEnvelope
		if err := json.Unmarshal([]byte(data), &env); err == nil && env.V == 1 {
			if env.Type == WSEventActionRequired {
				foundActionRequired = true
				var intent struct {
					Action               string            `json:"action"`
					Params               map[string]string `json:"params"`
					RequiresConfirmation bool              `json:"requires_confirmation"`
					ConfirmationToken    string            `json:"confirmation_token"`
				}
				require.NoError(t, json.Unmarshal(env.Payload, &intent))
				require.Equal(t, "container", intent.Action)
				require.Equal(t, "web", intent.Params["container_id"])
				require.True(t, intent.RequiresConfirmation)
				require.NotEmpty(t, intent.ConfirmationToken)
			}
			continue
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err == nil && chunk.Done {
			foundDone = true
		}
	}
	require.NoError(t, scanner.Err())
	require.True(t, foundActionRequired, "expected action_required event")
	require.True(t, foundDone, "expected done event")
}

func newMockLLMUnknownToolServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_xyz","type":"function","function":{"name":"nonexistent_tool","arguments":"{}"}}]}}]}`,
			`data: {"choices":[{"finish_reason":"tool_calls"}]}`,
			`data: [DONE]`,
		}
		for _, chunk := range chunks {
			fmt.Fprintf(w, "%s\n\n", chunk)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
}

func TestStreamHandler_ToolCall_UnknownTool(t *testing.T) {
	mockLLM := newMockLLMUnknownToolServer(t)
	defer mockLLM.Close()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	store := NewAISettingsStore(mockLLM.URL, "test-key", "test-model")

	reg := action.NewActionRegistry()

	handler := &AssistantHandler{
		settingsStore: store,
		registry:      reg,
		tokenService:  assistant.NewTokenService("", 0),
		l:             &testLogger{},
	}

	app.Post("/stream", handler.Stream)

	body, _ := json.Marshal(map[string]interface{}{
		"messages": []assistant.ChatMessage{{Role: "user", Content: "do something"}},
	})

	req := httptest.NewRequest("POST", "/stream", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	foundError := false
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var env WSEventEnvelope
		if err := json.Unmarshal([]byte(data), &env); err == nil && env.V == 1 {
			if env.Type == WSEventError {
				foundError = true
				var payload WSPayloadError
				require.NoError(t, json.Unmarshal(env.Payload, &payload))
				require.Contains(t, payload.Message, "unknown tool")
				require.Contains(t, payload.Message, "nonexistent_tool")
			}
		}
	}
	require.NoError(t, scanner.Err())
	require.True(t, foundError, "expected error event for unknown tool")
}
