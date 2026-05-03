package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	v1 "github.com/lminimum/LiteDock/internal/controller/restapi/v1"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/usecase"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/stretchr/testify/require"
)

// --- mockNetworkUseCase ---
// In-memory implementation of usecase.Network for integration testing.
// Simulates Docker network CRUD operations without a real Docker daemon.

type mockNetworkUseCase struct {
	mu       sync.RWMutex
	networks map[string]*entity.Network // key: "machineID:name"
}

func newMockNetworkUseCase() *mockNetworkUseCase {
	return &mockNetworkUseCase{
		networks: make(map[string]*entity.Network),
	}
}

func (m *mockNetworkUseCase) ListNetworks(_ context.Context, machineID string) ([]entity.Network, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]entity.Network, 0)
	prefix := machineID + ":"

	for key, net := range m.networks {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			result = append(result, *net)
		}
	}

	return result, nil
}

func (m *mockNetworkUseCase) CreateNetwork(_ context.Context, machineID, name, driver string) (*entity.Network, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if driver == "" {
		driver = "bridge"
	}

	key := machineID + ":" + name
	if _, exists := m.networks[key]; exists {
		return nil, fmt.Errorf("network already exists: %s", name)
	}

	n := &entity.Network{
		ID:     name,
		Name:   name,
		Driver: driver,
		Scope:  "local",
	}

	m.networks[key] = n

	return n, nil
}

var builtInNetworks = map[string]bool{"bridge": true, "host": true, "none": true}

func (m *mockNetworkUseCase) DeleteNetwork(_ context.Context, machineID, networkName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if builtInNetworks[networkName] {
		return fmt.Errorf("cannot delete built-in network: %s", networkName)
	}

	key := machineID + ":" + networkName
	if _, exists := m.networks[key]; !exists {
		return fmt.Errorf("network not found: %s", networkName)
	}

	delete(m.networks, key)

	return nil
}

func (m *mockNetworkUseCase) InspectNetwork(_ context.Context, machineID, networkName string) (*entity.Network, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := machineID + ":" + networkName
	n, exists := m.networks[key]
	if !exists {
		return nil, fmt.Errorf("network not found: %s", networkName)
	}

	return n, nil
}

var _ usecase.Network = (*mockNetworkUseCase)(nil)

// --- test setup ---

func setupTestApp(uc usecase.Network) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	l := logger.New("error")
	v1.NewNetworkRoutes(app.Group("/v1"), uc, l)

	return app
}

// testRequest is a helper for making HTTP test requests against the app.
func testRequest(t *testing.T, app *fiber.App, method, path string, body io.Reader) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), method, path, body)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp
}

// decodeJSON decodes the response body into v.
func decodeJSON(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()

	defer resp.Body.Close()

	err := json.NewDecoder(resp.Body).Decode(v)
	require.NoError(t, err)
}

// --- tests ---

type apiResponse struct {
	Code float64     `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// TestNetworkCreateAndList verifies that creating a network via the API
// causes the network to appear in the list endpoint.
func TestNetworkCreateAndList(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	// 1. List empty
	resp := testRequest(t, app, http.MethodGet, "/v1/machines/local/networks", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body apiResponse
	decodeJSON(t, resp, &body)
	require.Equal(t, float64(200), body.Code)

	netsData, ok := body.Data.(map[string]interface{})
	require.True(t, ok)
	networks, ok := netsData["networks"].([]interface{})
	require.True(t, ok)
	require.Len(t, networks, 0)

	// 2. Create a network
	createBody := `{"name":"my-test-net","driver":"bridge"}`
	resp = testRequest(t, app, http.MethodPost, "/v1/machines/local/networks", strings.NewReader(createBody))
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResponse apiResponse
	decodeJSON(t, resp, &createResponse)
	require.Equal(t, float64(201), createResponse.Code)

	data, ok := createResponse.Data.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "my-test-net", data["name"])
	require.Equal(t, "bridge", data["driver"])

	// 3. List again - verify network appears
	resp = testRequest(t, app, http.MethodGet, "/v1/machines/local/networks", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	decodeJSON(t, resp, &body)
	require.Equal(t, float64(200), body.Code)

	netsData, ok = body.Data.(map[string]interface{})
	require.True(t, ok)
	networks, ok = netsData["networks"].([]interface{})
	require.True(t, ok)
	require.Len(t, networks, 1)
	require.Equal(t, "my-test-net", networks[0].(map[string]interface{})["name"])
}

// TestNetworkDeleteAndVerifyRemoval verifies that deleting a network
// removes it from the list.
func TestNetworkDeleteAndVerifyRemoval(t *testing.T) {
	uc := newMockNetworkUseCase()
	app := setupTestApp(uc)

	_, err := uc.CreateNetwork(context.Background(), "local", "to-delete", "bridge")
	require.NoError(t, err)

	resp := testRequest(t, app, http.MethodGet, "/v1/machines/local/networks", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body apiResponse
	decodeJSON(t, resp, &body)

	netsData, _ := body.Data.(map[string]interface{})
	networks := netsData["networks"].([]interface{})
	require.Len(t, networks, 1)
	require.Equal(t, "to-delete", networks[0].(map[string]interface{})["name"])

	resp = testRequest(t, app, http.MethodDelete, "/v1/machines/local/networks/to-delete", nil)
	var deleteBody apiResponse
	decodeJSON(t, resp, &deleteBody)
	require.Equal(t, float64(200), deleteBody.Code)
	require.Equal(t, "Network deleted successfully", deleteBody.Msg)

	resp = testRequest(t, app, http.MethodGet, "/v1/machines/local/networks", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	decodeJSON(t, resp, &body)
	netsData, _ = body.Data.(map[string]interface{})
	networks = netsData["networks"].([]interface{})
	require.Len(t, networks, 0)
}

func TestDeleteBuiltInNetworkRejected(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	for _, name := range []string{"bridge", "host", "none"} {
		t.Run("reject_"+name, func(t *testing.T) {
			resp := testRequest(t, app, http.MethodDelete, "/v1/machines/local/networks/"+name, nil)
			var body apiResponse
			decodeJSON(t, resp, &body)
			require.Equal(t, float64(500), body.Code)
			require.Contains(t, body.Msg, "cannot delete built-in network")
			require.Contains(t, body.Msg, name)
		})
	}
}

func TestCreateNetworkInvalidBody(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	resp := testRequest(t, app, http.MethodPost, "/v1/machines/local/networks", strings.NewReader(`{"driver":"bridge"}`))
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body apiResponse
	decodeJSON(t, resp, &body)
	require.Equal(t, float64(400), body.Code)
	require.NotEmpty(t, body.Msg)
}

func TestCreateNetworkInvalidJSON(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	resp := testRequest(t, app, http.MethodPost, "/v1/machines/local/networks", strings.NewReader(`{invalid}`))
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body apiResponse
	decodeJSON(t, resp, &body)
	require.Equal(t, float64(400), body.Code)
}

func TestInspectNetwork(t *testing.T) {
	uc := newMockNetworkUseCase()
	app := setupTestApp(uc)

	_, err := uc.CreateNetwork(context.Background(), "local", "inspect-me", "overlay")
	require.NoError(t, err)

	resp := testRequest(t, app, http.MethodGet, "/v1/machines/local/networks/inspect-me", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body apiResponse
	decodeJSON(t, resp, &body)
	require.Equal(t, float64(200), body.Code)

	data, ok := body.Data.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "inspect-me", data["name"])
	require.Equal(t, "overlay", data["driver"])
}

func TestInspectNetworkNotFound(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	resp := testRequest(t, app, http.MethodGet, "/v1/machines/local/networks/nonexistent", nil)
	var body apiResponse
	decodeJSON(t, resp, &body)
	require.Equal(t, float64(500), body.Code)
	require.Contains(t, body.Msg, "not found")
}

func TestDeleteNetworkNotFound(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	resp := testRequest(t, app, http.MethodDelete, "/v1/machines/local/networks/nonexistent", nil)
	var body apiResponse
	decodeJSON(t, resp, &body)
	require.Equal(t, float64(500), body.Code)
	require.Contains(t, body.Msg, "not found")
}

func TestCreateNetworkWithDefaultDriver(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	resp := testRequest(t, app, http.MethodPost, "/v1/machines/local/networks", strings.NewReader(`{"name":"default-driver-net"}`))
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var body apiResponse
	decodeJSON(t, resp, &body)
	require.Equal(t, float64(201), body.Code)

	data, ok := body.Data.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "default-driver-net", data["name"])
	require.Equal(t, "bridge", data["driver"])
}
