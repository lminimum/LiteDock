package assistant

import (
	"sync"
	"time"
)

var (
	_rateLimitMaxRequests = 5
	_rateLimitWindow      = 10 * time.Second
)

type sessionLimiter struct {
	mu       sync.Mutex
	requests []time.Time
}

func newSessionLimiter() *sessionLimiter {
	return &sessionLimiter{}
}

func (sl *sessionLimiter) allow() bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-_rateLimitWindow)

	var validRequests []time.Time
	for _, t := range sl.requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	if len(validRequests) >= _rateLimitMaxRequests {
		sl.requests = validRequests
		return false
	}

	sl.requests = append(validRequests, now)
	return true
}

type RateLimiter struct {
	mu       sync.RWMutex
	sessions map[string]*sessionLimiter
	cleanup  *time.Ticker
	done     chan struct{}
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		sessions: make(map[string]*sessionLimiter),
		cleanup:  time.NewTicker(5 * time.Minute),
		done:     make(chan struct{}),
	}
	go rl.cleanupExpired()
	return rl
}

func (rl *RateLimiter) Allow(sessionID string) bool {
	rl.mu.RLock()
	limiter, exists := rl.sessions[sessionID]
	rl.mu.RUnlock()

	if exists {
		return limiter.allow()
	}

	rl.mu.Lock()
	limiter, exists = rl.sessions[sessionID]
	if !exists {
		limiter = newSessionLimiter()
		rl.sessions[sessionID] = limiter
	}
	rl.mu.Unlock()

	return limiter.allow()
}

func (rl *RateLimiter) cleanupExpired() {
	for {
		select {
		case <-rl.cleanup.C:
			rl.mu.Lock()
			for sessionID, limiter := range rl.sessions {
				limiter.mu.Lock()
				if len(limiter.requests) == 0 {
					delete(rl.sessions, sessionID)
				}
				limiter.mu.Unlock()
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

func (rl *RateLimiter) Close() {
	close(rl.done)
	rl.cleanup.Stop()
}
