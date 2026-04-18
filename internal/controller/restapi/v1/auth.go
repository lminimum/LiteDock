package v1

import (
	"github.com/lminimum/LiteDock/config"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/usecase"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// AuthHandler -.
type AuthHandler struct {
	auth usecase.Auth
	l    logger.Interface
	cfg  *config.Config
}

// Login - handles POST /auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	token, user, err := h.auth.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		h.l.Error(err, "Login failed")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"token":   token,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// Register - handles POST /auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	type RegisterRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	user, err := h.auth.Register(c.Context(), req.Username, req.Email, req.Password, role)
	if err != nil {
		h.l.Error(err, "Registration failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User registered successfully",
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// SetupStatus - handles GET /auth/setup-status
func (h *AuthHandler) SetupStatus(c *fiber.Ctx) error {
	complete, err := h.auth.IsSetupComplete(c.Context())
	if err != nil {
		h.l.Error(err, "SetupStatus failed")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"setup_complete": complete,
	})
}

// GetMe - handles GET /auth/me
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	// Get token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Authorization header required",
		})
	}

	// Extract token from "Bearer <token>"
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	user, err := h.auth.GetCurrentUser(c.Context(), token)
	if err != nil {
		h.l.Error(err, "GetCurrentUser failed")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid or expired token",
		})
	}

	return c.JSON(entity.User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	})
}
