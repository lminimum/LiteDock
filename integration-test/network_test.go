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

// decodeJSON decodes the response body into a map.
func decodeJSON(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()

	defer resp.Body.Close()

	var result map[string]interface{}

	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	return result
}

// --- tests ---

// TestNetworkCreateAndList verifies that creating a network via the API
// causes the network to appear in the list endpoint.
func TestNetworkCreateAndList(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	// 1. List empty
	resp := testRequest(t, app, http.MethodGet, "/v1/machines/local/networks", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeJSON(t, resp)
	require.True(t, body["success"].(bool))

	networks, ok := body["networks"].([]interface{})
	require.True(t, ok, "expected 'networks' array in response")
	require.Len(t, networks, 0, "expected empty network list initially")

	// 2. Create a network
	createBody := `{"name":"my-test-net","driver":"bridge"}`
	resp = testRequest(t, app, http.MethodPost, "/v1/machines/local/networks", strings.NewReader(createBody))
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	createResponse := decodeJSON(t, resp)
	require.True(t, createResponse["success"].(bool))

	data, ok := createResponse["data"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "my-test-net", data["name"])
	require.Equal(t, "bridge", data["driver"])

	// 3. List again - verify network appears
	resp = testRequest(t, app, http.MethodGet, "/v1/machines/local/networks", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body = decodeJSON(t, resp)
	require.True(t, body["success"].(bool))

	networks, ok = body["networks"].([]interface{})
	require.True(t, ok)
	require.Len(t, networks, 1, "expected 1 network in list after creation")
	require.Equal(t, "my-test-net", networks[0].(map[string]interface{})["name"])
}

// TestNetworkDeleteAndVerifyRemoval verifies that deleting a network
// removes it from the list.
func TestNetworkDeleteAndVerifyRemoval(t *testing.T) {
	uc := newMockNetworkUseCase()
	app := setupTestApp(uc)

	// Pre-create a network via the mock so we have something to delete
	_, err := uc.CreateNetwork(context.Background(), "local", "to-delete", "bridge")
	require.NoError(t, err)

	// 1. Verify it's in the list
	resp := testRequest(t, app, http.MethodGet, "/v1/machines/local/networks", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeJSON(t, resp)
	networks := body["networks"].([]interface{})
	require.Len(t, networks, 1)
	require.Equal(t, "to-delete", networks[0].(map[string]interface{})["name"])

	// 2. Delete the network
	resp = testRequest(t, app, http.MethodDelete, "/v1/machines/local/networks/to-delete", nil)
	deleteBody := decodeJSON(t, resp)
	require.True(t, deleteBody["success"].(bool))

	// 3. Verify it's gone
	resp = testRequest(t, app, http.MethodGet, "/v1/machines/local/networks", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body = decodeJSON(t, resp)
	networks = body["networks"].([]interface{})
	require.Len(t, networks, 0, "expected empty list after deletion")
}

// TestDeleteBuiltInNetworkRejected verifies that attempting to delete
// built-in Docker networks (bridge, host, none) is rejected.
func TestDeleteBuiltInNetworkRejected(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	builtIns := []string{"bridge", "host", "none"}

	for _, name := range builtIns {
		t.Run("reject_"+name, func(t *testing.T) {
			resp := testRequest(t, app, http.MethodDelete, "/v1/machines/local/networks/"+name, nil)

			body := decodeJSON(t, resp)
			require.False(t, body["success"].(bool), "expected failure for built-in network %s", name)

			msg, ok := body["message"].(string)
			require.True(t, ok)
			require.Contains(t, msg, "cannot delete built-in network")
			require.Contains(t, msg, name)
		})
	}
}

// TestCreateNetworkInvalidBody verifies that missing required fields
// in the request body produce a 400 validation error.
func TestCreateNetworkInvalidBody(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	// Missing required 'name' field
	body := `{"driver":"bridge"}`
	resp := testRequest(t, app, http.MethodPost, "/v1/machines/local/networks", strings.NewReader(body))
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	respBody := decodeJSON(t, resp)
	require.False(t, respBody["success"].(bool))
	require.NotEmpty(t, respBody["message"])
}

// TestCreateNetworkInvalidJSON verifies that malformed JSON
// produces a 400 error response.
func TestCreateNetworkInvalidJSON(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	body := `{invalid}`
	resp := testRequest(t, app, http.MethodPost, "/v1/machines/local/networks", strings.NewReader(body))
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	respBody := decodeJSON(t, resp)
	require.False(t, respBody["success"].(bool))
}

// TestInspectNetwork returns network details for an existing network.
func TestInspectNetwork(t *testing.T) {
	uc := newMockNetworkUseCase()
	app := setupTestApp(uc)

	// Pre-create a network
	_, err := uc.CreateNetwork(context.Background(), "local", "inspect-me", "overlay")
	require.NoError(t, err)

	resp := testRequest(t, app, http.MethodGet, "/v1/machines/local/networks/inspect-me", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeJSON(t, resp)
	require.True(t, body["success"].(bool))

	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "inspect-me", data["name"])
	require.Equal(t, "overlay", data["driver"])
}

// TestInspectNetworkNotFound verifies that inspecting a non-existent
// network returns an error response.
func TestInspectNetworkNotFound(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	resp := testRequest(t, app, http.MethodGet, "/v1/machines/local/networks/nonexistent", nil)

	body := decodeJSON(t, resp)
	require.False(t, body["success"].(bool))
	require.Contains(t, body["message"].(string), "not found")
}

// TestDeleteNetworkNotFound verifies error handling when deleting
// a non-existent network.
func TestDeleteNetworkNotFound(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	resp := testRequest(t, app, http.MethodDelete, "/v1/machines/local/networks/nonexistent", nil)

	body := decodeJSON(t, resp)
	require.False(t, body["success"].(bool))
	require.Contains(t, body["message"].(string), "not found")
}

// TestCreateNetworkWithDefaultDriver verifies that when no driver is specified,
// "bridge" is used as the default.
func TestCreateNetworkWithDefaultDriver(t *testing.T) {
	app := setupTestApp(newMockNetworkUseCase())

	createBody := `{"name":"default-driver-net"}`
	resp := testRequest(t, app, http.MethodPost, "/v1/machines/local/networks", strings.NewReader(createBody))
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	createResponse := decodeJSON(t, resp)
	require.True(t, createResponse["success"].(bool))

	data := createResponse["data"].(map[string]interface{})
	require.Equal(t, "default-driver-net", data["name"])
	require.Equal(t, "bridge", data["driver"])
}
