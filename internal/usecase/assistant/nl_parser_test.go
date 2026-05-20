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

type mockTestAction struct {
	name         string
	description  string
	actionParams []action.ParamDef
	destructive  bool
	confirmMsg   string
}

func (m *mockTestAction) Name() string                                           { return m.name }
func (m *mockTestAction) Description() string                                    { return m.description }
func (m *mockTestAction) Params() []action.ParamDef                              { return m.actionParams }
func (m *mockTestAction) Validate(_ map[string]interface{}) error                { return nil }
func (m *mockTestAction) Destructive(_ map[string]interface{}) bool              { return m.destructive }
func (m *mockTestAction) ConfirmationMessage(_ map[string]interface{}) string    { return m.confirmMsg }
func (m *mockTestAction) Execute(_ context.Context, _ map[string]interface{}) (*action.ActionResult, error) {
	return &action.ActionResult{Success: true}, nil
}

// testNLRules returns a set of NL rules for use in parser tests.
// Action names match the real registry pattern: "start_container", "stop_container", etc.
func testNLRules() []engine.Rule {
	return []engine.Rule{
		{
			Name:        "start_container",
			Patterns:    []string{"start nginx container", "start redis container", "start web app"},
			Intent:      "container_start",
			Action:      "start_container",
			Description: "启动容器",
		},
		{
			Name:        "stop_container",
			Patterns:    []string{"stop nginx container", "stop redis container"},
			Intent:      "container_stop",
			Action:      "stop_container",
			Description: "停止容器",
		},
		{
			Name:        "list_containers",
			Patterns:    []string{"list containers", "show containers"},
			Intent:      "container_list",
			Action:      "list_containers",
			Description: "查看容器列表",
		},
		{
			Name:        "delete_image",
			Patterns:    []string{"delete nginx image", "remove redis image"},
			Intent:      "image_delete",
			Action:      "delete_image",
			Description: "删除镜像",
		},
		{
			Name:        "prune_images",
			Patterns:    []string{"prune unused images", "clean up images"},
			Intent:      "image_prune",
			Action:      "prune_images",
			Description: "清理未使用的镜像",
		},
	}
}

// mockOperationAction implements action.Action with operation-based destructiveness.
type mockOperationAction struct {
	name             string
	description      string
	actionParams     []action.ParamDef
	destructiveOps   map[string]bool
}

func (m *mockOperationAction) Name() string                            { return m.name }
func (m *mockOperationAction) Description() string                     { return m.description }
func (m *mockOperationAction) Params() []action.ParamDef               { return m.actionParams }
func (m *mockOperationAction) Validate(_ map[string]interface{}) error { return nil }
func (m *mockOperationAction) Destructive(params map[string]interface{}) bool {
	op, _ := params["operation"].(string)
	return m.destructiveOps[op]
}
func (m *mockOperationAction) ConfirmationMessage(_ map[string]interface{}) string { return "" }
func (m *mockOperationAction) Execute(_ context.Context, _ map[string]interface{}) (*action.ActionResult, error) {
	return &action.ActionResult{Success: true}, nil
}

// newTestRegistry creates an action registry with container and image actions.
func newTestRegistry() *action.ActionRegistry {
	reg := action.NewActionRegistry()
	_ = reg.Register(&mockOperationAction{
		name:        "container",
		description: "Manage Docker containers",
		actionParams: []action.ParamDef{
			{Name: "operation", Type: "string", Required: true, Description: "Container operation"},
			{Name: "container_id", Type: "string", Required: true, Description: "Container ID or name"},
		},
		destructiveOps: map[string]bool{
			"stop_container":      true,
			"restart_container":   true,
			"delete_container":    true,
		},
	})
	_ = reg.Register(&mockOperationAction{
		name:        "image",
		description: "Manage Docker images",
		actionParams: []action.ParamDef{
			{Name: "operation", Type: "string", Required: true, Description: "Image operation"},
			{Name: "machine_id", Type: "string", Required: true, Description: "Machine ID"},
		},
		destructiveOps: map[string]bool{
			"prune_images": true,
			"delete_image": true,
		},
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
			wantAction:        "start_container",
			wantDescription:   "启动容器",
			wantContainerName: "nginx",
		},
		{
			name:              "stop redis container",
			input:             "stop redis container",
			wantIntent:        "container_stop",
			wantAction:        "stop_container",
			wantDescription:   "停止容器",
			wantContainerName: "redis",
		},
		{
			name:              "list containers",
			input:             "list containers",
			wantIntent:        "container_list",
			wantAction:        "list_containers",
			wantDescription:   "查看容器列表",
			wantContainerName: "",
		},
		{
			name:              "partial match - just start",
			input:             "start",
			wantIntent:        "container_start",
			wantAction:        "start_container",
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
	require.Equal(t, "start_container", resp.Action)
	require.Contains(t, resp.Params, "container_name")
	require.Equal(t, "web", resp.Params["container_name"])
}

func TestNLParser_Parse_ShowContainers(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	reg := newTestRegistry()
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})
	uc.SetActionRegistry(reg)

	resp, err := uc.Parse(context.Background(), "show containers")

	require.NoError(t, err)
	require.Equal(t, "container_list", resp.Intent)
	require.Equal(t, "container", resp.Action)
	require.Equal(t, "list_containers", resp.Params["operation"])
}

func TestNLParser_Parse_WithImageName(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	reg := newTestRegistry()
	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})
	uc.SetActionRegistry(reg)

	resp, err := uc.Parse(context.Background(), "delete nginx image")

	require.NoError(t, err)
	require.Equal(t, "image_delete", resp.Intent)
	require.Equal(t, "image", resp.Action)
	require.Contains(t, resp.Params, "container_name")
	require.Equal(t, "nginx", resp.Params["container_name"])
	require.Equal(t, "delete_image", resp.Params["operation"])
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
							Name:      "container",
							Arguments: `{"operation": "start_container", "container_id": "nginx"}`,
						},
					},
				},
			}, nil
		},
	}

	uc := NewNLParser(mockLLM, reg, eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "start nginx")

	require.NoError(t, err)
	require.Equal(t, "container", resp.Intent)
	require.Equal(t, "container", resp.Action)
	require.Contains(t, resp.Description, "Executing: container")
	require.Equal(t, "start_container", resp.Params["operation"])
	require.Equal(t, "nginx", resp.Params["container_id"])
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
	require.Equal(t, "container", resp.Action)
	require.Equal(t, "启动容器", resp.Description)
	require.Equal(t, "nginx", resp.Params["container_name"])
	require.Equal(t, "start_container", resp.Params["operation"])
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
	require.Equal(t, "list_containers", resp.Action)
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
							Name:      "container",
							Arguments: `{"operation": "list_containers"}`,
						},
					},
				},
			}, nil
		},
	}

	uc := NewNLParser(mockLLM, reg, eng, &mockNLTTokenizer{}, &mockLogger{})

	resp, err := uc.Parse(context.Background(), "list all containers")

	require.NoError(t, err)
	require.Equal(t, "container", resp.Intent)
	require.Equal(t, "container", resp.Action)
	require.Contains(t, resp.Description, "Executing: container")
	require.Equal(t, "list_containers", resp.Params["operation"])
}

// ---------- TF-IDF confirmation hardening tests ----------

func newDestructiveTestRegistry() *action.ActionRegistry {
	reg := action.NewActionRegistry()
	_ = reg.Register(&mockOperationAction{
		name:        "container",
		description: "Manage Docker containers",
		actionParams: []action.ParamDef{
			{Name: "operation", Type: "string", Required: true, Description: "Container operation"},
			{Name: "container_id", Type: "string", Required: true, Description: "Container ID or name"},
		},
		destructiveOps: map[string]bool{
			"stop_container":    true,
			"restart_container": true,
			"delete_container":  true,
		},
	})
	_ = reg.Register(&mockOperationAction{
		name:        "image",
		description: "Manage Docker images",
		actionParams: []action.ParamDef{
			{Name: "operation", Type: "string", Required: true, Description: "Image operation"},
			{Name: "machine_id", Type: "string", Required: true, Description: "Machine ID"},
		},
		destructiveOps: map[string]bool{
			"prune_images": true,
			"delete_image": true,
		},
	})
	return reg
}

func TestNLParser_Parse_TFIDF_StopRequiresConfirmation(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	reg := newDestructiveTestRegistry()

	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})
	uc.SetActionRegistry(reg)

	resp, err := uc.Parse(context.Background(), "stop nginx container")

	require.NoError(t, err)
	require.Equal(t, "container_stop", resp.Intent)
	require.Equal(t, "container", resp.Action)
	require.True(t, resp.RequiresConfirmation)
	require.Equal(t, "stop_container", resp.Params["operation"])
}

func TestNLParser_Parse_TFIDF_PruneRequiresConfirmation(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	reg := newDestructiveTestRegistry()

	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})
	uc.SetActionRegistry(reg)

	resp, err := uc.Parse(context.Background(), "prune unused images")

	require.NoError(t, err)
	require.Equal(t, "image_prune", resp.Intent)
	require.Equal(t, "image", resp.Action)
	require.True(t, resp.RequiresConfirmation)
	require.Equal(t, "prune_images", resp.Params["operation"])
}

func TestNLParser_Parse_TFIDF_UnknownDestructiveKeywordFailsSafe(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})

	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})

	tests := []struct {
		name  string
		input string
	}{
		{name: "chinese delete all", input: "删除所有容器"},
		{name: "chinese stop", input: "停止这个容器"},
		{name: "english destroy", input: "destroy all data"},
		{name: "english drop", input: "drop the database"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := uc.Parse(context.Background(), tt.input)

			require.NoError(t, err)
			require.True(t, resp.RequiresConfirmation, "expected RequiresConfirmation=true for destructive input: %s", tt.input)
			require.Equal(t, "该操作需要确认", resp.Description)
		})
	}
}

func TestNLParser_Parse_TFIDF_ReadOnlyStillWorks(t *testing.T) {
	eng := engine.NewEngine(testNLRules(), &mockNLTTokenizer{})
	reg := newDestructiveTestRegistry()

	uc := NewNLParserUseCase(eng, &mockNLTTokenizer{}, &mockLogger{})
	uc.SetActionRegistry(reg)

	resp, err := uc.Parse(context.Background(), "list containers")

	require.NoError(t, err)
	require.Equal(t, "container_list", resp.Intent)
	require.Equal(t, "container", resp.Action)
	require.False(t, resp.RequiresConfirmation)
	require.Equal(t, "list_containers", resp.Params["operation"])
}
