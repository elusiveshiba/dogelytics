package server

import (
	"fmt"
	"net/http"
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
					fmt.Sprintf("Rate limit exceeded. Maximum %d requests per minute.", s.config.APIKeyRateLimit),
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
					fmt.Sprintf("Rate limit exceeded. Maximum %d requests per minute.", s.config.RateLimit),
				)
				return
			}
		}

		next.ServeHTTP(w, r)
	}
}
