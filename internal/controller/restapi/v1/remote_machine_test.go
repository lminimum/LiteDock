package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/entity"
)

type mockRemoteMachineUseCase struct {
	machines map[string]*entity.RemoteMachine
}

type mockLogger struct{}

func (m *mockLogger) Error(message interface{}, args ...interface{})     {}
func (m *mockLogger) Info(msg string, args ...interface{})              {}
func (m *mockLogger) Debug(msg interface{}, args ...interface{})        {}
func (m *mockLogger) Warn(msg string, args ...interface{})                {}
func (m *mockLogger) Fatal(message interface{}, args ...interface{})     {}
func (m *mockLogger) With(key, value string) interface{}               { return m }

func (m *mockRemoteMachineUseCase) Create(ctx context.Context, machine *entity.RemoteMachine) (*entity.RemoteMachine, error) {
	if machine.ID == "" {
		machine.ID = "generated-id"
	}
	m.machines[machine.ID] = machine
	return machine, nil
}

func (m *mockRemoteMachineUseCase) GetByID(ctx context.Context, id string) (*entity.RemoteMachine, error) {
	if v, ok := m.machines[id]; ok {
		return v, nil
	}
	return nil, errors.New("not found")
}

func (m *mockRemoteMachineUseCase) List(ctx context.Context) ([]entity.RemoteMachine, error) {
	result := make([]entity.RemoteMachine, 0, len(m.machines))
	for _, v := range m.machines {
		result = append(result, *v)
	}
	return result, nil
}

func (m *mockRemoteMachineUseCase) Update(ctx context.Context, machine *entity.RemoteMachine) error {
	m.machines[machine.ID] = machine
	return nil
}

func (m *mockRemoteMachineUseCase) Delete(ctx context.Context, id string) error {
	delete(m.machines, id)
	return nil
}

func (m *mockRemoteMachineUseCase) GetByHost(ctx context.Context, host string) (*entity.RemoteMachine, error) {
	return nil, errors.New("not found")
}

func (m *mockRemoteMachineUseCase) TestConnection(ctx context.Context, id string) error {
	return nil
}

func (m *mockRemoteMachineUseCase) ListContainers(ctx context.Context, machineID string) ([]entity.Container, error) {
	return []entity.Container{}, nil
}

func (m *mockRemoteMachineUseCase) GetContainerLogs(ctx context.Context, machineID, containerID, tail string) (string, error) {
	return "log output", nil
}

func (m *mockRemoteMachineUseCase) ExecContainer(ctx context.Context, machineID, containerID string, cmd []string) (string, error) {
	return "exec output", nil
}

func (m *mockRemoteMachineUseCase) StartContainer(ctx context.Context, machineID, containerID string) error {
	return nil
}

func (m *mockRemoteMachineUseCase) StopContainer(ctx context.Context, machineID, containerID string) error {
	return nil
}

func (m *mockRemoteMachineUseCase) RestartContainer(ctx context.Context, machineID, containerID string) error {
	return nil
}

func (m *mockRemoteMachineUseCase) RemoveContainer(ctx context.Context, machineID, containerID string, force bool) error {
	return nil
}

func (m *mockRemoteMachineUseCase) InspectContainer(ctx context.Context, machineID, containerID string) (interface{}, error) {
	return nil, nil
}

func TestRemoteMachineHandlerCreate(t *testing.T) {
	app := fiber.New()
	repo := &mockRemoteMachineUseCase{machines: make(map[string]*entity.RemoteMachine)}
	handler := &RemoteMachineHandler{uc: repo, l: &mockLogger{}, v: validator.New(validator.WithRequiredStructEnabled())}

	app.Post("/machines", handler.Create)

	body := `{"name":"test-server","host":"192.168.1.100","port":22,"username":"root","auth_method":"password","password":"secret"}`
	req := httptest.NewRequest("POST", "/machines", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if result["success"] != true {
		t.Error("expected success=true")
	}
}

func TestRemoteMachineHandlerList(t *testing.T) {
	app := fiber.New()
	repo := &mockRemoteMachineUseCase{machines: make(map[string]*entity.RemoteMachine)}
	handler := &RemoteMachineHandler{uc: repo, l: &mockLogger{}, v: validator.New(validator.WithRequiredStructEnabled())}

	app.Get("/machines", handler.List)

	repo.machines["test-id"] = &entity.RemoteMachine{
		ID:       "test-id",
		Name:     "test-server",
		Host:     "192.168.1.100",
		Port:     22,
		Username: "root",
	}

	req := httptest.NewRequest("GET", "/machines", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRemoteMachineHandlerGet(t *testing.T) {
	app := fiber.New()
	repo := &mockRemoteMachineUseCase{machines: make(map[string]*entity.RemoteMachine)}
	handler := &RemoteMachineHandler{uc: repo, l: &mockLogger{}, v: validator.New(validator.WithRequiredStructEnabled())}

	app.Get("/machines/:id", handler.Get)

	repo.machines["test-id"] = &entity.RemoteMachine{
		ID:   "test-id",
		Name: "test-server",
		Host: "192.168.1.100",
	}

	req := httptest.NewRequest("GET", "/machines/test-id", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRemoteMachineHandlerDelete(t *testing.T) {
	app := fiber.New()
	repo := &mockRemoteMachineUseCase{machines: make(map[string]*entity.RemoteMachine)}
	handler := &RemoteMachineHandler{uc: repo, l: &mockLogger{}, v: validator.New(validator.WithRequiredStructEnabled())}

	app.Delete("/machines/:id", handler.Delete)

	repo.machines["test-id"] = &entity.RemoteMachine{ID: "test-id"}

	req := httptest.NewRequest("DELETE", "/machines/test-id", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRemoteMachineHandlerTestConnection(t *testing.T) {
	app := fiber.New()
	repo := &mockRemoteMachineUseCase{machines: make(map[string]*entity.RemoteMachine)}
	handler := &RemoteMachineHandler{uc: repo, l: &mockLogger{}, v: validator.New(validator.WithRequiredStructEnabled())}

	app.Post("/machines/:id/test", handler.TestConnection)

	repo.machines["test-id"] = &entity.RemoteMachine{ID: "test-id"}

	req := httptest.NewRequest("POST", "/machines/test-id/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRemoteMachineHandlerListContainers(t *testing.T) {
	app := fiber.New()
	repo := &mockRemoteMachineUseCase{machines: make(map[string]*entity.RemoteMachine)}
	handler := &RemoteMachineHandler{uc: repo, l: &mockLogger{}, v: validator.New(validator.WithRequiredStructEnabled())}

	app.Get("/machines/:id/containers", handler.ListContainers)

	repo.machines["test-id"] = &entity.RemoteMachine{ID: "test-id"}

	req := httptest.NewRequest("GET", "/machines/test-id/containers", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRemoteMachineHandlerGetContainerLogs(t *testing.T) {
	app := fiber.New()
	repo := &mockRemoteMachineUseCase{machines: make(map[string]*entity.RemoteMachine)}
	handler := &RemoteMachineHandler{uc: repo, l: &mockLogger{}, v: validator.New(validator.WithRequiredStructEnabled())}

	app.Get("/machines/:id/containers/:containerId/logs", handler.GetContainerLogs)

	repo.machines["test-id"] = &entity.RemoteMachine{ID: "test-id"}

	req := httptest.NewRequest("GET", "/machines/test-id/containers/container-1/logs?tail=50", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRemoteMachineHandlerExecContainer(t *testing.T) {
	app := fiber.New()
	repo := &mockRemoteMachineUseCase{machines: make(map[string]*entity.RemoteMachine)}
	handler := &RemoteMachineHandler{uc: repo, l: &mockLogger{}, v: validator.New(validator.WithRequiredStructEnabled())}

	app.Post("/machines/:id/containers/:containerId/exec", handler.ExecContainer)

	repo.machines["test-id"] = &entity.RemoteMachine{ID: "test-id"}

	body := `{"cmd":["ls","-la"]}`
	req := httptest.NewRequest("POST", "/machines/test-id/containers/container-1/exec", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRemoteMachineHandlerCreateValidationError(t *testing.T) {
	app := fiber.New()
	repo := &mockRemoteMachineUseCase{machines: make(map[string]*entity.RemoteMachine)}
	handler := &RemoteMachineHandler{uc: repo, l: &mockLogger{}, v: validator.New(validator.WithRequiredStructEnabled())}

	app.Post("/machines", handler.Create)

	body := `{"name":"","host":"","username":""}`
	req := httptest.NewRequest("POST", "/machines", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestRemoteMachineHandlerContainerLifecycle(t *testing.T) {
	app := fiber.New()
	repo := &mockRemoteMachineUseCase{machines: make(map[string]*entity.RemoteMachine)}
	handler := &RemoteMachineHandler{uc: repo, l: &mockLogger{}, v: validator.New(validator.WithRequiredStructEnabled())}

	app.Post("/machines/:id/containers/:containerId/start", handler.StartContainer)
	app.Post("/machines/:id/containers/:containerId/stop", handler.StopContainer)
	app.Post("/machines/:id/containers/:containerId/restart", handler.RestartContainer)
	app.Delete("/machines/:id/containers/:containerId", handler.RemoveContainer)

	repo.machines["test-id"] = &entity.RemoteMachine{ID: "test-id"}

	tests := []struct {
		method string
		path   string
	}{
		{"POST", "/machines/test-id/containers/container-1/start"},
		{"POST", "/machines/test-id/containers/container-1/stop"},
		{"POST", "/machines/test-id/containers/container-1/restart"},
		{"DELETE", "/machines/test-id/containers/container-1?force=true"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("%s %s failed: %v", tt.method, tt.path, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("%s %s expected 200, got %d", tt.method, tt.path, resp.StatusCode)
		}
	}
}
