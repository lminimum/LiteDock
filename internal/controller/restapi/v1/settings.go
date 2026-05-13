package v1

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/pkg/logger"
)

// AISettings holds configurable AI settings.
type AISettings struct {
	APIEndpoint string `json:"apiEndpoint"`
	APIKey      string `json:"apiKey"`
	ModelName   string `json:"modelName"`
}

// AISettingsStore provides thread-safe storage for AI settings.
type AISettingsStore struct {
	mu       sync.RWMutex
	settings AISettings
}

// NewAISettingsStore creates a new AISettingsStore with the given config values.
// If parameters are empty, defaults are applied: ModelName defaults to "gpt-4o".
func NewAISettingsStore(apiEndpoint, apiKey, modelName string) *AISettingsStore {
	if modelName == "" {
		modelName = "gpt-4o"
	}
	return &AISettingsStore{
		settings: AISettings{
			APIEndpoint: apiEndpoint,
			APIKey:      apiKey,
			ModelName:   modelName,
		},
	}
}

// Get returns a copy of the current AI settings.
func (s *AISettingsStore) Get() AISettings {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.settings
}

// Set updates the AI settings with the given values.
func (s *AISettingsStore) Set(settings AISettings) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings = settings
}

// SettingsHandler handles AI settings endpoints.
type SettingsHandler struct {
	store *AISettingsStore
	l     logger.Interface
}

// NewSettingsRoutes registers AI settings routes under /v1/assistant/settings.
func NewSettingsRoutes(protected fiber.Router, store *AISettingsStore, l logger.Interface) {
	h := &SettingsHandler{store: store, l: l}

	protected.Get("/assistant/settings", h.Get)
	protected.Put("/assistant/settings", h.Set)
}

// Get handles GET /v1/assistant/settings — returns current AI settings.
func (h *SettingsHandler) Get(c *fiber.Ctx) error {
	return successResponse(c, h.store.Get())
}

// Set handles PUT /v1/assistant/settings — updates AI settings.
func (h *SettingsHandler) Set(c *fiber.Ctx) error {
	var settings AISettings
	if err := c.BodyParser(&settings); err != nil {
		h.l.Error(err, "SettingsHandler - Set - BodyParser")

		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	h.store.Set(settings)

	return successResponse(c, h.store.Get())
}
