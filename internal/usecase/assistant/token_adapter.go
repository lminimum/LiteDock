package assistant

import (
	"time"

	"github.com/lminimum/LiteDock/internal/action"
)

// TokenValidatorAdapter wraps TokenService to satisfy action.TokenValidator.
type TokenValidatorAdapter struct {
	inner *TokenService
	ttl   time.Duration
}

// NewTokenValidatorAdapter creates an adapter that bridges TokenService to action.TokenValidator.
func NewTokenValidatorAdapter(ts *TokenService, ttl time.Duration) *TokenValidatorAdapter {
	if ttl == 0 {
		ttl = 2 * time.Minute
	}
	return &TokenValidatorAdapter{inner: ts, ttl: ttl}
}

// Validate implements action.TokenValidator.
func (a *TokenValidatorAdapter) Validate(tokenStr string, params action.TokenParams) error {
	return a.inner.Validate(tokenStr, ActionConfirmationToken{
		UserID:     params.UserID,
		SessionID:  params.SessionID,
		Action:     params.Action,
		ParamsHash: params.ParamsHash,
		RiskLevel:  params.RiskLevel,
		Expiry:     time.Now().Add(a.ttl),
	})
}
