package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/dogeorg/dogelytics/internal/config"
)

// RateLimiter tracks request counts per IP address
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string]*ipRateInfo
	limit    int
	window   time.Duration
}

// ipRateInfo tracks request count and window start time for an IP
type ipRateInfo struct {
	count       int
	windowStart time.Time
}

// NewRateLimiter creates a new rate limiter with the specified requests per minute
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rateLimiter := &RateLimiter{
		requests: make(map[string]*ipRateInfo),
		limit:    requestsPerMinute,
		window:   time.Minute,
	}

	// Start cleanup goroutine to remove old entries
	go rateLimiter.cleanup()

	return rateLimiter
}

// cleanup periodically removes expired rate limit entries
func (rateLimiter *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rateLimiter.mu.Lock()
		now := time.Now()
		for ip, info := range rateLimiter.requests {
			if now.Sub(info.windowStart) > rateLimiter.window {
				delete(rateLimiter.requests, ip)
			}
		}
		rateLimiter.mu.Unlock()
	}
}

// Allow checks if a request from the given IP should be allowed
func (rateLimiter *RateLimiter) Allow(ip string) bool {
	if rateLimiter.limit == 0 {
		return true // Rate limiting disabled
	}

	rateLimiter.mu.Lock()
	defer rateLimiter.mu.Unlock()

	now := time.Now()
	info, exists := rateLimiter.requests[ip]

	if !exists || now.Sub(info.windowStart) > rateLimiter.window {
		// New IP or window expired, reset counter
		rateLimiter.requests[ip] = &ipRateInfo{
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

// Middleware wraps an HTTP handler with rate limiting
func (rateLimiter *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Extract IP address
		ip := getClientIP(request)

		if !rateLimiter.Allow(ip) {
			log.Printf("[Dogelytics] Rate limit exceeded for IP: %s", ip)
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusTooManyRequests)
			response := config.ErrorResponse{
				Error:   "rate-limit-exceeded",
				Message: fmt.Sprintf("Rate limit exceeded. Maximum %d requests per minute.", rateLimiter.limit),
			}
			json.NewEncoder(writer).Encode(response)
			return
		}

		next(writer, request)
	}
}

// getClientIP extracts the client IP from the request
// Uses RemoteAddr only - doesn't trust proxy headers to prevent spoofing
func getClientIP(request *http.Request) string {
	ip, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return request.RemoteAddr
	}
	return ip
}
