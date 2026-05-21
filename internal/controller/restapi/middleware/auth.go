package middleware

import (
	"fmt"
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
		// Skip WebSocket upgrade requests — validate the JWT token only.
		if strings.ToLower(c.Get("Upgrade")) == "websocket" {
			tokenString := c.Query("token")
			if tokenString == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"code": 401,
					"msg":  "missing websocket token",
				})
			}

			token, err := parseJWT(tokenString, cfg.Auth.JWTSecret)
			if err != nil || !token.Valid {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"code": 401,
					"msg":  "invalid or expired token",
				})
			}
			if !storeJWTClaims(c, token) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"code": 401,
					"msg":  "invalid token claims",
				})
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

		token, err := parseJWT(tokenString, cfg.Auth.JWTSecret)

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": 401,
				"msg":  "invalid or expired token",
			})
		}

		if !storeJWTClaims(c, token) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": 401,
				"msg":  "invalid token claims",
			})
		}

		return c.Next()
	}
}

func parseJWT(tokenString, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

func storeJWTClaims(c *fiber.Ctx, token *jwt.Token) bool {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return false
	}

	username, _ := claims["username"].(string)
	c.Locals("userID", userID)
	c.Locals("username", username)
	return true
}
