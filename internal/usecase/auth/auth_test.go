package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/stretchr/testify/require"
)

// --- mockLogger ---

type mockLogger struct{}

func (m *mockLogger) Debug(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Info(_ string, _ ...interface{})       {}
func (m *mockLogger) Warn(_ string, _ ...interface{})       {}
func (m *mockLogger) Error(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Fatal(_ interface{}, _ ...interface{}) {}

var _ logger.Interface = (*mockLogger)(nil)

// --- mockUserRepo ---

type mockUserRepo struct {
	createUserFn        func(ctx context.Context, user *entity.User) error
	getUserByIDFn       func(ctx context.Context, id string) (*entity.User, error)
	getUserByUsernameFn func(ctx context.Context, username string) (*entity.User, error)
	getUserByEmailFn    func(ctx context.Context, email string) (*entity.User, error)
	updateUserFn        func(ctx context.Context, user *entity.User) error
	deleteUserFn        func(ctx context.Context, id string) error
	verifyPasswordFn    func(ctx context.Context, username, password string) (bool, error)
	updatePasswordFn    func(ctx context.Context, userID, password string) error
	countUsersFn        func(ctx context.Context) (int64, error)
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *entity.User) error {
	if m.createUserFn != nil {
		return m.createUserFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	if m.getUserByUsernameFn != nil {
		return m.getUserByUsernameFn(ctx, username)
	}
	return nil, nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	if m.getUserByEmailFn != nil {
		return m.getUserByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, user *entity.User) error {
	if m.updateUserFn != nil {
		return m.updateUserFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) DeleteUser(ctx context.Context, id string) error {
	if m.deleteUserFn != nil {
		return m.deleteUserFn(ctx, id)
	}
	return nil
}

func (m *mockUserRepo) VerifyPassword(ctx context.Context, username, password string) (bool, error) {
	if m.verifyPasswordFn != nil {
		return m.verifyPasswordFn(ctx, username, password)
	}
	return false, nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, userID, password string) error {
	if m.updatePasswordFn != nil {
		return m.updatePasswordFn(ctx, userID, password)
	}
	return nil
}

func (m *mockUserRepo) CountUsers(ctx context.Context) (int64, error) {
	if m.countUsersFn != nil {
		return m.countUsersFn(ctx)
	}
	return 0, nil
}

var _ repo.UserRepo = (*mockUserRepo)(nil)

// --- helpers ---

const testJWTSecret = "test-secret-key-for-unit-tests"

func newUseCaseForTest(userRepo repo.UserRepo, l logger.Interface, jwtSecret string) *UseCase {
	return &UseCase{
		repo:      userRepo,
		logger:    l,
		jwtSecret: jwtSecret,
	}
}

// generateTestToken creates a valid JWT for testing using HS256.
func generateTestToken(userID, username string, exp time.Time, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      exp.Unix(),
		"iat":      time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic("generateTestToken: " + err.Error())
	}
	return tokenString
}

// --- Login tests ---

func TestLogin_Success(t *testing.T) {
	mockRepo := &mockUserRepo{
		verifyPasswordFn: func(_ context.Context, _, _ string) (bool, error) {
			return true, nil
		},
		getUserByUsernameFn: func(_ context.Context, _ string) (*entity.User, error) {
			return &entity.User{
				ID:       "user-1",
				Username: "testuser",
				Email:    "test@example.com",
				Role:     "user",
			}, nil
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	token, user, err := uc.Login(context.Background(), "testuser", "password123")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotNil(t, user)
	require.Equal(t, "testuser", user.Username)
	require.Equal(t, "user-1", user.ID)
}

func TestLogin_InvalidPassword(t *testing.T) {
	mockRepo := &mockUserRepo{
		verifyPasswordFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	token, user, err := uc.Login(context.Background(), "testuser", "wrongpassword")
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrInvalidCredentials)
	require.Empty(t, token)
	require.Nil(t, user)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := &mockUserRepo{
		verifyPasswordFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, errors.ErrUserNotFound
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	token, user, err := uc.Login(context.Background(), "nonexistent", "password")
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrUserNotFound)
	require.Empty(t, token)
	require.Nil(t, user)
}

func TestLogin_VerifyPasswordError(t *testing.T) {
	mockRepo := &mockUserRepo{
		verifyPasswordFn: func(_ context.Context, _, _ string) (bool, error) {
			return false, errors.ErrTestFail
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	token, user, err := uc.Login(context.Background(), "testuser", "password")
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrTestFail)
	require.Empty(t, token)
	require.Nil(t, user)
}

// --- IsSetupComplete tests ---

func TestIsSetupComplete_True(t *testing.T) {
	mockRepo := &mockUserRepo{
		countUsersFn: func(_ context.Context) (int64, error) {
			return 5, nil
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	complete, err := uc.IsSetupComplete(context.Background())
	require.NoError(t, err)
	require.True(t, complete)
}

func TestIsSetupComplete_False(t *testing.T) {
	mockRepo := &mockUserRepo{
		countUsersFn: func(_ context.Context) (int64, error) {
			return 0, nil
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	complete, err := uc.IsSetupComplete(context.Background())
	require.NoError(t, err)
	require.False(t, complete)
}

func TestIsSetupComplete_Error(t *testing.T) {
	mockRepo := &mockUserRepo{
		countUsersFn: func(_ context.Context) (int64, error) {
			return 0, errors.ErrTestFail
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	complete, err := uc.IsSetupComplete(context.Background())
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrTestFail)
	require.False(t, complete)
}

// --- Register tests ---

func TestRegister_Success(t *testing.T) {
	var createdUser *entity.User

	mockRepo := &mockUserRepo{
		getUserByUsernameFn: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, errors.ErrUserNotFound
		},
		createUserFn: func(_ context.Context, user *entity.User) error {
			createdUser = user
			return nil
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	user, err := uc.Register(context.Background(), "newuser", "new@example.com", "securepass", "user")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "newuser", user.Username)
	require.Equal(t, "new@example.com", user.Email)
	require.Equal(t, "user", user.Role)
	require.NotEmpty(t, user.ID)
	require.NotNil(t, createdUser)

	// Password should be passed through (hashing is repo responsibility)
	require.Equal(t, "securepass", createdUser.Password)
}

func TestRegister_UsernameExists(t *testing.T) {
	mockRepo := &mockUserRepo{
		getUserByUsernameFn: func(_ context.Context, _ string) (*entity.User, error) {
			return &entity.User{ID: "existing-id", Username: "existinguser"}, nil
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	user, err := uc.Register(context.Background(), "existinguser", "email@example.com", "password", "user")
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrUsernameExists)
	require.Nil(t, user)
}

func TestRegister_GetUserByUsernameError(t *testing.T) {
	mockRepo := &mockUserRepo{
		getUserByUsernameFn: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, errors.ErrNotFound
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	user, err := uc.Register(context.Background(), "newuser", "email@example.com", "password", "user")
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrNotFound)
	require.Nil(t, user)
}

// --- GetCurrentUser tests ---

func TestGetCurrentUser_ValidToken(t *testing.T) {
	const userID = "user-42"
	const username = "activeuser"

	tokenString := generateTestToken(userID, username, time.Now().Add(time.Hour*24), testJWTSecret)

	mockRepo := &mockUserRepo{
		getUserByIDFn: func(_ context.Context, id string) (*entity.User, error) {
			require.Equal(t, userID, id)
			return &entity.User{
				ID:       userID,
				Username: username,
				Email:    "active@example.com",
				Role:     "admin",
			}, nil
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	user, err := uc.GetCurrentUser(context.Background(), tokenString)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, userID, user.ID)
	require.Equal(t, username, user.Username)
	require.Equal(t, "admin", user.Role)
}

func TestGetCurrentUser_InvalidToken(t *testing.T) {
	uc := newUseCaseForTest(&mockUserRepo{}, &mockLogger{}, testJWTSecret)

	user, err := uc.GetCurrentUser(context.Background(), "not-a-valid-token-at-all")
	require.Error(t, err)
	require.Nil(t, user)
}

func TestGetCurrentUser_UserNotFound(t *testing.T) {
	const userID = "ghost-user"
	const username = "deleteduser"

	tokenString := generateTestToken(userID, username, time.Now().Add(time.Hour*24), testJWTSecret)

	mockRepo := &mockUserRepo{
		getUserByIDFn: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, errors.ErrUserNotFound
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	user, err := uc.GetCurrentUser(context.Background(), tokenString)
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrUserNotFound)
	require.Nil(t, user)
}

func TestGetCurrentUser_ExpiredToken(t *testing.T) {
	const userID = "user-99"
	const username = "expireduser"

	// Create a token that expired 1 hour ago
	tokenString := generateTestToken(userID, username, time.Now().Add(-time.Hour), testJWTSecret)

	uc := newUseCaseForTest(&mockUserRepo{}, &mockLogger{}, testJWTSecret)

	user, err := uc.GetCurrentUser(context.Background(), tokenString)
	require.Error(t, err)
	require.Nil(t, user)
}

// --- RefreshToken tests ---

func TestRefreshToken_Success(t *testing.T) {
	const userID = "user-55"
	const username = "refresher"

	tokenString := generateTestToken(userID, username, time.Now().Add(time.Hour*24), testJWTSecret)

	mockRepo := &mockUserRepo{
		getUserByUsernameFn: func(_ context.Context, name string) (*entity.User, error) {
			require.Equal(t, username, name)
			return &entity.User{
				ID:       userID,
				Username: username,
				Email:    "refresh@example.com",
				Role:     "user",
			}, nil
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	newToken, err := uc.RefreshToken(context.Background(), tokenString)
	require.NoError(t, err)
	require.NotEmpty(t, newToken)

	// Verify the new token is parseable and contains the same claims
	parsed, err := jwt.Parse(newToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(testJWTSecret), nil
	})
	require.NoError(t, err)
	require.True(t, parsed.Valid)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	require.Equal(t, userID, claims["user_id"])
	require.Equal(t, username, claims["username"])
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	uc := newUseCaseForTest(&mockUserRepo{}, &mockLogger{}, testJWTSecret)

	newToken, err := uc.RefreshToken(context.Background(), "not-a-valid-token")
	require.Error(t, err)
	require.Empty(t, newToken)
}

func TestRefreshToken_ExpiredToken(t *testing.T) {
	const userID = "user-77"
	const username = "staleuser"

	// Create an already-expired token
	tokenString := generateTestToken(userID, username, time.Now().Add(-time.Hour), testJWTSecret)

	uc := newUseCaseForTest(&mockUserRepo{}, &mockLogger{}, testJWTSecret)

	newToken, err := uc.RefreshToken(context.Background(), tokenString)
	require.Error(t, err)
	require.Empty(t, newToken)
}

func TestRefreshToken_UserNotFound(t *testing.T) {
	const userID = "user-88"
	const username = "goneuser"

	tokenString := generateTestToken(userID, username, time.Now().Add(time.Hour*24), testJWTSecret)

	mockRepo := &mockUserRepo{
		getUserByUsernameFn: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, errors.ErrUserNotFound
		},
	}

	uc := newUseCaseForTest(mockRepo, &mockLogger{}, testJWTSecret)

	newToken, err := uc.RefreshToken(context.Background(), tokenString)
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrUserNotFound)
	require.Empty(t, newToken)
}
