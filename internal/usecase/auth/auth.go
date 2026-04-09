package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// UseCase -.
type UseCase struct {
	repo   repo.UserRepo
	logger logger.Interface
}

var _ UseCaseInterface = (*UseCase)(nil)

// New -.
func New(repo repo.UserRepo, l logger.Interface) *UseCase {
	return &UseCase{repo: repo, logger: l}
}

// Login authenticates user and returns token
func (uc *UseCase) Login(ctx context.Context, username, password string) (string, *entity.User, error) {
	verified, err := uc.repo.VerifyPassword(ctx, username, password)
	if err != nil {
		return "", nil, fmt.Errorf("UseCase - Login - uc.repo.VerifyPassword: %w", err)
	}

	if !verified {
		return "", nil, fmt.Errorf("invalid credentials")
	}

	user, err := uc.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", nil, fmt.Errorf("UseCase - Login - uc.repo.GetUserByUsername: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24 hour expiry
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte("secret-key")) // In production, use environment variable
	if err != nil {
		return "", nil, fmt.Errorf("UseCase - Login - token.SignedString: %w", err)
	}

	return tokenString, user, nil
}

// Register creates a new user
func (uc *UseCase) Register(ctx context.Context, username, password, email, role string) (*entity.User, error) {
	// Check if user already exists
	_, err := uc.repo.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, fmt.Errorf("username already exists")
	}

	// If there's an error but it's not "not found", then there's a real error
	if err != nil && err.Error() != "usecase: user not found" {
		return nil, fmt.Errorf("UseCase - Register - uc.repo.GetUserByUsername: %w", err)
	}

	// Create new user
	newUser := entity.User{
		ID:        uuid.New().String(),
		Username:  username,
		Email:     email,
		Password:  password, // The repo will hash this
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = uc.repo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("UseCase - Register - uc.repo.CreateUser: %w", err)
	}

	return &newUser, nil
}

// GetCurrentUser gets user info from token
func (uc *UseCase) GetCurrentUser(ctx context.Context, tokenString string) (*entity.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret-key"), nil // In production, use environment variable
	})
	if err != nil {
		return nil, fmt.Errorf("UseCase - GetCurrentUser - jwt.Parse: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid token claims")
		}

		user, err := uc.repo.GetUserByID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("UseCase - GetCurrentUser - uc.repo.GetUserByID: %w", err)
		}

		return user, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken refreshes authentication token
func (uc *UseCase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// For simplicity, we'll just create a new token
	// In a real implementation, you'd have a refresh token system

	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret-key"), nil // In production, use environment variable
	})
	if err != nil {
		return "", fmt.Errorf("UseCase - RefreshToken - jwt.Parse: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["username"].(string)
		if !ok {
			return "", fmt.Errorf("invalid token claims")
		}

		// Get user by username
		user, err := uc.repo.GetUserByUsername(ctx, username)
		if err != nil {
			return "", fmt.Errorf("UseCase - RefreshToken - uc.repo.GetUserByUsername: %w", err)
		}

		// Create a new token
		newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  user.ID,
			"username": user.Username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24 hour expiry
			"iat":      time.Now().Unix(),
		})

		newTokenString, err := newToken.SignedString([]byte("secret-key"))
		if err != nil {
			return "", fmt.Errorf("UseCase - RefreshToken - newToken.SignedString: %w", err)
		}

		return newTokenString, nil
	}

	return "", fmt.Errorf("invalid token")
}
