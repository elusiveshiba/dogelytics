package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestRateLimiter_Middleware(t *testing.T) {
	rateLimiter := NewRateLimiter(2)

	// Create a test handler
	handler := rateLimiter.Middleware(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("OK"))
	})

	// First request - should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	recorder1 := httptest.NewRecorder()
	handler(recorder1, req1)
	if recorder1.Code != http.StatusOK {
		t.Errorf("First request should succeed, got status %d", recorder1.Code)
	}

	// Second request - should succeed
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	recorder2 := httptest.NewRecorder()
	handler(recorder2, req2)
	if recorder2.Code != http.StatusOK {
		t.Errorf("Second request should succeed, got status %d", recorder2.Code)
	}

	// Third request - should fail with 429
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:1234"
	recorder3 := httptest.NewRecorder()
	handler(recorder3, req3)
	if recorder3.Code != http.StatusTooManyRequests {
		t.Errorf("Third request should fail with 429, got status %d", recorder3.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expectedIP string
	}{
		{
			name:       "X-Forwarded-For header",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.1"},
			expectedIP: "203.0.113.1",
		},
		{
			name:       "X-Real-IP header",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Real-IP": "203.0.113.2"},
			expectedIP: "203.0.113.2",
		},
		{
			name:       "RemoteAddr fallback",
			remoteAddr: "203.0.113.3:1234",
			headers:    map[string]string{},
			expectedIP: "203.0.113.3",
		},
		{
			name:       "X-Forwarded-For takes precedence",
			remoteAddr: "10.0.0.1:1234",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.4",
				"X-Real-IP":       "203.0.113.5",
			},
			expectedIP: "203.0.113.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/test", nil)
			request.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				request.Header.Set(k, v)
			}

			ip := getClientIP(request)
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

