package assistant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRateLimiter_AllowWithinLimit(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Close()

	for i := 0; i < _rateLimitMaxRequests; i++ {
		allowed := rl.Allow("session1")
		require.True(t, allowed, "request %d should be allowed", i+1)
	}
}

func TestRateLimiter_BlockExcessRequests(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Close()

	for i := 0; i < _rateLimitMaxRequests; i++ {
		rl.Allow("session1")
	}

	allowed := rl.Allow("session1")
	require.False(t, allowed, "request beyond limit should be blocked")
}

func TestRateLimiter_PerSession(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Close()

	for i := 0; i < _rateLimitMaxRequests; i++ {
		rl.Allow("session1")
	}

	allowed := rl.Allow("session2")
	require.True(t, allowed, "different session should not be affected")
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	origWindow := _rateLimitWindow
	origMax := _rateLimitMaxRequests
	_rateLimitWindow = 100 * time.Millisecond
	_rateLimitMaxRequests = 2
	defer func() {
		_rateLimitWindow = origWindow
		_rateLimitMaxRequests = origMax
	}()

	rl := NewRateLimiter()
	defer rl.Close()

	rl.Allow("session1")
	rl.Allow("session1")
	require.False(t, rl.Allow("session1"))

	time.Sleep(150 * time.Millisecond)

	require.True(t, rl.Allow("session1"), "requests should be allowed after window expires")
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Close()

	var wg sync.WaitGroup
	sessionCount := 10
	requestsPerSession := 20

	for s := 0; s < sessionCount; s++ {
		sessionID := "session" + string(rune('0'+s))
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < requestsPerSession; i++ {
				rl.Allow(sessionID)
			}
		}()
	}

	wg.Wait()
}

func TestRateLimiter_MultipleSessions(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Close()

	sessions := []string{"s1", "s2", "s3", "s4", "s5"}

	for _, s := range sessions {
		for i := 0; i < _rateLimitMaxRequests; i++ {
			rl.Allow(s)
		}
	}

	for _, s := range sessions {
		allowed := rl.Allow(s)
		require.False(t, allowed, "session %s should be blocked", s)
	}
}

func TestRetryOnServerError(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[{"message":{"content":"success"}}]}`))
	}))
	defer server.Close()

	client := NewLLMClient(server.URL, "test-key", "test-model")
	resp, err := client.ChatCompletion(context.Background(),
		[]ChatMessage{{Role: "user", Content: "Hi"}}, nil)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "success", resp.Content)
	require.Equal(t, 3, attempt, "should have retried 2 times")
}

func TestNoRetryOnClientError(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewLLMClient(server.URL, "test-key", "test-model")

	_, err := client.ChatCompletion(context.Background(),
		[]ChatMessage{{Role: "user", Content: "Hi"}}, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "API error (status 400)")
	require.Equal(t, 1, attempt, "should not have retried on client error")
}

func TestErrorMessages(t *testing.T) {
	testCases := []struct {
		name     string
		errMsg   string
		expected string
	}{
		{"timeout", "timeout: context deadline exceeded", MsgTimeout},
		{"auth failure", "API error (status 401): unauthorized", MsgAuthFailed},
		{"rate limited", "API error (status 429): rate limited", MsgRateLimited},
		{"server error", "server error: status 500: internal error", MsgTimeout},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := &testError{msg: tc.errMsg}
			result := MapErrorToUserMessage(err)
			require.Equal(t, tc.expected, result)
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
