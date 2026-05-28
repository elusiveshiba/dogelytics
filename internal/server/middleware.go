package server

import (
	"fmt"
	"math"
	"net/http"
	"time"
)

// RateLimitMiddleware applies per-key or per-IP rate limiting to API endpoints.
func (s *Server) RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if key, ok := getAPIKey(r.Context()); ok && key.KID != "" {
			if s.apiLimiter != nil && !s.apiLimiter.Allow(key.KID) {
				s.sendError(
					w,
					http.StatusTooManyRequests,
					"rate-limit-exceeded",
					rateLimitMessage(s.config.APIKeyRateLimit, s.apiLimiter.RetryAfter(key.KID)),
				)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		if s.ipLimiter != nil {
			clientIP := getClientIP(r)
			if !s.ipLimiter.Allow(clientIP) {
				s.sendError(
					w,
					http.StatusTooManyRequests,
					"rate-limit-exceeded",
					rateLimitMessage(s.config.RateLimit, s.ipLimiter.RetryAfter(clientIP)),
				)
				return
			}
		}

		next.ServeHTTP(w, r)
	}
}

func rateLimitMessage(limit int, retryAfter time.Duration) string {
	minutes := int(math.Ceil(retryAfter.Minutes()))
	if minutes < 1 {
		minutes = 1
	}
	return fmt.Sprintf("Rate limit exceeded. Maximum %d requests per minute. Try again in %d minute(s).", limit, minutes)
}
