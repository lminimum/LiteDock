package v1

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
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
