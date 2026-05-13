package assistant

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newMockLLMServer(t *testing.T, statusCode int, responseBody interface{}, apiKey string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey != "" {
			require.Equal(t, "Bearer "+apiKey, r.Header.Get("Authorization"),
				"request must carry the expected API key")
		}
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.Equal(t, "/v1/chat/completions", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if responseBody != nil {
			err := json.NewEncoder(w).Encode(responseBody)
			require.NoError(t, err)
		}
	}))
}

func newLLMClientForTest(serverURL string) *LLMClient {
	return NewLLMClient(serverURL, "test-key", "test-model")
}

func TestChatCompletion_TextResponse(t *testing.T) {
	respBody := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"content": "Hello! How can I help you?",
				},
			},
		},
	}
	server := newMockLLMServer(t, http.StatusOK, respBody, "test-key")
	defer server.Close()

	client := newLLMClientForTest(server.URL)
	resp, err := client.ChatCompletion(context.Background(),
		[]ChatMessage{{Role: "user", Content: "Hi"}}, nil)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "Hello! How can I help you?", resp.Content)
	require.Empty(t, resp.ToolCalls)
}

func TestChatCompletion_ToolCallResponse(t *testing.T) {
	respBody := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"content": "",
					"tool_calls": []map[string]interface{}{
						{
							"id":   "call_123",
							"type": "function",
							"function": map[string]interface{}{
								"name":      "get_weather",
								"arguments": `{"location": "Shanghai"}`,
							},
						},
					},
				},
			},
		},
	}
	server := newMockLLMServer(t, http.StatusOK, respBody, "test-key")
	defer server.Close()

	client := newLLMClientForTest(server.URL)
	resp, err := client.ChatCompletion(context.Background(), []ChatMessage{
		{Role: "user", Content: "What's the weather in Shanghai?"},
	}, nil)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Content)
	require.Len(t, resp.ToolCalls, 1)
	require.Equal(t, "call_123", resp.ToolCalls[0].ID)
	require.Equal(t, "function", resp.ToolCalls[0].Type)
	require.Equal(t, "get_weather", resp.ToolCalls[0].Function.Name)
	require.Equal(t, `{"location": "Shanghai"}`, resp.ToolCalls[0].Function.Arguments)
}

func TestChatCompletion_AuthenticationError(t *testing.T) {
	errBody := map[string]interface{}{
		"error": map[string]interface{}{
			"message": "Invalid API key",
			"type":    "authentication_error",
		},
	}
	server := newMockLLMServer(t, http.StatusUnauthorized, errBody, "test-key")
	defer server.Close()

	client := newLLMClientForTest(server.URL)
	resp, err := client.ChatCompletion(context.Background(),
		[]ChatMessage{{Role: "user", Content: "Hi"}}, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "API error (status 401)")
	require.Contains(t, err.Error(), "Invalid API key")
	require.Nil(t, resp)
}

func TestChatCompletion_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newLLMClientForTest(server.URL)
	client.httpClient = &http.Client{Timeout: 10 * time.Millisecond}

	resp, err := client.ChatCompletion(context.Background(),
		[]ChatMessage{{Role: "user", Content: "Hi"}}, nil)

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestChatCompletion_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newLLMClientForTest(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resp, err := client.ChatCompletion(ctx,
		[]ChatMessage{{Role: "user", Content: "Hi"}}, nil)

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestChatCompletion_EmptyResponse(t *testing.T) {
	respBody := map[string]interface{}{
		"choices": []interface{}{},
	}
	server := newMockLLMServer(t, http.StatusOK, respBody, "test-key")
	defer server.Close()

	client := newLLMClientForTest(server.URL)
	resp, err := client.ChatCompletion(context.Background(),
		[]ChatMessage{{Role: "user", Content: "Hi"}}, nil)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Content)
	require.Empty(t, resp.ToolCalls)
}

func TestChatCompletion_WithTools(t *testing.T) {
	tools := []ToolDef{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_weather",
				Description: "Get weather for a location",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]string{"type": "string"},
					},
				},
			},
		},
	}

	respBody := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"content": "Let me check the weather.",
					"tool_calls": []map[string]interface{}{
						{
							"id":   "call_456",
							"type": "function",
							"function": map[string]interface{}{
								"name":      "get_weather",
								"arguments": `{"location": "Beijing"}`,
							},
						},
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.Empty(t, r.Header.Get("Authorization"))
		require.Equal(t, "/v1/chat/completions", r.URL.Path)

		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		require.Contains(t, reqBody, "tools")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(respBody)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewLLMClient(server.URL, "", "test-model")
	resp, err := client.ChatCompletion(context.Background(),
		[]ChatMessage{{Role: "user", Content: "What's the weather in Beijing?"}},
		tools)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "Let me check the weather.", resp.Content)
	require.Len(t, resp.ToolCalls, 1)
	require.Equal(t, "call_456", resp.ToolCalls[0].ID)
	require.Equal(t, "get_weather", resp.ToolCalls[0].Function.Name)
}

func TestNewLLMClient_Defaults(t *testing.T) {
	client := NewLLMClient("https://api.example.com", "my-key", "gpt-4")

	require.NotNil(t, client)
	require.Equal(t, "https://api.example.com", client.endpoint)
	require.Equal(t, "my-key", client.apiKey)
	require.Equal(t, "gpt-4", client.model)
	require.NotNil(t, client.httpClient)
	require.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestCloseIdleConnections(t *testing.T) {
	server := newMockLLMServer(t, http.StatusOK, map[string]interface{}{
		"choices": []map[string]interface{}{
			{"message": map[string]interface{}{"content": "ok"}},
		},
	}, "")
	defer server.Close()

	client := newLLMClientForTest(server.URL)
	_, err := client.ChatCompletion(context.Background(),
		[]ChatMessage{{Role: "user", Content: "Hi"}}, nil)
	require.NoError(t, err)

	client.CloseIdleConnections()
}

func TestStreamChatCompletion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		require.Equal(t, "/v1/chat/completions", r.URL.Path)

		// Verify stream=true is in request
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		require.Equal(t, true, reqBody["stream"])

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`data: {"choices":[{"delta":{"content":"Hel"}}]}`,
			`data: {"choices":[{"delta":{"content":"lo"}}]}`,
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
	defer server.Close()

	client := newLLMClientForTest(server.URL)
	resp, err := client.StreamChatCompletion(context.Background(),
		[]ChatMessage{{Role: "user", Content: "Hi"}}, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	require.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	// Read all streaming chunks
	var fullContent string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		require.NoError(t, json.Unmarshal([]byte(data), &chunk))
		if len(chunk.Choices) > 0 {
			fullContent += chunk.Choices[0].Delta.Content
		}
	}
	require.NoError(t, scanner.Err())
	require.Equal(t, "Hello!", fullContent)
}

func TestStreamChatCompletion_Error(t *testing.T) {
	errBody := map[string]interface{}{
		"error": map[string]interface{}{"message": "model not found"},
	}
	server := newMockLLMServer(t, http.StatusNotFound, errBody, "test-key")
	defer server.Close()

	client := newLLMClientForTest(server.URL)
	resp, err := client.StreamChatCompletion(context.Background(),
		[]ChatMessage{{Role: "user", Content: "Hi"}}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "API error (status 404)")
	require.Nil(t, resp)
}

func TestStreamChatCompletion_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	client := newLLMClientForTest(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resp, err := client.StreamChatCompletion(ctx,
		[]ChatMessage{{Role: "user", Content: "Hi"}}, nil)
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestStreamChatCompletion_WithTools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		require.Contains(t, reqBody, "tools")
		require.Equal(t, true, reqBody["stream"])

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n")
		fmt.Fprintf(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := newLLMClientForTest(server.URL)
	tools := []ToolDef{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_weather",
				Description: "Get weather",
			},
		},
	}
	resp, err := client.StreamChatCompletion(context.Background(),
		[]ChatMessage{{Role: "user", Content: "weather?"}}, tools)
	require.NoError(t, err)
	require.NotNil(t, resp)
	resp.Body.Close()
}
