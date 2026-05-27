package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dogeorg/doge"
	"github.com/dogeorg/dogelytics/internal/config"
)

// HandleHealth responds to health check requests.
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "method-not-allowed", "Only GET and OPTIONS methods are allowed")
		return
	}

	height, err := s.indexerStore.CurrentHeight(r.Context())
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "database-error", fmt.Sprintf("Database error: %v", err))
		return
	}

	s.sendJSON(w, http.StatusOK, HealthResponse{
		OK:     true,
		Height: height,
	})
}

// HandleBalance responds to balance query requests.
func (s *Server) HandleBalance(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "method-not-allowed", "Only GET and OPTIONS methods are allowed")
		return
	}

	address := r.URL.Query().Get("address")
	if address == "" {
		s.sendError(w, http.StatusBadRequest, "missing-parameter", "Missing 'address' parameter in query string")
		return
	}

	pubkeyHash, err := doge.Base58DecodeCheck(address)
	if err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid-address", "Invalid Dogecoin address format")
		return
	}

	if len(pubkeyHash) != 21 {
		s.sendError(w, http.StatusBadRequest, "invalid-address", "Invalid Dogecoin address length")
		return
	}

	scriptType := scriptTypeFromVersionByte(pubkeyHash[0])
	if scriptType == doge.ScriptTypeNone {
		s.sendError(w, http.StatusBadRequest, "unsupported-address", "Unsupported address type")
		return
	}

	balance, err := s.indexerStore.GetBalance(r.Context(), scriptType, pubkeyHash[1:], s.config.Confirmations)
	if err != nil {
		s.logBalanceRequest(r, address, false)
		s.sendError(w, http.StatusInternalServerError, "database-error", fmt.Sprintf("Failed to get balance: %v", err))
		return
	}

	s.logBalanceRequest(r, address, true)
	s.sendJSON(w, http.StatusOK, config.BalanceResponse{
		Incoming:  balance.Incoming,
		Available: balance.Available,
		Outgoing:  balance.Outgoing,
		Current:   balance.Current,
	})
}

func (s *Server) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", s.config.CorsOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Max-Age", "3600")
}

func (s *Server) sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[Dogelytics] failed to encode JSON response: %v", err)
	}
}

func (s *Server) sendError(w http.ResponseWriter, status int, errorType string, message string) {
	s.sendJSON(w, status, config.ErrorResponse{
		Error:   errorType,
		Message: message,
	})
}

func (s *Server) logBalanceRequest(r *http.Request, address string, success bool) {
	if s.authStore == nil {
		return
	}

	apiKey := ""
	if key, ok := getAPIKey(r.Context()); ok {
		apiKey = key.KID
	}

	authStore := s.authStore.WithCtx(r.Context())
	if err := authStore.LogRequest(getClientIP(r), apiKey, address, success); err != nil {
		log.Printf("[Dogelytics] failed to log request: %v", err)
	}
}

// scriptTypeFromVersionByte converts an address version byte to a script type.
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
