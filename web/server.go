package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dogeorg/doge"
	"github.com/dogeorg/dogelytics/spec"
	"github.com/dogeorg/dogelytics/store"
)

// Server handles HTTP requests for balance and health endpoints
type Server struct {
	store      *store.Store
	config     *spec.Config
	ipLimiter  *RateLimiter
	apiLimiter *RateLimiter
}

// NewServer creates a new Server instance
func NewServer(store *store.Store, config *spec.Config, ipLimiter *RateLimiter, apiLimiter *RateLimiter) *Server {
	return &Server{
		store:      store,
		config:     config,
		ipLimiter:  ipLimiter,
		apiLimiter: apiLimiter,
	}
}

// HandleHealth responds to health check requests
func (server *Server) HandleHealth(writer http.ResponseWriter, request *http.Request) {
	server.setCORSHeaders(writer, request)

	if request.Method == http.MethodOptions {
		writer.WriteHeader(http.StatusOK)
		return
	}

	if request.Method != http.MethodGet {
		server.sendError(writer, http.StatusMethodNotAllowed, "method-not-allowed", "Only GET and OPTIONS methods are allowed")
		return
	}

	// Test database connection
	_, err := server.store.GetResumePoint()
	if err != nil {
		server.sendError(writer, http.StatusInternalServerError, "database-error", fmt.Sprintf("Database error: %v", err))
		return
	}

	height, err := server.store.GetCurrentHeight()
	if err != nil {
		server.sendError(writer, http.StatusInternalServerError, "database-error", fmt.Sprintf("Database error: %v", err))
		return
	}

	// Log health check with current height for progress tracking
	log.Printf("[Dogelytics] Health check: indexed height = %d", height)

	response := map[string]interface{}{
		"ok":     true,
		"height": height,
	}

	server.sendJSON(writer, http.StatusOK, response)
}

// HandleBalance responds to balance query requests
func (server *Server) HandleBalance(writer http.ResponseWriter, request *http.Request) {
	// API endpoint: set CORS and continue
	server.setCORSHeaders(writer, request)

	if request.Method == http.MethodOptions {
		writer.WriteHeader(http.StatusOK)
		return
	}

	if request.Method != http.MethodGet {
		server.sendError(writer, http.StatusMethodNotAllowed, "method-not-allowed", "Only GET and OPTIONS methods are allowed")
		return
	}

	// Get address parameter
	address := request.URL.Query().Get("address")

	// Log balance request with client IP
	clientIP := getClientIP(request)
	log.Printf("[Dogelytics] Balance request from %s for address: %s", clientIP, address)

	if address == "" {
		server.sendError(writer, http.StatusBadRequest, "missing-parameter", "Missing 'address' parameter in query string")
		return
	}

	// Decode Dogecoin address
	pubkeyHash, err := doge.Base58DecodeCheck(address)
	if err != nil {
		server.sendError(writer, http.StatusBadRequest, "invalid-address", "Invalid Dogecoin address format")
		return
	}

	if len(pubkeyHash) != 21 {
		server.sendError(writer, http.StatusBadRequest, "invalid-address", "Invalid Dogecoin address length")
		return
	}

	// Extract version byte and hash
	kind := scriptTypeFromVersionByte(pubkeyHash[0])
	if kind == doge.ScriptTypeNone {
		server.sendError(writer, http.StatusBadRequest, "unsupported-address", "Unsupported address type")
		return
	}

	hash := pubkeyHash[1:]

	// Get balance from database using the store's context
	storeWithCtx := server.store.WithCtx(request.Context())
	balance, err := storeWithCtx.GetBalance(kind, hash, server.config.Confirmations)
	if err != nil {
		server.sendError(writer, http.StatusInternalServerError, "database-error", fmt.Sprintf("Failed to get balance: %v", err))
		return
	}

	// Calculate current balance
	balance.Current = balance.Available + balance.Incoming

	// Send response
	response := spec.BalanceResponse{
		Incoming:  balance.Incoming,
		Available: balance.Available,
		Outgoing:  balance.Outgoing,
		Current:   balance.Current,
	}

	server.sendJSON(writer, http.StatusOK, response)
}

// RateLimitMiddleware applies per-key or per-IP rate limiting
func (server *Server) RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Exempt health and auth/UI related endpoints
		if r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/login") || strings.HasPrefix(r.URL.Path, "/register") || strings.HasPrefix(r.URL.Path, "/auth/") {
			next.ServeHTTP(w, r)
			return
		}
		// If we have a valid API key in context, rate limit by key
		if k, ok := getAPIKey(r.Context()); ok && k.KID != "" {
			if server.apiLimiter != nil && !server.apiLimiter.Allow(k.KID) {
				writer := w
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusTooManyRequests)
				response := spec.ErrorResponse{
					Error:   "rate-limit-exceeded",
					Message: fmt.Sprintf("Rate limit exceeded. Maximum %d requests per minute.", server.config.APIKeyRateLimit),
				}
				_ = json.NewEncoder(writer).Encode(response)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		// Otherwise, rate limit by client IP
		if server.ipLimiter != nil {
			ip := getClientIP(r)
			if !server.ipLimiter.Allow(ip) {
				writer := w
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusTooManyRequests)
				response := spec.ErrorResponse{
					Error:   "rate-limit-exceeded",
					Message: fmt.Sprintf("Rate limit exceeded. Maximum %d requests per minute.", server.config.RateLimit),
				}
				_ = json.NewEncoder(writer).Encode(response)
				return
			}
		}
		next.ServeHTTP(w, r)
	}
}

// setCORSHeaders sets appropriate CORS headers on the response
func (server *Server) setCORSHeaders(writer http.ResponseWriter, request *http.Request) {
	origin := server.config.CorsOrigin
	if origin == "*" {
		// If wildcard, use the request origin
		if requestOrigin := request.Header.Get("Origin"); requestOrigin != "" {
			origin = requestOrigin
		}
	}

	writer.Header().Set("Access-Control-Allow-Origin", origin)
	writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Api-Key")
	writer.Header().Set("Access-Control-Max-Age", "3600")
}

// sendJSON sends a JSON response with the given status code
func (server *Server) sendJSON(writer http.ResponseWriter, status int, data interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if err := json.NewEncoder(writer).Encode(data); err != nil {
		log.Printf("[Dogelytics] failed to encode JSON response: %v", err)
	}
}

// sendError sends a JSON error response
func (server *Server) sendError(writer http.ResponseWriter, status int, errorType string, message string) {
	response := spec.ErrorResponse{
		Error:   errorType,
		Message: message,
	}
	server.sendJSON(writer, status, response)
}

// scriptTypeFromVersionByte converts an address version byte to a script type
func scriptTypeFromVersionByte(version byte) doge.ScriptType {
	switch version {
	case doge.DogeMainNetChain.P2PKH_Address_Prefix,
		doge.DogeTestNetChain.P2PKH_Address_Prefix,
		doge.DogeRegTestChain.P2PKH_Address_Prefix:
		return doge.ScriptTypeP2PKH
	case doge.DogeMainNetChain.P2SH_Address_Prefix,
		doge.DogeTestNetChain.P2SH_Address_Prefix,
		doge.DogeRegTestChain.P2SH_Address_Prefix:
		return doge.ScriptTypeP2SH
	case doge.DogeMainNetChain.PKey_Prefix,
		doge.DogeTestNetChain.PKey_Prefix,
		doge.DogeRegTestChain.PKey_Prefix:
		return doge.ScriptTypeP2PK
	}
	return doge.ScriptTypeNone
}
