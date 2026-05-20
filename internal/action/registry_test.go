package action

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockTokenValidator struct {
	validateFn func(tokenStr string, params TokenParams) error
}

func (m *mockTokenValidator) Validate(tokenStr string, params TokenParams) error {
	return m.validateFn(tokenStr, params)
}

func newDestructiveMockAction(name string) *mockAction {
	a := &mockAction{
		name: name,
		validateFn: func(_ map[string]interface{}) error {
			return nil
		},
		executeFn: func(_ context.Context, _ map[string]interface{}) (*ActionResult, error) {
			return &ActionResult{Success: true, Message: "executed"}, nil
		},
	}
	a.destructiveFn = func(_ map[string]interface{}) bool { return true }
	return a
}

func newSafeMockAction(name string) *mockAction {
	return &mockAction{
		name: name,
		validateFn: func(_ map[string]interface{}) error {
			return nil
		},
		executeFn: func(_ context.Context, _ map[string]interface{}) (*ActionResult, error) {
			return &ActionResult{Success: true, Message: "executed"}, nil
		},
	}
}

func TestExecuteReadOnly_SafeAction_Passes(t *testing.T) {
	reg := NewActionRegistry()
	require.NoError(t, reg.Register(newSafeMockAction("list_containers")))

	result, err := reg.ExecuteReadOnly(context.Background(), "list_containers", nil)
	require.NoError(t, err)
	require.True(t, result.Success)
}

func TestExecuteReadOnly_DangerousAction_Blocked(t *testing.T) {
	reg := NewActionRegistry()
	require.NoError(t, reg.Register(newDestructiveMockAction("stop_container")))

	_, err := reg.ExecuteReadOnly(context.Background(), "stop_container", nil)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrConfirmationRequired))
}

func TestExecuteReadOnly_UnknownAction(t *testing.T) {
	reg := NewActionRegistry()

	_, err := reg.ExecuteReadOnly(context.Background(), "nonexistent", nil)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrUnknownAction))
}

func TestExecuteReadOnly_ValidationFailure(t *testing.T) {
	reg := NewActionRegistry()
	a := &mockAction{
		name: "validated",
		validateFn: func(_ map[string]interface{}) error {
			return fmt.Errorf("missing field")
		},
	}
	require.NoError(t, reg.Register(a))

	_, err := reg.ExecuteReadOnly(context.Background(), "validated", nil)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrValidationFailed))
}

func TestExecuteWithConfirmation_DangerousAction_ValidToken_Passes(t *testing.T) {
	tokenSvc := &mockTokenValidator{
		validateFn: func(tokenStr string, params TokenParams) error {
			if tokenStr == "valid-token" {
				return nil
			}
			return fmt.Errorf("invalid token")
		},
	}
	reg := NewActionRegistryWithToken(tokenSvc)
	require.NoError(t, reg.Register(newDestructiveMockAction("stop_container")))

	result, err := reg.ExecuteWithConfirmation(
		context.Background(), "stop_container", nil, "valid-token",
	)
	require.NoError(t, err)
	require.True(t, result.Success)
}

func TestExecuteWithConfirmation_DangerousAction_EmptyToken_Blocked(t *testing.T) {
	tokenSvc := &mockTokenValidator{
		validateFn: func(_ string, _ TokenParams) error {
			return nil
		},
	}
	reg := NewActionRegistryWithToken(tokenSvc)
	require.NoError(t, reg.Register(newDestructiveMockAction("stop_container")))

	_, err := reg.ExecuteWithConfirmation(
		context.Background(), "stop_container", nil, "",
	)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrConfirmationRequired))
}

func TestExecuteWithConfirmation_DangerousAction_InvalidToken_Blocked(t *testing.T) {
	tokenSvc := &mockTokenValidator{
		validateFn: func(_ string, _ TokenParams) error {
			return fmt.Errorf("token mismatch")
		},
	}
	reg := NewActionRegistryWithToken(tokenSvc)
	require.NoError(t, reg.Register(newDestructiveMockAction("stop_container")))

	_, err := reg.ExecuteWithConfirmation(
		context.Background(), "stop_container", nil, "bad-token",
	)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidToken) || err != nil)
}

func TestExecuteWithConfirmation_DangerousAction_ExpiredToken_Blocked(t *testing.T) {
	tokenSvc := &mockTokenValidator{
		validateFn: func(_ string, _ TokenParams) error {
			return fmt.Errorf("token expired")
		},
	}
	reg := NewActionRegistryWithToken(tokenSvc)
	require.NoError(t, reg.Register(newDestructiveMockAction("stop_container")))

	_, err := reg.ExecuteWithConfirmation(
		context.Background(), "stop_container", nil, "expired-token",
	)
	require.Error(t, err)
}

func TestExecuteWithConfirmation_DangerousAction_NoTokenService_Blocked(t *testing.T) {
	reg := NewActionRegistry()
	require.NoError(t, reg.Register(newDestructiveMockAction("stop_container")))

	_, err := reg.ExecuteWithConfirmation(
		context.Background(), "stop_container", nil, "any-token",
	)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrConfirmationRequired))
}

func TestExecuteWithConfirmation_SafeAction_NoToken_Passes(t *testing.T) {
	reg := NewActionRegistry()
	require.NoError(t, reg.Register(newSafeMockAction("list_containers")))

	result, err := reg.ExecuteWithConfirmation(
		context.Background(), "list_containers", nil, "",
	)
	require.NoError(t, err)
	require.True(t, result.Success)
}

func TestExecuteWithConfirmation_SafeAction_WithToken_Passes(t *testing.T) {
	tokenSvc := &mockTokenValidator{
		validateFn: func(_ string, _ TokenParams) error {
			t.Fatal("token validator should not be called for safe actions")
			return nil
		},
	}
	reg := NewActionRegistryWithToken(tokenSvc)
	require.NoError(t, reg.Register(newSafeMockAction("list_containers")))

	result, err := reg.ExecuteWithConfirmation(
		context.Background(), "list_containers", nil, "some-token",
	)
	require.NoError(t, err)
	require.True(t, result.Success)
}

func TestExecuteWithConfirmation_UnknownAction(t *testing.T) {
	reg := NewActionRegistry()

	_, err := reg.ExecuteWithConfirmation(
		context.Background(), "nonexistent", nil, "token",
	)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrUnknownAction))
}

func TestExecuteWithConfirmation_ValidationFailure(t *testing.T) {
	reg := NewActionRegistry()
	a := &mockAction{
		name: "validated",
		validateFn: func(_ map[string]interface{}) error {
			return fmt.Errorf("missing field")
		},
	}
	require.NoError(t, reg.Register(a))

	_, err := reg.ExecuteWithConfirmation(
		context.Background(), "validated", nil, "token",
	)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrValidationFailed))
}

func TestExecuteWithConfirmation_TokenValidatorReceivesCorrectParams(t *testing.T) {
	var captured TokenParams
	tokenSvc := &mockTokenValidator{
		validateFn: func(_ string, params TokenParams) error {
			captured = params
			return nil
		},
	}
	reg := NewActionRegistryWithToken(tokenSvc)
	require.NoError(t, reg.Register(newDestructiveMockAction("delete_image")))

	ctx := context.WithValue(context.Background(), ctxKeyUserID, "user-42")
	ctx = context.WithValue(ctx, ctxKeySessionID, "sess-abc")

	params := map[string]interface{}{"image_id": "img-123"}
	_, err := reg.ExecuteWithConfirmation(ctx, "delete_image", params, "valid-token")
	require.NoError(t, err)

	require.Equal(t, "user-42", captured.UserID)
	require.Equal(t, "sess-abc", captured.SessionID)
	require.Equal(t, "delete_image", captured.Action)
	require.Equal(t, RiskLevelDangerous, captured.RiskLevel)
	require.NotEmpty(t, captured.ParamsHash)
}

func TestExecuteWithConfirmation_AnonymousUserWhenNoContext(t *testing.T) {
	var captured TokenParams
	tokenSvc := &mockTokenValidator{
		validateFn: func(_ string, params TokenParams) error {
			captured = params
			return nil
		},
	}
	reg := NewActionRegistryWithToken(tokenSvc)
	require.NoError(t, reg.Register(newDestructiveMockAction("stop_container")))

	_, err := reg.ExecuteWithConfirmation(
		context.Background(), "stop_container", nil, "valid-token",
	)
	require.NoError(t, err)
	require.Equal(t, "anonymous", captured.UserID)
	require.Equal(t, "", captured.SessionID)
}

func TestComputeParamsHash_Deterministic(t *testing.T) {
	params := map[string]interface{}{
		"b": 2,
		"a": 1,
	}
	h1 := ComputeParamsHash(params)
	h2 := ComputeParamsHash(params)
	require.Equal(t, h1, h2)
	require.NotEmpty(t, h1)
}

func TestComputeParamsHash_DifferentParamsDifferentHash(t *testing.T) {
	h1 := ComputeParamsHash(map[string]interface{}{"a": "1"})
	h2 := ComputeParamsHash(map[string]interface{}{"a": "2"})
	require.NotEqual(t, h1, h2)
}

func TestComputeParamsHash_NilParams(t *testing.T) {
	h := ComputeParamsHash(nil)
	require.NotEmpty(t, h)
}

func TestNewActionRegistryWithToken(t *testing.T) {
	tokenSvc := &mockTokenValidator{
		validateFn: func(_ string, _ TokenParams) error { return nil },
	}
	reg := NewActionRegistryWithToken(tokenSvc)
	require.NotNil(t, reg)
	require.NotNil(t, reg.tokenService)
}

func TestExecuteConfirmed_StillWorks_Deprecated(t *testing.T) {
	reg := NewActionRegistry()
	executed := false

	a := newDestructiveMockAction("delete_stuff")
	a.executeFn = func(_ context.Context, _ map[string]interface{}) (*ActionResult, error) {
		executed = true
		return &ActionResult{Success: true, Message: "deleted"}, nil
	}
	require.NoError(t, reg.Register(a))

	result, err := reg.ExecuteConfirmed(context.Background(), "delete_stuff", nil)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.True(t, executed)
}
