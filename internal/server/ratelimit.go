package server

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter tracks request counts per client identifier.
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string]*rateInfo
	limit    int
	window   time.Duration
}

// rateInfo tracks request count and window start time for a client.
type rateInfo struct {
	count       int
	windowStart time.Time
}

// NewRateLimiter creates a new rate limiter with the specified requests per minute.
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rateLimiter := &RateLimiter{
		requests: make(map[string]*rateInfo),
		limit:    requestsPerMinute,
		window:   time.Minute,
	}

	// Start cleanup goroutine to remove old entries
	go rateLimiter.cleanup()

	return rateLimiter
}

// cleanup periodically removes expired rate limit entries.
func (rateLimiter *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rateLimiter.mu.Lock()
		now := time.Now()
		for clientID, info := range rateLimiter.requests {
			if now.Sub(info.windowStart) > rateLimiter.window {
				delete(rateLimiter.requests, clientID)
			}
		}
		rateLimiter.mu.Unlock()
	}
}

// Allow checks if a request from the given client should be allowed.
func (rateLimiter *RateLimiter) Allow(clientID string) bool {
	if rateLimiter.limit == 0 {
		return true
	}

	rateLimiter.mu.Lock()
	defer rateLimiter.mu.Unlock()

	now := time.Now()
	info, exists := rateLimiter.requests[clientID]

	if !exists || now.Sub(info.windowStart) > rateLimiter.window {
		rateLimiter.requests[clientID] = &rateInfo{
			count:       1,
			windowStart: now,
		}
		return true
	}

	if info.count >= rateLimiter.limit {
		return false
	}

	info.count++
	return true
}

// RetryAfter returns how long a client should wait before retrying.
func (rateLimiter *RateLimiter) RetryAfter(clientID string) time.Duration {
	if rateLimiter.limit == 0 {
		return 0
	}

	rateLimiter.mu.RLock()
	defer rateLimiter.mu.RUnlock()

	info, exists := rateLimiter.requests[clientID]
	if !exists {
		return 0
	}

	remaining := rateLimiter.window - time.Since(info.windowStart)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// getClientIP extracts the client IP from the request.
// Uses RemoteAddr only and does not trust proxy headers.
func getClientIP(request *http.Request) string {
	ip, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return request.RemoteAddr
	}
	return ip
}
