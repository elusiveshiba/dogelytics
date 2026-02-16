package server

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"

	"github.com/dogeorg/dogelytics/internal/config"
)

func TestSetSessionCookie_HTTPNotSecure(t *testing.T) {
	s := &Server{
		config: &config.Config{
			SessionSecret: "test-secret",
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost:4420/login", nil)

	s.setSessionCookie(w, r, "user-123")

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != sessionCookieName {
		t.Fatalf("unexpected cookie name: %s", cookies[0].Name)
	}
	if cookies[0].Secure {
		t.Fatalf("cookie should not be Secure for HTTP request")
	}
}

func TestSetSessionCookie_HTTPSecure(t *testing.T) {
	s := &Server{
		config: &config.Config{
			SessionSecret: "test-secret",
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "https://localhost/login", nil)
	r.TLS = &tls.ConnectionState{}

	s.setSessionCookie(w, r, "user-123")

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if !cookies[0].Secure {
		t.Fatalf("cookie should be Secure for HTTPS request")
	}
}

