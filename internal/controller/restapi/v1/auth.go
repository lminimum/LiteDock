package v1

import (
	stderrors "errors"

	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/config"
	"github.com/lminimum/LiteDock/internal/usecase"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/logger"
)

// AuthHandler -.
type AuthHandler struct {
	auth usecase.Auth
	l    logger.Interface
	cfg  *config.Config
}

// Login - handles POST /auth/login
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	token, user, err := h.auth.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		h.l.Error(err, "Login failed")
		return errorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	return successResponse(c, fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// Register - handles POST /auth/register
// @Summary User registration
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Registration details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	type RegisterRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	user, err := h.auth.Register(c.Context(), req.Username, req.Email, req.Password, role)
	if err != nil {
		h.l.Error(err, "Registration failed")
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return successResponse(c, fiber.Map{
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
// @Summary Check if initial setup is complete
// @Description Returns whether the system has been initially configured
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/setup-status [get]
func (h *AuthHandler) SetupStatus(c *fiber.Ctx) error {
	complete, err := h.auth.IsSetupComplete(c.Context())
	if err != nil {
		h.l.Error(err, "SetupStatus failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, fiber.Map{
		"setup_complete": complete,
	})
}

// GetMe - handles GET /auth/me
// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entity.User
// @Failure 401 {object} map[string]interface{}
// @Router /auth/me [get]
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	// Get token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return errorResponse(c, fiber.StatusUnauthorized, "Authorization header required")
	}

	// Extract token from "Bearer <token>"
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	user, err := h.auth.GetCurrentUser(c.Context(), token)
	if err != nil {
		if !stderrors.Is(err, errors.ErrUserNotFound) && !stderrors.Is(err, errors.ErrInvalidToken) && !stderrors.Is(err, errors.ErrInvalidTokenClaims) && !stderrors.Is(err, errors.ErrUnexpectedSignMethod) {
			h.l.Error(err, "GetCurrentUser failed")
		}
		return errorResponse(c, fiber.StatusUnauthorized, "Invalid or expired token")
	}

	return successResponse(c, fiber.Map{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}
