package assistant

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfirmationToken_ValidToken(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)
	params := map[string]string{"container_id": "abc123", "action": "stop"}

	token, err := ts.Generate(ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
		Expiry:     time.Now().Add(2 * time.Minute),
	})
	require.NoError(t, err)
	require.NotEmpty(t, token)

	err = ts.Validate(token, ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
	})
	require.NoError(t, err)
}

func TestConfirmationToken_ExpiredToken(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)
	params := map[string]string{"container_id": "abc123"}

	token, err := ts.Generate(ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
		Expiry:     time.Now().Add(-1 * time.Minute),
	})
	require.NoError(t, err)

	err = ts.Validate(token, ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
	})
	require.ErrorIs(t, err, ErrTokenExpired)
}

func TestConfirmationToken_TamperedToken(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)
	params := map[string]string{"container_id": "abc123"}

	token, err := ts.Generate(ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
		Expiry:     time.Now().Add(2 * time.Minute),
	})
	require.NoError(t, err)

	tamperedToken := token + "tampered"

	err = ts.Validate(tamperedToken, ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
	})
	require.ErrorIs(t, err, ErrTokenInvalid)
}

func TestConfirmationToken_ParamsMismatch(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)
	params := map[string]string{"container_id": "abc123"}

	token, err := ts.Generate(ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
		Expiry:     time.Now().Add(2 * time.Minute),
	})
	require.NoError(t, err)

	differentParams := map[string]string{"container_id": "different_id"}
	err = ts.Validate(token, ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(differentParams),
		RiskLevel:  "high",
	})
	require.ErrorIs(t, err, ErrTokenMismatch)
}

func TestConfirmationToken_ActionMismatch(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)
	params := map[string]string{"container_id": "abc123"}

	token, err := ts.Generate(ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
		Expiry:     time.Now().Add(2 * time.Minute),
	})
	require.NoError(t, err)

	err = ts.Validate(token, ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "start_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
	})
	require.ErrorIs(t, err, ErrTokenMismatch)
}

func TestConfirmationToken_RiskLevelMismatch(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)
	params := map[string]string{"container_id": "abc123"}

	token, err := ts.Generate(ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
		Expiry:     time.Now().Add(2 * time.Minute),
	})
	require.NoError(t, err)

	err = ts.Validate(token, ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "low",
	})
	require.ErrorIs(t, err, ErrTokenMismatch)
}

func TestConfirmationToken_UserIDMismatch(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)
	params := map[string]string{"container_id": "abc123"}

	token, err := ts.Generate(ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
		Expiry:     time.Now().Add(2 * time.Minute),
	})
	require.NoError(t, err)

	err = ts.Validate(token, ActionConfirmationToken{
		UserID:     "user2",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
	})
	require.ErrorIs(t, err, ErrTokenMismatch)
}

func TestConfirmationToken_SessionIDMismatch(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)
	params := map[string]string{"container_id": "abc123"}

	token, err := ts.Generate(ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
		Expiry:     time.Now().Add(2 * time.Minute),
	})
	require.NoError(t, err)

	err = ts.Validate(token, ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session2",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
	})
	require.ErrorIs(t, err, ErrTokenMismatch)
}

func TestConfirmationToken_DifferentSecret(t *testing.T) {
	ts1 := NewTokenService("secret-one", 2*time.Minute)
	ts2 := NewTokenService("secret-two", 2*time.Minute)
	params := map[string]string{"container_id": "abc123"}

	token, err := ts1.Generate(ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts1.ParamsHash(params),
		RiskLevel:  "high",
		Expiry:     time.Now().Add(2 * time.Minute),
	})
	require.NoError(t, err)

	err = ts2.Validate(token, ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts1.ParamsHash(params),
		RiskLevel:  "high",
	})
	require.ErrorIs(t, err, ErrTokenInvalid)
}

func TestConfirmationToken_InvalidBase64(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)

	err := ts.Validate("not-valid-base64!!!", ActionConfirmationToken{
		UserID:    "user1",
		SessionID: "session1",
		Action:    "stop_container",
	})
	require.ErrorIs(t, err, ErrTokenInvalid)
}

func TestConfirmationToken_MalformedToken(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)

	err := ts.Validate("dGVzdA", ActionConfirmationToken{
		UserID:    "user1",
		SessionID: "session1",
		Action:    "stop_container",
	})
	require.ErrorIs(t, err, ErrTokenInvalid)
}

func TestConfirmationToken_ParamsHash_Consistency(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)
	params := map[string]string{"b": "2", "a": "1", "c": "3"}

	hash1 := ts.ParamsHash(params)
	hash2 := ts.ParamsHash(params)
	require.Equal(t, hash1, hash2, "same params should produce same hash regardless of iteration order")

	params2 := map[string]string{"a": "1", "b": "2", "c": "3"}
	hash3 := ts.ParamsHash(params2)
	require.Equal(t, hash1, hash3, "different key order should produce same hash")

	params3 := map[string]string{"a": "1", "b": "2"}
	hash4 := ts.ParamsHash(params3)
	require.NotEqual(t, hash1, hash4, "different params should produce different hash")
}

func TestConfirmationToken_ParamsHash_NilParams(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)

	hash1 := ts.ParamsHash(nil)
	hash2 := ts.ParamsHash(nil)
	require.Equal(t, hash1, hash2, "nil params should produce consistent hash")
}

func TestConfirmationToken_ParamsHash_EmptyParams(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)

	hash := ts.ParamsHash(map[string]string{})
	require.NotEmpty(t, hash)
}

func TestConfirmationToken_DefaultSecret(t *testing.T) {
	ts := NewTokenService("", 2*time.Minute)
	require.NotNil(t, ts)
}

func TestConfirmationToken_DefaultTTL(t *testing.T) {
	ts := NewTokenService("test-secret", 0)
	params := map[string]string{"key": "value"}

	token, err := ts.Generate(ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "action1",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "low",
	})
	require.NoError(t, err)
	require.NotEmpty(t, token)

	time.Sleep(100 * time.Millisecond)

	err = ts.Validate(token, ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "action1",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "low",
	})
	require.NoError(t, err)
}

func TestConfirmationToken_ExpiryInThePast(t *testing.T) {
	ts := NewTokenService("test-secret-key", 2*time.Minute)
	params := map[string]string{"key": "value"}

	token := ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
		Expiry:     time.Now().Add(-1 * time.Hour),
	}

	_, err := ts.Generate(token)
	require.NoError(t, err)

	generatedToken, err := ts.Generate(token)
	require.NoError(t, err)

	err = ts.Validate(generatedToken, ActionConfirmationToken{
		UserID:     "user1",
		SessionID:  "session1",
		Action:     "stop_container",
		ParamsHash: ts.ParamsHash(params),
		RiskLevel:  "high",
	})
	require.ErrorIs(t, err, ErrTokenExpired)
}

func TestConfirmationToken_ComputeParamsHash(t *testing.T) {
	params := map[string]string{"key": "value"}
	hash := ComputeParamsHash(params)
	require.NotEmpty(t, hash)
	require.Len(t, hash, 64)
}
