package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lminimum/LiteDock/config"
	"github.com/stretchr/testify/require"
)

func signedToken(t *testing.T, secret string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  "user-1",
		"username": "alice",
	})
	encoded, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return encoded
}

func TestAuthRequiredWebSocketRequiresToken(t *testing.T) {
	app := fiber.New()
	app.Use(AuthRequired(&config.Config{Auth: config.Auth{JWTSecret: "secret"}}))
	app.Get("/ws", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Host", "localhost:8080")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthRequiredWebSocketAllowsValidToken(t *testing.T) {
	app := fiber.New()
	app.Use(AuthRequired(&config.Config{Auth: config.Auth{JWTSecret: "secret"}}))
	app.Get("/ws", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest("GET", "/ws?token="+signedToken(t, "secret"), nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthRequiredWebSocketAllowsTokenWithInvalidOrigin(t *testing.T) {
	app := fiber.New()
	app.Use(AuthRequired(&config.Config{Auth: config.Auth{JWTSecret: "secret"}}))
	app.Get("/ws", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest("GET", "/ws?token="+signedToken(t, "secret"), nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Origin", "http://evil.example")
	req.Header.Set("Host", "localhost:8080")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthRequiredWebSocketAllowsProxiedOrigin(t *testing.T) {
	app := fiber.New()
	app.Use(AuthRequired(&config.Config{Auth: config.Auth{JWTSecret: "secret"}}))
	app.Get("/ws", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest("GET", "/ws?token="+signedToken(t, "secret"), nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Origin", "http://192.168.7.162:3023")
	req.Header.Set("Host", "localhost:8080")
	req.Header.Set("X-Forwarded-Host", "192.168.7.162:3023")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}
