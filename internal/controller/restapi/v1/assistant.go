package v1

import (
	"bufio"
	"context"
	"encoding/json"
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
		l:             l,
	}

	assistantGroup := apiV1Group.Group("/assistant")
	assistantGroup.Post("/parse", h.Parse)
	assistantGroup.Post("/diagnose", h.Diagnose)
	assistantGroup.Post("/recommend", h.Recommend)
	assistantGroup.Post("/stream", h.Stream)
	assistantGroup.Get("/stream/ws", websocket.New(h.StreamWS))
	assistantGroup.Post("/execute", h.ExecuteAction)
}

// streamChunk is the SSE event payload sent to the client.
type streamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

// StreamRequest is the request body for POST /v1/assistant/stream.
type StreamRequest struct {
	Messages []assistant.ChatMessage `json:"messages"`
	Tools    []assistant.ToolDef     `json:"tools,omitempty"`
}

// ExecuteRequest represents a request to execute a confirmed action.
type ExecuteRequest struct {
	Action string            `json:"action"`
	Params map[string]string `json:"params"`
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

	resp, err := client.StreamChatCompletion(c.Context(), req.Messages, req.Tools)
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

		for scanner.Scan() {
			line := strings.TrimSuffix(scanner.Text(), "\r")
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				chunk, _ := json.Marshal(streamChunk{Content: "", Done: true})
				fmt.Fprintf(w, "data: %s\n\n", chunk)
				w.Flush()
				return
			}

			var openAIChunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(data), &openAIChunk); err != nil {
				continue
			}

			if len(openAIChunk.Choices) == 0 {
				continue
			}

			content := openAIChunk.Choices[0].Delta.Content
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

		// Ensure done event is sent even on scanner error
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

		if h.rateLimiter != nil && !h.rateLimiter.Allow("ws:"+c.RemoteAddr().String()) {
			c.WriteJSON(map[string]interface{}{"error": assistant.MapErrorToUserMessage(assistant.ErrRateLimited)})
			continue
		}

		var req struct {
			Messages []assistant.ChatMessage `json:"messages"`
			Tools    []assistant.ToolDef     `json:"tools,omitempty"`
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

		resp, err := client.StreamChatCompletion(ctx, req.Messages, req.Tools)
		if err != nil {
			c.WriteJSON(map[string]interface{}{"error": "LLM request failed: " + assistant.MapErrorToUserMessage(err)})
			h.l.Error(err, "AssistantHandler - StreamWS - StreamChatCompletion")
			continue
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 65536), 65536)

		doneSent := false
		for scanner.Scan() {
			line := strings.TrimSuffix(scanner.Text(), "\r")
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				c.WriteJSON(map[string]interface{}{"done": true})
				doneSent = true
				break
			}

			var openAIChunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(data), &openAIChunk); err != nil {
				continue
			}
			if len(openAIChunk.Choices) > 0 && openAIChunk.Choices[0].Delta.Content != "" {
				content := openAIChunk.Choices[0].Delta.Content
				c.WriteJSON(map[string]interface{}{"content": content, "done": false})
			}
		}

		resp.Body.Close()

		if scanner.Err() != nil {
			h.l.Error(scanner.Err(), "AssistantHandler - StreamWS - scanner")
		}

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
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Action == "" {
		return errorResponse(c, fiber.StatusBadRequest, "action is required")
	}

	// Convert string params to interface{} for the action system
	params := make(map[string]interface{}, len(req.Params))
	for k, v := range req.Params {
		params[k] = v
	}

	result, err := h.registry.ExecuteConfirmed(c.Context(), req.Action, params)
	if err != nil {
		h.l.Error(fmt.Errorf("AssistantHandler - ExecuteAction - registry.ExecuteConfirmed: %w", err), "")
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to execute action: "+err.Error())
	}

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
