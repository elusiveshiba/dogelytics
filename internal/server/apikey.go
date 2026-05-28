package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dogeorg/dogelytics/internal/config"
)

// parseBearerOrHeader extracts token from Authorization: Bearer or X-Api-Key
func parseBearerOrHeader(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:]), true
	}
	if v := r.Header.Get("X-Api-Key"); v != "" {
		return strings.TrimSpace(v), true
	}
	return "", false
}

// parseToken expects format dglk_<kid>.<secret>
func parseToken(tok string) (string, string, bool) {
	if !strings.HasPrefix(tok, "dglk_") {
		return "", "", false
	}
	body := strings.TrimPrefix(tok, "dglk_")
	parts := strings.Split(body, ".")
	if len(parts) != 2 {
		return "", "", false
	}
	kid := parts[0]
	secret := parts[1]
	if kid == "" || secret == "" {
		return "", "", false
	}
	return kid, secret, true
}

// APIKeyAuthMiddleware validates API key if present and attaches it to context
func (s *Server) APIKeyAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok, ok := parseBearerOrHeader(r)
		if !ok || tok == "" {
			next.ServeHTTP(w, r)
			return
		}
		if !s.hasAuthStore() {
			sendInvalidAPIKeyError(w)
			return
		}
		kid, secret, ok := parseToken(tok)
		if !ok {
			sendInvalidAPIKeyError(w)
			return
		}
		k, err := s.authStore.GetAPIKeyByKID(kid)
		if err != nil || k.ID == "" {
			sendInvalidAPIKeyError(w)
			return
		}
		// Expired or revoked?
		if (k.ExpiresAt.Valid && time.Now().After(k.ExpiresAt.Time)) || (k.RevokedAt.Valid) {
			sendInvalidAPIKeyError(w)
			return
		}
		// Verify secret
		if checkPassword(k.SecretHash, secret) != nil {
			sendInvalidAPIKeyError(w)
			return
		}
		ctx := withAPIKey(r.Context(), k)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func sendInvalidAPIKeyError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(config.ErrorResponse{
		Error:   "invalid-api-key",
		Message: "Invalid, expired, or revoked API key",
	})
}
