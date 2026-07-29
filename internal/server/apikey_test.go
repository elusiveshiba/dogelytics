package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestParseBearerOrHeader(t *testing.T) {
	t.Run("authorization bearer", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/balance", nil)
		r.Header.Set("Authorization", "Bearer dglk_abc.def")

		token, ok := parseBearerOrHeader(r)
		if !ok {
			t.Fatalf("expected token to be found")
		}
		if token != "dglk_abc.def" {
			t.Fatalf("unexpected token: %q", token)
		}
	})

	t.Run("x-api-key fallback", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/balance", nil)
		r.Header.Set("X-Api-Key", "dglk_xyz.123")

		token, ok := parseBearerOrHeader(r)
		if !ok {
			t.Fatalf("expected token to be found")
		}
		if token != "dglk_xyz.123" {
			t.Fatalf("unexpected token: %q", token)
		}
	})
}

func TestAPIKeySecretHashes(t *testing.T) {
	secret := "a-random-high-entropy-secret"
	hash := hashAPIKeySecret(secret)
	if valid, legacy := verifyAPIKeySecret(hash, secret); !valid || legacy {
		t.Fatalf("expected SHA-256 hash to verify without legacy marker")
	}
	if valid, _ := verifyAPIKeySecret(hash, "wrong"); valid {
		t.Fatal("expected wrong secret to fail")
	}

	legacyHash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("create legacy hash: %v", err)
	}
	if valid, legacy := verifyAPIKeySecret(string(legacyHash), secret); !valid || !legacy {
		t.Fatalf("expected bcrypt hash to verify as legacy")
	}
}

func TestParseToken(t *testing.T) {
	kid, secret, ok := parseToken("dglk_abc123.secret456")
	if !ok {
		t.Fatalf("expected token to parse")
	}
	if kid != "abc123" {
		t.Fatalf("unexpected kid: %q", kid)
	}
	if secret != "secret456" {
		t.Fatalf("unexpected secret: %q", secret)
	}

	_, _, ok = parseToken("not-a-token")
	if ok {
		t.Fatalf("expected malformed token to fail")
	}
}

func TestAPIKeyAuthMiddleware_NoToken_AllowsRequest(t *testing.T) {
	s := &Server{}
	nextCalled := false

	handler := s.APIKeyAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if !nextCalled {
		t.Fatalf("expected next handler to be called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAPIKeyAuthMiddleware_MalformedToken_ReturnsUnauthorized(t *testing.T) {
	s := &Server{}
	nextCalled := false

	handler := s.APIKeyAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	req.Header.Set("Authorization", "Bearer definitely-not-valid")
	rr := httptest.NewRecorder()
	handler(rr, req)

	if nextCalled {
		t.Fatalf("did not expect next handler to be called")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
