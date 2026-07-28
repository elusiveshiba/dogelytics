package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dogeorg/dogelytics/internal/config"
)

func TestCSRFProtectionRejectsCrossOriginPosts(t *testing.T) {
	srv := &Server{config: &config.Config{PublicURL: "https://admin.example.com"}}
	handler := srv.csrfProtection(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodPost, "http://internal/keys", strings.NewReader("kid=abc"))
	request.Header.Set("Origin", "https://attacker.example")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected cross-origin POST to fail, got %d", recorder.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "http://internal/keys", strings.NewReader("kid=abc"))
	request.Header.Set("Origin", "https://admin.example.com")
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected same-origin POST to pass, got %d", recorder.Code)
	}
}

func TestTrustedProxyControlsForwardedClientIP(t *testing.T) {
	srv := &Server{config: &config.Config{TrustedProxies: "10.0.0.0/8"}}
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "10.1.2.3:1234"
	request.Header.Set("X-Forwarded-For", "203.0.113.10, 10.1.2.3")
	if got := srv.getClientIP(request); got != "203.0.113.10" {
		t.Fatalf("forwarded client mismatch: got %q", got)
	}

	request.RemoteAddr = "192.0.2.5:1234"
	if got := srv.getClientIP(request); got != "192.0.2.5" {
		t.Fatalf("untrusted proxy should be ignored, got %q", got)
	}
}

func TestCORSAllowsConfiguredOriginsAndAuthenticationHeaders(t *testing.T) {
	srv := &Server{config: &config.Config{CorsOrigin: "https://one.example, https://two.example"}}
	request := httptest.NewRequest(http.MethodOptions, "/balance", nil)
	request.Header.Set("Origin", "https://two.example")
	recorder := httptest.NewRecorder()
	srv.setCORSHeaders(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://two.example" {
		t.Fatalf("allowed origin mismatch: %q", got)
	}
	headers := recorder.Header().Get("Access-Control-Allow-Headers")
	for _, required := range []string{"Authorization", "X-Api-Key"} {
		if !strings.Contains(headers, required) {
			t.Fatalf("missing CORS header %q in %q", required, headers)
		}
	}
}

func TestParseLimitedFormRejectsOversizedBody(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("email="+strings.Repeat("a", maxFormBytes)))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	if err := parseLimitedForm(recorder, request); err == nil {
		t.Fatal("expected oversized form to fail")
	}
}
