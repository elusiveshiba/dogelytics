package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dogeorg/dogelytics/internal/config"
	"github.com/dogeorg/dogelytics/internal/credentials"
	"golang.org/x/crypto/bcrypt"
)

const apiKeySHA256Prefix = credentials.APIKeySHA256Prefix

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
		clientID := "api-auth:" + s.getClientIP(r)
		if s.ipLimiter != nil && !s.ipLimiter.Allow(clientID) {
			w.Header().Set("Retry-After", "60")
			s.sendError(w, http.StatusTooManyRequests, "rate-limit-exceeded", "Too many API key authentication attempts")
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
		k, err := s.authStore.GetAPIKeyByKID(r.Context(), kid)
		if err != nil || k.ID == "" {
			sendInvalidAPIKeyError(w)
			return
		}
		// Expired or revoked?
		if (k.ExpiresAt.Valid && time.Now().After(k.ExpiresAt.Time)) || (k.RevokedAt.Valid) {
			sendInvalidAPIKeyError(w)
			return
		}
		valid, legacy := verifyAPIKeySecret(k.SecretHash, secret)
		if !valid {
			sendInvalidAPIKeyError(w)
			return
		}
		if legacy {
			if err := s.authStore.UpdateAPIKeySecretHash(r.Context(), k.KID, k.SecretHash, hashAPIKeySecret(secret)); err != nil {
				log.Printf("[Dogelytics] migrate API key hash: %v", err)
			}
		}
		ctx := withAPIKey(r.Context(), k)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func hashAPIKeySecret(secret string) string {
	return credentials.HashAPIKeySecret(secret)
}

func verifyAPIKeySecret(encodedHash, secret string) (valid bool, legacy bool) {
	if strings.HasPrefix(encodedHash, apiKeySHA256Prefix) {
		expected, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(encodedHash, apiKeySHA256Prefix))
		if err != nil || len(expected) != sha256.Size {
			return false, false
		}
		actual := sha256.Sum256([]byte(secret))
		return subtle.ConstantTimeCompare(expected, actual[:]) == 1, false
	}
	return bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(secret)) == nil, true
}

func sendInvalidAPIKeyError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(config.ErrorResponse{
		Error:   "invalid-api-key",
		Message: "Invalid, expired, or revoked API key",
	})
}
