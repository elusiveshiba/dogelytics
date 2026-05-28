package server

import "net/http"

// Handler returns the API handler for backwards compatibility.
func (s *Server) Handler() http.Handler {
	return s.APIHandler()
}

// APIHandler returns the public API handler.
func (s *Server) APIHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.HandleHealth)
	mux.HandleFunc("/balance", s.APIKeyAuthMiddleware(s.RateLimitMiddleware(s.HandleBalance)))
	mux.HandleFunc("/conversion", s.APIKeyAuthMiddleware(s.RateLimitMiddleware(s.HandleConversion)))

	return mux
}

// AdminHandler returns the admin UI handler.
func (s *Server) AdminHandler() http.Handler {
	mux := http.NewServeMux()
	registerFaviconRoutes(mux)
	mux.HandleFunc("/", s.HandleGETIndex)
	mux.HandleFunc("/login", methodHandler(map[string]http.HandlerFunc{
		http.MethodGet:  s.HandleGETLogin,
		http.MethodPost: s.HandlePOSTLogin,
	}))
	mux.HandleFunc("/register", methodHandler(map[string]http.HandlerFunc{
		http.MethodGet:  s.HandleGETRegister,
		http.MethodPost: s.HandlePOSTRegister,
	}))
	mux.HandleFunc("/logout", methodHandler(map[string]http.HandlerFunc{
		http.MethodPost: s.HandlePOSTLogout,
	}))
	mux.HandleFunc("/keys", s.RequireLogin(methodHandler(map[string]http.HandlerFunc{
		http.MethodGet:  s.HandleGETKeys,
		http.MethodPost: s.HandlePOSTCreateKey,
	})))
	mux.HandleFunc("/keys/revoke", s.RequireLogin(methodHandler(map[string]http.HandlerFunc{
		http.MethodPost: s.HandlePOSTRevokeKey,
	})))
	mux.HandleFunc("/keys/expire", s.RequireLogin(methodHandler(map[string]http.HandlerFunc{
		http.MethodPost: s.HandlePOSTExpireKey,
	})))
	mux.HandleFunc("/api/stats/usage", s.RequireLogin(s.HandleUsageStats))
	mux.HandleFunc("/api/stats/timeseries", s.RequireLogin(s.HandleUsageTimeSeries))
	return mux
}

// DashboardHandler returns the public dashboard UI handler.
func (s *Server) DashboardHandler() http.Handler {
	mux := http.NewServeMux()
	registerFaviconRoutes(mux)
	mux.HandleFunc("/docs", s.HandleGETDocs)
	mux.HandleFunc("/", s.HandleGETDashboard)
	mux.HandleFunc("/api/balance", s.APIKeyAuthMiddleware(s.RateLimitMiddleware(s.HandleDashboardBalance)))
	mux.HandleFunc("/api/conversion", s.APIKeyAuthMiddleware(s.RateLimitMiddleware(s.HandleConversion)))
	mux.HandleFunc("/api/dashboard-stats", s.APIKeyAuthMiddleware(s.RateLimitMiddleware(s.HandleDashboardStats)))
	return mux
}

func methodHandler(handlers map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler, ok := handlers[r.Method]
		if !ok {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}
