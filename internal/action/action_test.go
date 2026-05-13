package action

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

type mockAction struct {
	name           string
	description    string
	params         []ParamDef
	validateFn    func(map[string]interface{}) error
	destructiveFn func(map[string]interface{}) bool
	confirmMsgFn  func(map[string]interface{}) string
	executeFn     func(context.Context, map[string]interface{}) (*ActionResult, error)
}

func (m *mockAction) Name() string                                       { return m.name }
func (m *mockAction) Description() string                                { return m.description }
func (m *mockAction) Params() []ParamDef                                 { return m.params }
func (m *mockAction) Validate(params map[string]interface{}) error       { return m.validateFn(params) }
func (m *mockAction) Destructive(params map[string]interface{}) bool {
	if m.destructiveFn != nil {
		return m.destructiveFn(params)
	}
	return false
}
func (m *mockAction) ConfirmationMessage(params map[string]interface{}) string {
	if m.confirmMsgFn != nil {
		return m.confirmMsgFn(params)
	}
	return ""
}
func (m *mockAction) Execute(ctx context.Context, params map[string]interface{}) (*ActionResult, error) {
	return m.executeFn(ctx, params)
}

func TestRegisterAndGet(t *testing.T) {
	reg := NewActionRegistry()
	a := &mockAction{
		name:        "test_action",
		description: "A test action",
	}

	err := reg.Register(a)
	if err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	got, ok := reg.Get("test_action")
	if !ok {
		t.Fatal("Get() returned ok=false")
	}
	if got.Name() != "test_action" {
		t.Errorf("Get().Name() = %s, want %s", got.Name(), "test_action")
	}
}

func TestList(t *testing.T) {
	reg := NewActionRegistry()
	actions := []*mockAction{
		{name: "action1"},
		{name: "action2"},
		{name: "action3"},
	}

	for _, a := range actions {
		if err := reg.Register(a); err != nil {
			t.Fatalf("Register(%s) returned error: %v", a.name, err)
		}
	}

	listed := reg.List()
	if len(listed) != 3 {
		t.Fatalf("List() returned %d actions, want 3", len(listed))
	}

	names := make(map[string]bool)
	for _, a := range listed {
		names[a.Name()] = true
	}
	for _, a := range actions {
		if !names[a.name] {
			t.Errorf("List() missing action: %s", a.name)
		}
	}
}

func TestExecuteSuccess(t *testing.T) {
	reg := NewActionRegistry()
	a := &mockAction{
		name: "echo",
		validateFn: func(params map[string]interface{}) error {
			return nil
		},
		executeFn: func(ctx context.Context, params map[string]interface{}) (*ActionResult, error) {
			return &ActionResult{
				Success: true,
				Message: "executed",
				Data:    params,
			}, nil
		},
	}

	if err := reg.Register(a); err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	params := map[string]interface{}{"msg": "hello"}
	result, err := reg.Execute(context.Background(), "echo", params)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	if !result.Success {
		t.Error("Execute().Success = false, want true")
	}
	if result.Duration == "" {
		t.Error("Execute().Duration should not be empty")
	}
}

func TestExecuteUnknownAction(t *testing.T) {
	reg := NewActionRegistry()
	_, err := reg.Execute(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Fatal("Execute() expected error for unknown action")
	}
	if !errors.Is(err, ErrUnknownAction) {
		t.Errorf("Execute() error should wrap ErrUnknownAction, got: %v", err)
	}
}

func TestExecuteValidationFailure(t *testing.T) {
	reg := NewActionRegistry()
	a := &mockAction{
		name: "validated",
		validateFn: func(params map[string]interface{}) error {
			return fmt.Errorf("missing required field: name")
		},
		executeFn: func(ctx context.Context, params map[string]interface{}) (*ActionResult, error) {
			return &ActionResult{Success: true}, nil
		},
	}

	if err := reg.Register(a); err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	_, err := reg.Execute(context.Background(), "validated", nil)
	if err == nil {
		t.Fatal("Execute() expected error for validation failure")
	}
	if !errors.Is(err, ErrValidationFailed) {
		t.Errorf("Execute() error should wrap ErrValidationFailed, got: %v", err)
	}
}

func TestDuplicateRegistration(t *testing.T) {
	reg := NewActionRegistry()
	a := &mockAction{name: "dup"}

	if err := reg.Register(a); err != nil {
		t.Fatalf("First Register() returned error: %v", err)
	}

	err := reg.Register(a)
	if err == nil {
		t.Fatal("Second Register() expected error")
	}
	if !errors.Is(err, ErrActionAlreadyRegistered) {
		t.Errorf("Register() error should wrap ErrActionAlreadyRegistered, got: %v", err)
	}
}

func TestGenerateToolDefs(t *testing.T) {
	reg := NewActionRegistry()
	a := &mockAction{
		name:        "create_container",
		description: "Creates a new container",
		params: []ParamDef{
			{Name: "image", Type: "string", Required: true, Description: "Docker image name"},
			{Name: "name", Type: "string", Required: false, Description: "Container name"},
		},
	}

	if err := reg.Register(a); err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	tools := reg.GenerateToolDefs()
	if len(tools) != 1 {
		t.Fatalf("GenerateToolDefs() returned %d tools, want 1", len(tools))
	}

	tool := tools[0]
	if tool["type"] != "function" {
		t.Errorf("tool type = %v, want function", tool["type"])
	}

	fn, ok := tool["function"].(map[string]interface{})
	if !ok {
		t.Fatal("tool['function'] is not a map")
	}
	if fn["name"] != "create_container" {
		t.Errorf("function name = %v, want create_container", fn["name"])
	}
	if fn["description"] != "Creates a new container" {
		t.Errorf("function description = %v, want Creates a new container", fn["description"])
	}

	params, ok := fn["parameters"].(map[string]interface{})
	if !ok {
		t.Fatal("function['parameters'] is not a map")
	}
	if params["type"] != "object" {
		t.Errorf("parameters type = %v, want object", params["type"])
	}

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("parameters['properties'] is not a map")
	}
	if _, exists := props["image"]; !exists {
		t.Error("properties missing 'image'")
	}
	if _, exists := props["name"]; !exists {
		t.Error("properties missing 'name'")
	}

	req, ok := params["required"].([]string)
	if !ok {
		t.Fatal("parameters['required'] is not a []string")
	}
	if len(req) != 1 || req[0] != "image" {
		t.Errorf("required = %v, want [image]", req)
	}
}

func TestGenerateToolDefsEmpty(t *testing.T) {
	reg := NewActionRegistry()
	tools := reg.GenerateToolDefs()
	if len(tools) != 0 {
		t.Errorf("GenerateToolDefs() on empty registry = %d tools, want 0", len(tools))
	}
}

func TestConcurrentAccessSafety(t *testing.T) {
	reg := NewActionRegistry()

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a := &mockAction{
				name: fmt.Sprintf("concurrent_%d", i),
				validateFn: func(params map[string]interface{}) error {
					return nil
				},
				executeFn: func(ctx context.Context, params map[string]interface{}) (*ActionResult, error) {
					return &ActionResult{Success: true, Message: "ok"}, nil
				},
			}
			_ = reg.Register(a)
		}(i)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = reg.List()
			_, _ = reg.Get("concurrent_0")
		}()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = reg.Execute(context.Background(), "concurrent_0", nil)
		}()
	}

	wg.Wait()

	listed := reg.List()
	if len(listed) != 10 {
		t.Errorf("After concurrent ops, List() returned %d actions, want 10", len(listed))
	}
}

func TestGenerateToolDefsMultipleActions(t *testing.T) {
	reg := NewActionRegistry()

	actions := []*mockAction{
		{
			name:        "action_a",
			description: "First action",
			params: []ParamDef{
				{Name: "x", Type: "integer", Required: true, Description: "X value"},
			},
		},
		{
			name:        "action_b",
			description: "Second action",
			params: []ParamDef{
				{Name: "y", Type: "string", Required: false, Description: "Y value"},
			},
		},
	}

	for _, a := range actions {
		if err := reg.Register(a); err != nil {
			t.Fatalf("Register(%s) returned error: %v", a.name, err)
		}
	}

	tools := reg.GenerateToolDefs()
	if len(tools) != 2 {
		t.Fatalf("GenerateToolDefs() returned %d tools, want 2", len(tools))
	}
}

func TestGetNonExistent(t *testing.T) {
	reg := NewActionRegistry()
	_, ok := reg.Get("does_not_exist")
	if ok {
		t.Error("Get() for non-existent action returned ok=true")
	}
}

func TestNewActionRegistry(t *testing.T) {
	reg := NewActionRegistry()
	if reg == nil {
		t.Fatal("NewActionRegistry() returned nil")
	}
	if len(reg.actions) != 0 {
		t.Errorf("NewActionRegistry().actions has %d entries, want 0", len(reg.actions))
	}
}
