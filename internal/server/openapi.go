package server

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openAPISpecification []byte

// HandleOpenAPI serves the release API contract.
func (s *Server) HandleOpenAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "method-not-allowed", "Only GET is allowed")
		return
	}
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	_, _ = w.Write(openAPISpecification)
}
