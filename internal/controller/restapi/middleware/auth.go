package middleware

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lminimum/LiteDock/config"
)

// AuthRequired creates a middleware that protects routes using JWT authentication.
// It extracts the Bearer token from the Authorization header, validates it,
// and stores user information in the Fiber context locals.
func AuthRequired(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip WebSocket upgrade requests — validate via Origin + query param token
		if c.Get("Upgrade") == "websocket" {
			origin := c.Get("Origin")
			if origin == "" {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"code": 403,
					"msg":  "origin header required",
				})
			}

			parsedURL, err := url.Parse(origin)
			if err != nil {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"code": 403,
					"msg":  "invalid origin",
				})
			}

			host := parsedURL.Hostname()
			serverHost := c.Hostname()
			if h, _, splitErr := net.SplitHostPort(serverHost); splitErr == nil {
				serverHost = h
			}

			allowed := false
			for _, o := range []string{serverHost, "localhost", "127.0.0.1"} {
				if host == o {
					allowed = true
					break
				}
			}
			if !allowed {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"code": 403,
					"msg":  "origin not allowed",
				})
			}

			// If token is provided, it must be valid (401 if invalid).
			// If no token, allow access (backward compat with Dashboard WS).
			tokenString := c.Query("token")
			if tokenString != "" {
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return []byte(cfg.Auth.JWTSecret), nil
				})
				if err != nil || !token.Valid {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
						"code": 401,
						"msg":  "invalid or expired token",
					})
				}
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					c.Locals("userID", claims["user_id"])
					c.Locals("username", claims["username"])
				}
			}

			return c.Next()
		}

		// Extract token from Authorization: Bearer <token>
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": 401,
				"msg":  "missing or invalid authorization header",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate JWT (same logic as auth.go GetCurrentUser)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.Auth.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": 401,
				"msg":  "invalid or expired token",
			})
		}

		// Store user info in context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Locals("userID", claims["user_id"])
			c.Locals("username", claims["username"])
		}

		return c.Next()
	}
}
