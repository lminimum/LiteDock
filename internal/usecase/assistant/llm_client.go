package assistant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Retry constants
const (
	_maxRetries   = 2
	_retryBackoff = 1 * time.Second
)

// Retryable errors
var (
	_errRetryTimeout = errors.New("timeout")
	_errRetryServer  = errors.New("server error")
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ToolDef struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type ChatResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// LLMClient manages OpenAI-compatible API calls
type LLMClient struct {
	endpoint         string
	apiKey           string
	model            string
	httpClient       *http.Client
	streamHTTPClient *http.Client
}

// NewLLMClient creates a new LLM client
func NewLLMClient(endpoint, apiKey, model string) *LLMClient {
	// Normalize endpoint: strip trailing /v1 or /v1/ to avoid double /v1/ when
	// appending /v1/chat/completions later. Users often configure endpoints
	// like "https://api.openai.com/v1" or "http://localhost:11434/v1".
	endpoint = strings.TrimRight(endpoint, "/")
	endpoint = strings.TrimSuffix(endpoint, "/v1")

	return &LLMClient{
		endpoint:   endpoint,
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		// Streaming SSE responses can take minutes; no body-read timeout.
		streamHTTPClient: &http.Client{},
	}
}

// ChatCompletion sends a chat completion request and returns the response
func (c *LLMClient) ChatCompletion(ctx context.Context, messages []ChatMessage, tools []ToolDef) (*ChatResponse, error) {
	body := map[string]interface{}{
		"model":    c.model,
		"messages": messages,
	}
	if len(tools) > 0 {
		body["tools"] = tools
	}

	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	var lastErr error
	for attempt := 0; attempt <= _maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(_retryBackoff * time.Duration(attempt)):
			}
			req, _ = http.NewRequestWithContext(ctx, "POST", c.endpoint+"/v1/chat/completions", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			if c.apiKey != "" {
				req.Header.Set("Authorization", "Bearer "+c.apiKey)
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if isTimeout(err) {
				lastErr = fmt.Errorf("%w: %v", _errRetryTimeout, err)
				continue
			}
			return nil, fmt.Errorf("send request: %w", err)
		}

		if resp.StatusCode == 0 {
			resp.Body.Close()
			lastErr = fmt.Errorf("%w: connection closed", _errRetryTimeout)
			continue
		}

		if resp.StatusCode >= 500 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("%w: status %d: %s", _errRetryServer, resp.StatusCode, string(bodyBytes))
			continue
		}

		if resp.StatusCode != 200 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
		}

		var openAIResp struct {
			Choices []struct {
				Message struct {
					Content   string     `json:"content"`
					ToolCalls []ToolCall `json:"tool_calls"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode response: %w", err)
		}

		resp.Body.Close()

		if len(openAIResp.Choices) == 0 {
			return &ChatResponse{}, nil
		}

		msg := openAIResp.Choices[0].Message
		return &ChatResponse{
			Content:   msg.Content,
			ToolCalls: msg.ToolCalls,
		}, nil
	}

	return nil, lastErr
}

// StreamChatCompletion sends a streaming chat completion request and returns the raw response.
// The caller is responsible for reading the SSE response body and closing it.
func (c *LLMClient) StreamChatCompletion(ctx context.Context, messages []ChatMessage, tools []ToolDef) (*http.Response, error) {
	body := map[string]interface{}{
		"model":    c.model,
		"messages": messages,
		"stream":   true,
	}
	if len(tools) > 0 {
		body["tools"] = tools
	}

	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.streamHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline exceeded") ||
		strings.Contains(err.Error(), "Client.Timeout") {
		return true
	}
	return false
}

// CloseIdleConnections cleans up idle HTTP connections
func (c *LLMClient) CloseIdleConnections() {
	c.httpClient.CloseIdleConnections()
	c.streamHTTPClient.CloseIdleConnections()
}
