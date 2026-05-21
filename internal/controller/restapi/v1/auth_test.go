package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/usecase"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/stretchr/testify/require"
)

type authHandlerMock struct {
	getCurrentUserFn func(ctx context.Context, token string) (*entity.User, error)
}

func (m *authHandlerMock) Login(context.Context, string, string) (string, *entity.User, error) {
	return "", nil, nil
}

func (m *authHandlerMock) Register(context.Context, string, string, string, string) (*entity.User, error) {
	return nil, nil
}

func (m *authHandlerMock) GetCurrentUser(ctx context.Context, token string) (*entity.User, error) {
	if m.getCurrentUserFn != nil {
		return m.getCurrentUserFn(ctx, token)
	}
	return nil, nil
}

func (m *authHandlerMock) RefreshToken(context.Context, string) (string, error) {
	return "", nil
}

func (m *authHandlerMock) IsSetupComplete(context.Context) (bool, error) {
	return false, nil
}

var _ usecase.Auth = (*authHandlerMock)(nil)

type recordingLogger struct {
	errorCalls int
	warnCalls  int
}

func (l *recordingLogger) Debug(interface{}, ...interface{}) {}
func (l *recordingLogger) Info(string, ...interface{})       {}
func (l *recordingLogger) Warn(string, ...interface{})       { l.warnCalls++ }
func (l *recordingLogger) Error(interface{}, ...interface{}) { l.errorCalls++ }
func (l *recordingLogger) Fatal(interface{}, ...interface{}) {}

var _ logger.Interface = (*recordingLogger)(nil)

func TestAuthHandler_GetMe_SuppressesExpectedAuthFailures(t *testing.T) {
	h := &AuthHandler{
		auth: &authHandlerMock{
			getCurrentUserFn: func(_ context.Context, _ string) (*entity.User, error) {
				return nil, errors.Wrap(errors.ErrUserNotFound, "Auth.GetCurrentUser.GetUserByID")
			},
		},
		l: &recordingLogger{},
	}

	app := fiber.New()
	app.Get("/auth/me", h.GetMe)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer stale-token")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Equal(t, 0, h.l.(*recordingLogger).errorCalls)
	require.Equal(t, 0, h.l.(*recordingLogger).warnCalls)
}
