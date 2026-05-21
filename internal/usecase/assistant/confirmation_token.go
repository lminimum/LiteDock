package assistant

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	ErrTokenExpired  = errors.New("confirmation token expired")
	ErrTokenMismatch = errors.New("confirmation token mismatch")
	ErrTokenInvalid  = errors.New("confirmation token invalid")
)

type ActionConfirmationToken struct {
	UserID     string
	SessionID  string
	Action     string
	ParamsHash string
	RiskLevel  string
	Expiry     time.Time
}

type TokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenService(secret string, ttl time.Duration) *TokenService {
	if secret == "" {
		secret = "ai-confirmation-dev-secret-do-not-use-in-production"
	}
	if ttl == 0 {
		ttl = 2 * time.Minute
	}
	return &TokenService{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (ts *TokenService) Generate(token ActionConfirmationToken) (string, error) {
	if token.Expiry.IsZero() {
		token.Expiry = time.Now().Add(ts.ttl)
	}

	data := ts.buildSignatureData(token.UserID, token.SessionID, token.Action, token.ParamsHash, token.RiskLevel, token.Expiry.Unix())
	signature := ts.sign(data)

	tokenStr := base64.RawURLEncoding.EncodeToString([]byte(data + "." + signature))
	return tokenStr, nil
}

func (ts *TokenService) Validate(tokenStr string, params ActionConfirmationToken) error {
	decoded, err := base64.RawURLEncoding.DecodeString(tokenStr)
	if err != nil {
		return ErrTokenInvalid
	}

	parts := strings.SplitN(string(decoded), ".", 2)
	if len(parts) != 2 {
		return ErrTokenInvalid
	}

	data, providedSig := parts[0], parts[1]
	expectedSig := ts.sign(data)

	if !hmac.Equal([]byte(providedSig), []byte(expectedSig)) {
		return ErrTokenInvalid
	}

	parsed, err := ts.parseSignatureData(data)
	if err != nil {
		return ErrTokenInvalid
	}

	if time.Now().Unix() > parsed.expiry {
		return ErrTokenExpired
	}

	if parsed.userID != params.UserID ||
		parsed.sessionID != params.SessionID ||
		parsed.action != params.Action ||
		parsed.paramsHash != params.ParamsHash ||
		parsed.riskLevel != params.RiskLevel {
		return ErrTokenMismatch
	}

	return nil
}

func (ts *TokenService) ParamsHash(params map[string]string) string {
	if params == nil {
		params = make(map[string]string)
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
	}

	hash := sha256.Sum256([]byte(sb.String()))
	return fmt.Sprintf("%x", hash)
}

type signatureData struct {
	userID     string
	sessionID  string
	action     string
	paramsHash string
	riskLevel  string
	expiry     int64
}

func (ts *TokenService) buildSignatureData(userID, sessionID, action, paramsHash, riskLevel string, expiry int64) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d", userID, sessionID, action, paramsHash, riskLevel, expiry)
}

func (ts *TokenService) sign(data string) string {
	h := hmac.New(sha256.New, ts.secret)
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func (ts *TokenService) parseSignatureData(data string) (*signatureData, error) {
	parts := strings.Split(data, "|")
	if len(parts) != 6 {
		return nil, errors.New("invalid signature data format")
	}

	var expiry int64
	if _, err := fmt.Sscanf(parts[5], "%d", &expiry); err != nil {
		return nil, errors.New("invalid expiry value")
	}

	return &signatureData{
		userID:     parts[0],
		sessionID:  parts[1],
		action:     parts[2],
		paramsHash: parts[3],
		riskLevel:  parts[4],
		expiry:     expiry,
	}, nil
}

func ComputeParamsHash(params map[string]string) string {
	return (&TokenService{}).ParamsHash(params)
}
