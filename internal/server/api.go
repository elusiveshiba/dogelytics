package server

import (
	"encoding/json"
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

	response := HealthResponse{
		OK: true,
	}
	if progress, ok := s.loadChainProgress(r.Context()); ok {
		response.IndexerHeight = progress.IndexerHeight
		response.CoreBlocksHeight = progress.CoreBlocksHeight
		response.CoreHeadersHeight = progress.CoreHeadersHeight
		response.CoreSyncUpdatedAt = progress.CoreSyncUpdatedAt
	} else {
		height, err := s.indexerStore.CurrentHeight(r.Context())
		if err != nil {
			log.Printf("[Dogelytics] indexer health request failed: %v", err)
			s.sendError(w, http.StatusServiceUnavailable, "indexer-unavailable", "Indexer sync data is unavailable")
			return
		}
		response.IndexerHeight = height
	}

	s.sendJSON(w, http.StatusOK, response)
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

	s.serveBalanceJSON(w, r)
}

func (s *Server) serveBalanceJSON(w http.ResponseWriter, r *http.Request) {
	balance, address, ok := s.lookupBalance(r, w)
	if !ok {
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

func (s *Server) lookupBalance(r *http.Request, w http.ResponseWriter) (config.BalanceResponse, string, bool) {
	address := r.URL.Query().Get("address")
	if address == "" {
		s.sendError(w, http.StatusBadRequest, "missing-parameter", "Missing 'address' parameter in query string")
		return config.BalanceResponse{}, "", false
	}

	pubkeyHash, err := doge.Base58DecodeCheck(address)
	if err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid-address", "Invalid Dogecoin address format")
		return config.BalanceResponse{}, "", false
	}

	if len(pubkeyHash) != 21 {
		s.sendError(w, http.StatusBadRequest, "invalid-address", "Invalid Dogecoin address length")
		return config.BalanceResponse{}, "", false
	}

	scriptType := scriptTypeFromVersionByte(pubkeyHash[0])
	if scriptType == doge.ScriptTypeNone {
		s.sendError(w, http.StatusBadRequest, "unsupported-address", "Unsupported address type")
		return config.BalanceResponse{}, "", false
	}

	balance, err := s.indexerStore.GetBalance(r.Context(), address)
	if err != nil {
		s.logBalanceRequest(r, address, false)
		log.Printf("[Dogelytics] indexer balance request failed: %v", err)
		s.sendError(w, http.StatusBadGateway, "indexer-error", "Failed to retrieve balance from the indexer")
		return config.BalanceResponse{}, "", false
	}

	return config.BalanceResponse{
		Incoming:  balance.Incoming,
		Available: balance.Available,
		Outgoing:  balance.Outgoing,
		Current:   balance.Current,
	}, address, true
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

	if err := s.authStore.LogRequest(r.Context(), getClientIP(r), apiKey, address, success); err != nil {
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
