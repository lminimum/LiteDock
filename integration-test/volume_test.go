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

// --- mockVolumeUseCase ---

type mockVolumeUseCase struct {
	mu     sync.RWMutex
	volumes map[string]*entity.Volume // key: "machineID:name"
}

func newMockVolumeUseCase() *mockVolumeUseCase {
	return &mockVolumeUseCase{
		volumes: make(map[string]*entity.Volume),
	}
}

func (m *mockVolumeUseCase) ListVolumes(_ context.Context, machineID string) ([]entity.Volume, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]entity.Volume, 0)
	prefix := machineID + ":"

	for key, vol := range m.volumes {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			result = append(result, *vol)
		}
	}
	return result, nil
}

func (m *mockVolumeUseCase) CreateVolume(_ context.Context, machineID, name, driver string) (*entity.Volume, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := machineID + ":" + name
	if _, exists := m.volumes[key]; exists {
		return nil, fmt.Errorf("volume %q already exists", name)
	}

	if driver == "" {
		driver = "local"
	}

	vol := &entity.Volume{
		Name:      name,
		MachineID: machineID,
		Driver:    driver,
	}
	m.volumes[key] = vol
	return vol, nil
}

func (m *mockVolumeUseCase) DeleteVolume(_ context.Context, machineID, volumeName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := machineID + ":" + volumeName
	if _, exists := m.volumes[key]; !exists {
		return fmt.Errorf("volume %q not found", volumeName)
	}
	delete(m.volumes, key)
	return nil
}

func (m *mockVolumeUseCase) InspectVolume(_ context.Context, machineID, volumeName string) (*entity.Volume, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := machineID + ":" + volumeName
	vol, exists := m.volumes[key]
	if !exists {
		return nil, fmt.Errorf("volume %q not found", volumeName)
	}
	return vol, nil
}

var _ usecase.Volume = (*mockVolumeUseCase)(nil)

// --- helpers ---

type volumeResponse struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data *volumeData      `json:"data"`
}

type volumeData struct {
	Volume  *entity.Volume   `json:"volume"`
	Volumes []entity.Volume  `json:"volumes"`
	Message string           `json:"message"`
}

func volumeListResponse(t *testing.T, body io.Reader) (int, string, []entity.Volume) {
	t.Helper()
	var raw struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Volumes []entity.Volume `json:"volumes"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(body).Decode(&raw))
	return raw.Code, raw.Msg, raw.Data.Volumes
}

func volumeCreateResponse(t *testing.T, body io.Reader) (int, string, *entity.Volume) {
	t.Helper()
	var raw struct {
		Code int            `json:"code"`
		Msg  string         `json:"msg"`
		Data *entity.Volume `json:"data"`
	}
	require.NoError(t, json.NewDecoder(body).Decode(&raw))
	return raw.Code, raw.Msg, raw.Data
}

func volumeGetResponse(t *testing.T, body io.Reader) (int, string, *entity.Volume) {
	t.Helper()
	var raw struct {
		Code int            `json:"code"`
		Msg  string         `json:"msg"`
		Data *entity.Volume `json:"data"`
	}
	require.NoError(t, json.NewDecoder(body).Decode(&raw))
	return raw.Code, raw.Msg, raw.Data
}

func volumeDeleteResponse(t *testing.T, body io.Reader) (int, string) {
	t.Helper()
	var raw struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	require.NoError(t, json.NewDecoder(body).Decode(&raw))
	return raw.Code, raw.Msg
}

func volumeErrorResponse(t *testing.T, body io.Reader) (int, string) {
	t.Helper()
	var raw struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data *struct{} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(body).Decode(&raw))
	return raw.Code, raw.Msg
}

func newVolumeTestApp() *fiber.App {
	app := fiber.New()
	api := app.Group("/v1")
	mockUC := newMockVolumeUseCase()
	log := logger.New("error")

	v1.NewVolumeRoutes(api, mockUC, log)
	return app
}

// --- tests ---

func TestVolumeCreateAndList(t *testing.T) {
	app := newVolumeTestApp()

	body := `{"name":"test-vol","driver":"local"}`
	req, _ := http.NewRequest(http.MethodPost, "/v1/machines/local/volumes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	code, msg, vol := volumeCreateResponse(t, resp.Body)
	require.Equal(t, 201, code)
	require.Equal(t, "created", msg)
	require.NotNil(t, vol)
	require.Equal(t, "test-vol", vol.Name)

	// List volumes
	req2, _ := http.NewRequest(http.MethodGet, "/v1/machines/local/volumes", nil)
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	code2, _, volumes := volumeListResponse(t, resp2.Body)
	require.Equal(t, 200, code2)
	require.Len(t, volumes, 1)
	require.Equal(t, "test-vol", volumes[0].Name)
}

func TestVolumeDeleteAndVerifyRemoval(t *testing.T) {
	app := newVolumeTestApp()

	// Create
	createBody := `{"name":"delete-test-vol","driver":"local"}`
	req, _ := http.NewRequest(http.MethodPost, "/v1/machines/local/volumes", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	// Delete
	req2, _ := http.NewRequest(http.MethodDelete, "/v1/machines/local/volumes/delete-test-vol", nil)
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	code, delMsg := volumeDeleteResponse(t, resp2.Body)
	require.Equal(t, 200, code)
	require.Contains(t, delMsg, "deleted")

	// Verify removal
	req3, _ := http.NewRequest(http.MethodGet, "/v1/machines/local/volumes", nil)
	resp3, err := app.Test(req3)
	require.NoError(t, err)
	defer resp3.Body.Close()

	_, _, volumes := volumeListResponse(t, resp3.Body)
	require.Len(t, volumes, 0)
}

func TestCreateVolumeInvalidBody(t *testing.T) {
	app := newVolumeTestApp()

	body := `{"name":""}`
	req, _ := http.NewRequest(http.MethodPost, "/v1/machines/local/volumes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	code, _ := volumeErrorResponse(t, resp.Body)
	require.Equal(t, 400, code)
}

func TestCreateVolumeInvalidJSON(t *testing.T) {
	app := newVolumeTestApp()

	body := `{invalid json}`
	req, _ := http.NewRequest(http.MethodPost, "/v1/machines/local/volumes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	code, _ := volumeErrorResponse(t, resp.Body)
	require.Equal(t, 400, code)
}

func TestInspectVolume(t *testing.T) {
	app := newVolumeTestApp()

	// Create first
	createBody := `{"name":"inspect-vol","driver":"local"}`
	req, _ := http.NewRequest(http.MethodPost, "/v1/machines/local/volumes", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	// Inspect
	req2, _ := http.NewRequest(http.MethodGet, "/v1/machines/local/volumes/inspect-vol", nil)
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	code, _, vol := volumeGetResponse(t, resp2.Body)
	require.Equal(t, 200, code)
	require.NotNil(t, vol)
	require.Equal(t, "inspect-vol", vol.Name)
	require.Equal(t, "local", vol.Driver)
}

func TestInspectVolumeNotFound(t *testing.T) {
	app := newVolumeTestApp()

	req, _ := http.NewRequest(http.MethodGet, "/v1/machines/local/volumes/nonexistent", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	code, _ := volumeErrorResponse(t, resp.Body)
	require.Equal(t, 500, code)
}

func TestDeleteVolumeNotFound(t *testing.T) {
	app := newVolumeTestApp()

	req, _ := http.NewRequest(http.MethodDelete, "/v1/machines/local/volumes/nonexistent", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	code, _ := volumeErrorResponse(t, resp.Body)
	require.Equal(t, 500, code)
}

func TestCreateVolumeWithDefaultDriver(t *testing.T) {
	app := newVolumeTestApp()

	// Create without driver - should default to "local" in mock
	body := `{"name":"default-driver-vol"}`
	req, _ := http.NewRequest(http.MethodPost, "/v1/machines/local/volumes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	code, msg, vol := volumeCreateResponse(t, resp.Body)
	require.Equal(t, 201, code)
	require.Equal(t, "created", msg)
	require.NotNil(t, vol)
	require.Equal(t, "default-driver-vol", vol.Name)
}
