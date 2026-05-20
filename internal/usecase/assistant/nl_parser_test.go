package assistant

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/lminimum/LiteDock/internal/action"
	"github.com/lminimum/LiteDock/pkg/assistant/engine"
	"github.com/stretchr/testify/require"
)

// mockNLTTokenizer splits on whitespace and lowercases tokens.
type mockNLTTokenizer struct{}

func (m *mockNLTTokenizer) Tokenize(input string) ([]string, error) {
	if strings.TrimSpace(input) == "" {
		return nil, nil
	}
	return strings.Fields(strings.ToLower(input)), nil
}

// mockLLMClient implements LLMClientInterface for testing.
type mockLLMClient struct {
	chatCompletionFn func(ctx context.Context, messages []ChatMessage, tools []ToolDef) (*ChatResponse, error)
}

func (m *mockLLMClient) ChatCompletion(ctx context.Context, messages []ChatMessage, tools []ToolDef) (*ChatResponse, error) {
	return m.chatCompletionFn(ctx, messages, tools)
}

// mockTestAction implements action.Action for testing.
type mockTestAction struct {
	name        string
	description string
	actionParams []action.ParamDef
}

func (m *mockTestAction) Name() string                            { return m.name }
func (m *mockTestAction) Description() string                     { return m.description }
func (m *mockTestAction) Params() []action.ParamDef               { return m.actionParams }
func (m *mockTestAction) Validate(_ map[string]interface{}) error { return nil }
func (m *mockTestAction) Destructive(_ map[string]interface{}) bool          { return false }
func (m *mockTestAction) ConfirmationMessage(_ map[string]interface{}) string { return "" }
func (m *mockTestAction) Execute(_ context.Context, _ map[string]interface{}) (*action.ActionResult, error) {
	return &action.ActionResult{Success: true}, nil
}

// testNLRules returns a set of NL rules for use in parser tests.
func testNLRules() []engine.Rule {
	return []engine.Rule{
		{
			Name:        "start_container",
			Patterns:    []string{"start nginx container", "start redis container", "start web app"},
			Intent:      "container_start",
			Action:      "start",
			Description: "启动容器",
		},
		{
			Name:        "stop_container",
			Patterns:    []string{"stop nginx container", "stop redis container"},
			Intent:      "container_stop",
			Action:      "stop",
			Description: "停止容器",
		},
		{
			Name:        "list_containers",
			Patterns:    []string{"list containers", "show containers"},
			Intent:      "container_list",
			Action:      "list",
			Description: "查看容器列表",
		},
		{
			Name:        "delete_image",
			Patterns:    []string{"delete nginx image", "remove redis image"},
			Intent:      "image_delete",
			Action:      "delete",
			Description: "删除镜像",
		},
	}
}

// newTestRegistry creates an action registry with a test action for LLM tests.
func newTestRegistry() *action.ActionRegistry {
	reg := action.NewActionRegistry()
	_ = reg.Register(&mockTestAction{
		name:        "start_container",
		description: "Starts a Docker container",
		actionParams: []action.ParamDef{
			{Name: "name", Type: "string", Required: true, Description: "Container name"},
		},
	})
	_ = reg.Register(&mockTestAction{
		name:        "list_containers",
		description: "Lists all Docker containers",
	})
	return reg
}

// ---------- TF-IDF fallback tests (LLM unavailable) ----------

func TestNLParser_Parse_HappyPath(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		wantIntent        string
		wantAction        string
		wantDescription   string
		wantContainerName string
	}{
		{
			name:              "start nginx container",
			input:             "start nginx container",
			wantIntent:        "container_start",
			wantAction:        "start",
			wantDescription:   "启动容器",
			wantContainerName: "nginx",
		},
		{
			name:              "stop redis container",
			input:             "stop redis container",
			wantIntent:        "container_stop",
			wantAction:        "stop",
			wantDescription:   "停止容器",
			wantContainerName: "redis",
		},
		{
			name:              "list containers",
			input:             "list containers",
			wantIntent:        "container_list",
			wantAction:        "list",
			wantDescription:   "查看容器列表",
			wantContainerName: "",
		},
		{
			name:              "partial match - just start",
			input:             "start",
			wantIntent:        "container_start",
			wantAction:        "start",
			wantDescription:   "启动容器",
			wantContainerName: "",
		},
	}

	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := uc.Parse(context.Background(), tt.input)

			require.NoError(t, err)
			require.Equal(t, tt.wantIntent, resp.Intent)
			require.Equal(t, tt.wantAction, resp.Action)
			require.Equal(t, tt.wantDescription, resp.Description)
			require.NotNil(t, resp.Params)

			if tt.wantContainerName != "" {
				require.Contains(t, resp.Params, "container_name")
				require.Equal(t, tt.wantContainerName, resp.Params["container_name"])
			} else {
				_, hasContainerName := resp.Params["container_name"]
				require.False(t, hasContainerName)
			}
		})
	}
}

func TestNLParser_Parse_UnknownInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "completely unrelated", input: "do something completely unrelated now"},
		{name: "random tokens", input: "xylophone zephyr quantum"},
		{name: "single unknown word", input: "python"},
	}

	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := uc.Parse(context.Background(), tt.input)

			require.NoError(t, err)
			require.Equal(t, "unknown", resp.Intent)
			require.Equal(t, "未识别您的指令", resp.Description)
			require.NotNil(t, resp.Params)
		})
	}
}

func TestNLParser_Parse_EmptyInput(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "")

	require.Error(t, err)
	require.Contains(t, err.Error(), "请输入指令")
	require.Empty(t, resp.Intent)
	require.Empty(t, resp.Action)
}

func TestNLParser_Parse_WhitespaceOnly(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})

	tests := []struct {
		name  string
		input string
	}{
		{"spaces", "   "},
		{"tabs", "\t\t\t"},
		{"newlines and spaces", "\n \t \n  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := uc.Parse(context.Background(), tt.input)

			require.Error(t, err)
			require.Contains(t, err.Error(), "请输入指令")
			require.Empty(t, resp.Intent)
		})
	}
}

func TestNLParser_Parse_ExceedsLength(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})

	longInput := strings.Repeat("a", 501)

	resp, err := uc.Parse(context.Background(), longInput)

	require.Error(t, err)
	require.Contains(t, err.Error(), "500")
	require.Empty(t, resp.Intent)
}

func TestNLParser_Parse_AtMaxLength(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})

	atLimit := strings.Repeat("a", 500)

	resp, err := uc.Parse(context.Background(), atLimit)

	require.NoError(t, err)
	require.Equal(t, "unknown", resp.Intent)
	require.Equal(t, "未识别您的指令", resp.Description)
}

func TestNLParser_Parse_StartWebApp(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "start web app")

	require.NoError(t, err)
	require.Equal(t, "container_start", resp.Intent)
	require.Equal(t, "start", resp.Action)
	require.Contains(t, resp.Params, "container_name")
	require.Equal(t, "web", resp.Params["container_name"])
}

func TestNLParser_Parse_ShowContainers(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "show containers")

	require.NoError(t, err)
	require.Equal(t, "container_list", resp.Intent)
	require.Equal(t, "list", resp.Action)
	require.Empty(t, resp.Params)
}

func TestNLParser_Parse_WithImageName(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "delete nginx image")

	require.NoError(t, err)
	require.Equal(t, "image_delete", resp.Intent)
	require.Equal(t, "delete", resp.Action)
	require.Contains(t, resp.Params, "container_name")
	require.Equal(t, "nginx", resp.Params["container_name"])
}

func TestNLParser_Parse_MultipleStopWords(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "start nginx container")

	require.NoError(t, err)
	require.Equal(t, "container_start", resp.Intent)
	require.Equal(t, "nginx", resp.Params["container_name"])
	require.NotContains(t, resp.Params, "container")
}

// ---------- LLM path tests ----------

func TestNLParser_Parse_LLM_ToolCall(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	reg := newTestRegistry()

	mockLLM := &mockLLMClient{
		chatCompletionFn: func(_ context.Context, _ []ChatMessage, _ []ToolDef) (*ChatResponse, error) {
			return &ChatResponse{
				ToolCalls: []ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{
							Name:      "start_container",
							Arguments: `{"name": "nginx"}`,
						},
					},
				},
			}, nil
		},
	}

	uc := NewNLParser(mockLLM, reg, eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "start nginx")

	require.NoError(t, err)
	require.Equal(t, "start_container", resp.Intent)
	require.Equal(t, "start_container", resp.Action)
	require.Contains(t, resp.Description, "Executing: start_container")
	require.Equal(t, "nginx", resp.Params["name"])
}

func TestNLParser_Parse_LLM_PlainText(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	reg := newTestRegistry()

	mockLLM := &mockLLMClient{
		chatCompletionFn: func(_ context.Context, _ []ChatMessage, _ []ToolDef) (*ChatResponse, error) {
			return &ChatResponse{
				Content: "I can help you with that. What would you like to do?",
			}, nil
		},
	}

	uc := NewNLParser(mockLLM, reg, eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "hello")

	require.NoError(t, err)
	require.Equal(t, "chat", resp.Intent)
	require.Empty(t, resp.Action)
	require.Equal(t, "I can help you with that. What would you like to do?", resp.Description)
	require.NotNil(t, resp.Params)
}

func TestNLParser_Parse_LLM_Error_FallbackToTFIDF(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	reg := newTestRegistry()

	mockLLM := &mockLLMClient{
		chatCompletionFn: func(_ context.Context, _ []ChatMessage, _ []ToolDef) (*ChatResponse, error) {
			return nil, fmt.Errorf("LLM unavailable")
		},
	}

	uc := NewNLParser(mockLLM, reg, eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "start nginx container")

	require.NoError(t, err)
	require.Equal(t, "container_start", resp.Intent)
	require.Equal(t, "start", resp.Action)
	require.Equal(t, "启动容器", resp.Description)
	require.Equal(t, "nginx", resp.Params["container_name"])
}

func TestNLParser_Parse_LLM_Error_FallbackToUnknown(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	reg := newTestRegistry()

	mockLLM := &mockLLMClient{
		chatCompletionFn: func(_ context.Context, _ []ChatMessage, _ []ToolDef) (*ChatResponse, error) {
			return nil, fmt.Errorf("LLM unavailable")
		},
	}

	uc := NewNLParser(mockLLM, reg, eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "xylophone zephyr")

	require.NoError(t, err)
	require.Equal(t, "unknown", resp.Intent)
	require.Equal(t, "未识别您的指令", resp.Description)
}

func TestNLParser_Parse_LLM_NoClient_FallbackToTFIDF(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "list containers")

	require.NoError(t, err)
	require.Equal(t, "container_list", resp.Intent)
	require.Equal(t, "list", resp.Action)
	require.Equal(t, "查看容器列表", resp.Description)
}

func TestNLParser_Parse_LLM_ToolCall_EmptyArgs(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	reg := newTestRegistry()

	mockLLM := &mockLLMClient{
		chatCompletionFn: func(_ context.Context, _ []ChatMessage, _ []ToolDef) (*ChatResponse, error) {
			return &ChatResponse{
				ToolCalls: []ToolCall{
					{
						ID:   "call_2",
						Type: "function",
						Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{
							Name:      "list_containers",
							Arguments: `{}`,
						},
					},
				},
			}, nil
		},
	}

	uc := NewNLParser(mockLLM, reg, eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "list all containers")

	require.NoError(t, err)
	require.Equal(t, "list_containers", resp.Intent)
	require.Equal(t, "list_containers", resp.Action)
	require.Contains(t, resp.Description, "Executing: list_containers")
	require.Empty(t, resp.Params)
}
