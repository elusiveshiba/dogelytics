package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dogeorg/dogelytics/internal/config"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Create rate limiter with 3 requests per minute
	rateLimiter := NewRateLimiter(3)

	ip := "192.168.1.1"

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !rateLimiter.Allow(ip) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if rateLimiter.Allow(ip) {
		t.Error("4th request should be denied")
	}

	// Different IP should be allowed
	if !rateLimiter.Allow("192.168.1.2") {
		t.Error("Request from different IP should be allowed")
	}
}

func TestRateLimiter_Disabled(t *testing.T) {
	// Create rate limiter with 0 (disabled)
	rateLimiter := NewRateLimiter(0)

	ip := "192.168.1.1"

	// All requests should be allowed when disabled
	for i := 0; i < 100; i++ {
		if !rateLimiter.Allow(ip) {
			t.Errorf("Request %d should be allowed when rate limiting is disabled", i+1)
		}
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	// This is a short test that doesn't wait a full minute
	// Instead we'll verify the logic works correctly
	rateLimiter := NewRateLimiter(2)

	ip := "192.168.1.1"

	// First 2 requests
	if !rateLimiter.Allow(ip) {
		t.Error("First request should be allowed")
	}
	if !rateLimiter.Allow(ip) {
		t.Error("Second request should be allowed")
	}

	// Third should be denied
	if rateLimiter.Allow(ip) {
		t.Error("Third request should be denied")
	}

	// Manually expire the window by modifying the stored time
	rateLimiter.mu.Lock()
	if info, exists := rateLimiter.requests[ip]; exists {
		info.windowStart = time.Now().Add(-2 * time.Minute)
	}
	rateLimiter.mu.Unlock()

	// Now request should be allowed again
	if !rateLimiter.Allow(ip) {
		t.Error("Request should be allowed after window expiry")
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		expectedIP string
	}{
		{
			name:       "RemoteAddr host and port",
			remoteAddr: "203.0.113.3:1234",
			expectedIP: "203.0.113.3",
		},
		{
			name:       "RemoteAddr fallback on split failure",
			remoteAddr: "203.0.113.9",
			expectedIP: "203.0.113.9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/test", nil)
			request.RemoteAddr = tt.remoteAddr
			ip := getClientIP(request)
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

func TestServerRateLimitMiddleware(t *testing.T) {
	srv := &Server{
		config:    &config.Config{RateLimit: 1},
		ipLimiter: NewRateLimiter(1),
	}

	handler := srv.RateLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req1 := httptest.NewRequest(http.MethodGet, "/balance", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	rec1 := httptest.NewRecorder()
	handler(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request should succeed, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/balance", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	rec2 := httptest.NewRecorder()
	handler(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request should be rate limited, got %d", rec2.Code)
	}

	var resp config.ErrorResponse
	if err := json.NewDecoder(rec2.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error != "rate-limit-exceeded" {
		t.Fatalf("unexpected error code %q", resp.Error)
	}
}
