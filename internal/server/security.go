package server

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

const maxFormBytes = 64 << 10

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) csrfProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !unsafeMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}
		if strings.EqualFold(r.Header.Get("Sec-Fetch-Site"), "cross-site") {
			http.Error(w, "cross-site request rejected", http.StatusForbidden)
			return
		}

		source := r.Header.Get("Origin")
		if source == "" {
			source = r.Header.Get("Referer")
		}
		if source != "" && !s.sameOrigin(r, source) {
			http.Error(w, "cross-origin request rejected", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func unsafeMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func (s *Server) sameOrigin(r *http.Request, source string) bool {
	parsed, err := url.Parse(source)
	if err != nil || parsed.Host == "" {
		return false
	}
	expectedScheme := "http"
	expectedHost := r.Host
	if s.config.PublicURL != "" {
		publicURL, err := url.Parse(s.config.PublicURL)
		if err != nil {
			return false
		}
		expectedScheme = publicURL.Scheme
		expectedHost = publicURL.Host
	} else if s.requestIsHTTPS(r) {
		expectedScheme = "https"
	}
	return strings.EqualFold(parsed.Scheme, expectedScheme) && strings.EqualFold(parsed.Host, expectedHost)
}

func (s *Server) requestIsHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	peer := getClientIP(r)
	return trustedProxy(peer, s.config.TrustedProxies) && strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func parseLimitedForm(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBytes)
	if err := r.ParseForm(); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return errors.New("form is too large")
		}
		return errors.New("invalid form")
	}
	return nil
}

// AuthRateLimitMiddleware limits password hashing attempts by client IP.
func (s *Server) AuthRateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := "auth:" + s.getClientIP(r)
		if s.authLimiter != nil && !s.authLimiter.Allow(clientID) {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "too many authentication attempts", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}
