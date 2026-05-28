package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dogeorg/doge/koinu"
	"github.com/dogeorg/dogelytics/internal/config"
	"github.com/dogeorg/dogelytics/internal/indexer"
)

func TestAdminHandlerRoutes(t *testing.T) {
	srv := newTestServer(&fakeBalanceStore{}, &config.Config{
		CorsOrigin:    "*",
		EnableAdminUI: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rec := httptest.NewRecorder()
	srv.AdminHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Registration Disabled") {
		t.Fatalf("expected registration disabled page, got %q", rec.Body.String())
	}
}

func TestDashboardHandlerRoutes(t *testing.T) {
	srv := newTestServer(&fakeBalanceStore{height: 321}, &config.Config{
		CorsOrigin:        "*",
		EnableDashboardUI: true,
		DashboardUIPort:   4422,
		EnableAdminUI:     true,
		AdminUIPort:       4421,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "dogelytics.local:4422"
	rec := httptest.NewRecorder()
	srv.DashboardHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Dogelytics Dashboard") {
		t.Fatalf("expected dashboard title in %q", body)
	}
	if !strings.Contains(body, "Open admin UI") {
		t.Fatalf("expected admin UI link in %q", body)
	}
}

func TestDashboardBalanceRateLimit(t *testing.T) {
	srv := NewServer(
		&fakeBalanceStore{
			balance: indexer.Balance{
				Incoming:  koinu.Koinu(1),
				Available: koinu.Koinu(2),
				Outgoing:  koinu.Koinu(3),
				Current:   koinu.Koinu(4),
			},
		},
		nil,
		&config.Config{
			CorsOrigin:        "*",
			Confirmations:     6,
			RateLimit:         1,
			APIKeyRateLimit:   5,
			EnableDashboardUI: true,
		},
		NewRateLimiter(1),
		NewRateLimiter(5),
	)

	req1 := httptest.NewRequest(http.MethodGet, "/api/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8", nil)
	req1.RemoteAddr = "192.168.1.10:1234"
	rec1 := httptest.NewRecorder()
	srv.DashboardHandler().ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request should succeed, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8", nil)
	req2.RemoteAddr = "192.168.1.10:1234"
	rec2 := httptest.NewRecorder()
	srv.DashboardHandler().ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request should be rate limited, got %d", rec2.Code)
	}
}

func TestDashboardStatsWithoutAuthStore(t *testing.T) {
	srv := newTestServer(&fakeBalanceStore{height: 777}, &config.Config{
		CorsOrigin:        "*",
		EnableDashboardUI: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard-stats", nil)
	rec := httptest.NewRecorder()
	srv.DashboardHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp DashboardStatsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode dashboard stats response: %v", err)
	}
	if resp.Available {
		t.Fatalf("expected unavailable stats response, got %+v", resp)
	}
	if resp.Height != 777 {
		t.Fatalf("expected height 777, got %+v", resp)
	}
}
