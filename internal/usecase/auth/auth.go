package auth

import (
	"context"
	"time"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UseCase struct {
	repo   repo.UserRepo
	logger logger.Interface
}

var _ UseCaseInterface = (*UseCase)(nil)

func New(repo repo.UserRepo, l logger.Interface) *UseCase {
	return &UseCase{repo: repo, logger: l}
}

func (uc *UseCase) Login(ctx context.Context, username, password string) (string, *entity.User, error) {
	verified, err := uc.repo.VerifyPassword(ctx, username, password)
	if err != nil {
		return "", nil, errors.Wrap(err, "Auth.Login.VerifyPassword")
	}

	if !verified {
		return "", nil, errors.Wrap(errors.ErrInvalidCredentials, "Auth.Login.InvalidCredentials")
	}

	user, err := uc.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", nil, errors.Wrap(err, "Auth.Login.GetUserByUsername")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte("secret-key"))
	if err != nil {
		return "", nil, errors.Wrap(err, "Auth.Login.SignedString")
	}

	return tokenString, user, nil
}

func (uc *UseCase) IsSetupComplete(ctx context.Context) (bool, error) {
	count, err := uc.repo.CountUsers(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (uc *UseCase) Register(ctx context.Context, username, email, password, role string) (*entity.User, error) {
	_, err := uc.repo.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, errors.Wrap(errors.ErrUsernameExists, "Auth.Register.UsernameExists")
	}

	if err != nil && !errors.Is(err, errors.ErrUserNotFound) {
		return nil, errors.Wrap(err, "Auth.Register.GetUserByUsername")
	}

	newUser := entity.User{
		ID:        uuid.New().String(),
		Username:  username,
		Email:     email,
		Password:  password,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = uc.repo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, errors.Wrap(err, "Auth.Register.CreateUser")
	}

	return &newUser, nil
}

func (uc *UseCase) GetCurrentUser(ctx context.Context, tokenString string) (*entity.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Wrap(errors.ErrUnexpectedSignMethod, "Auth.GetCurrentUser.UnexpectedSigningMethod")
		}
		return []byte("secret-key"), nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "Auth.GetCurrentUser.Parse")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return nil, errors.Wrap(errors.ErrInvalidTokenClaims, "Auth.GetCurrentUser.InvalidClaims")
		}

		user, err := uc.repo.GetUserByID(ctx, userID)
		if err != nil {
			return nil, errors.Wrap(err, "Auth.GetCurrentUser.GetUserByID")
		}

		return user, nil
	}

	return nil, errors.Wrap(errors.ErrInvalidToken, "Auth.GetCurrentUser.InvalidToken")
}

func (uc *UseCase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Wrap(errors.ErrUnexpectedSignMethod, "Auth.RefreshToken.UnexpectedSigningMethod")
		}
		return []byte("secret-key"), nil
	})
	if err != nil {
		return "", errors.Wrap(err, "Auth.RefreshToken.Parse")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["username"].(string)
		if !ok {
			return "", errors.Wrap(errors.ErrInvalidTokenClaims, "Auth.RefreshToken.InvalidClaims")
		}

		user, err := uc.repo.GetUserByUsername(ctx, username)
		if err != nil {
			return "", errors.Wrap(err, "Auth.RefreshToken.GetUserByUsername")
		}

		newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  user.ID,
			"username": user.Username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
			"iat":      time.Now().Unix(),
		})

		newTokenString, err := newToken.SignedString([]byte("secret-key"))
		if err != nil {
			return "", errors.Wrap(err, "Auth.RefreshToken.SignedString")
		}

		return newTokenString, nil
	}

	return "", errors.Wrap(errors.ErrInvalidToken, "Auth.RefreshToken.InvalidToken")
}
