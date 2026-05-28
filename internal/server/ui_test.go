package server

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dogeorg/doge/koinu"
	"github.com/dogeorg/dogelytics/internal/config"
	"github.com/dogeorg/dogelytics/internal/indexer"
)

type fakeCoreClient struct {
	height int64
	err    error
}

func (f fakeCoreClient) BlockchainHeight(context.Context) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.height, nil
}

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
		CorsOrigin:           "*",
		EnableDashboardUI:    true,
		DashboardUIPort:      4422,
		EnableAdminUI:        true,
		AdminUIPort:          4421,
		EnableDashboardTips:  true,
		DashboardTipsAddress: "DChPB3HbQgNYgWRrpeRKqNT6939rRLceNz",
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
	for _, want := range []string{
		"Wallets",
		"Blockchain",
		"Total wallets checked (24h)",
		"Unique wallets checked (24h)",
		"Total wallets checked",
		"Unique wallets checked",
		"Indexed Height",
		"Blockchain Height",
		`id="stat-unique-wallets"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected dashboard stats layout to contain %q in %q", want, body)
		}
	}
	if !strings.Contains(body, `class="docs-button" href="/docs">docs</a>`) {
		t.Fatalf("expected docs button in %q", body)
	}
	if !strings.Contains(body, "Such coffee?") {
		t.Fatalf("expected tips widget in %q", body)
	}
	if !strings.Contains(body, "Enjoying Dogelytics? Send a tip to:") {
		t.Fatalf("expected tips prompt in %q", body)
	}
	if !strings.Contains(body, "DChPB3HbQgNYgWRrpeRKqNT6939rRLceNz") {
		t.Fatalf("expected tips address in %q", body)
	}
	if strings.Contains(body, "Copy address") {
		t.Fatalf("did not expect copy button text in %q", body)
	}
	if !strings.Contains(body, `src="https://fetch.dogecoin.org/doge-qr.js"`) {
		t.Fatalf("expected doge-qr script in %q", body)
	}
	if !strings.Contains(body, `<doge-qr address="DChPB3HbQgNYgWRrpeRKqNT6939rRLceNz" size="sm" background="#fff" fill="#000"></doge-qr>`) {
		t.Fatalf("expected doge-qr element in %q", body)
	}
	if !strings.Contains(body, "applyMinimised(true)") {
		t.Fatalf("expected tips widget to start minimised in %q", body)
	}
	if !strings.Contains(body, "formatIndexedProgress(payload.height, payload.blockchain_height)") {
		t.Fatalf("expected indexed progress formatting in %q", body)
	}
	if !strings.Contains(body, "formatIndexedProgress(payload.blockchain_height, payload.blockchain_height)") {
		t.Fatalf("expected blockchain progress formatting in %q", body)
	}
	if !strings.Contains(body, "formatInteger(payload.unique_wallets_checked)") {
		t.Fatalf("expected overall unique wallet formatting in %q", body)
	}
	if !strings.Contains(body, "'Ð' + formatDogeAmount(payload.current)") {
		t.Fatalf("expected formatted Dogecoin balance in %q", body)
	}
	if !strings.Contains(body, "font-size: 14px;") {
		t.Fatalf("expected larger balance font size in %q", body)
	}
}

func TestDashboardDocsRoute(t *testing.T) {
	srv := newTestServer(&fakeBalanceStore{}, &config.Config{
		CorsOrigin:        "*",
		EnableDashboardUI: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	srv.DashboardHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	for _, want := range []string{
		"Dogelytics API Docs",
		"https://api.dogelytics.com",
		"GET /balance",
		"GET /conversion",
		"currency=usd",
		"cached row is older than one hour",
		`class="dashboard-button" href="/">dashboard</a>`,
		"Authorization: Bearer YOUR_API_KEY",
		"GET /health",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected docs page to contain %q in %q", want, body)
		}
	}
	if strings.Contains(body, "Back to dashboard") {
		t.Fatalf("did not expect old back link in %q", body)
	}
}

func TestDashboardHandlerServesFavicons(t *testing.T) {
	srv := newTestServer(&fakeBalanceStore{}, &config.Config{
		CorsOrigin:        "*",
		EnableDashboardUI: true,
	})

	tests := []string{
		"/favicon.ico",
		"/apple-touch-icon.png",
		"/site.webmanifest",
		"/favicons/favicon-32x32.png",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			srv.DashboardHandler().ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}
		})
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
	srv.coreClient = fakeCoreClient{height: 1000}

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
	if resp.BlockchainHeight != 1000 {
		t.Fatalf("expected blockchain height 1000, got %+v", resp)
	}
	if resp.UniqueWalletsChecked != 0 {
		t.Fatalf("expected zero overall unique wallets without auth store, got %+v", resp)
	}
	if resp.IndexedPercent == nil || math.Abs(*resp.IndexedPercent-77.7) > 0.0001 {
		t.Fatalf("expected indexed percent 77.7, got %+v", resp)
	}
}

func TestDashboardStatsIgnoresCoreHeightErrors(t *testing.T) {
	srv := newTestServer(&fakeBalanceStore{height: 777}, &config.Config{
		CorsOrigin:        "*",
		EnableDashboardUI: true,
	})
	srv.coreClient = fakeCoreClient{err: errors.New("core unavailable")}

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
	if resp.BlockchainHeight != 0 || resp.IndexedPercent != nil {
		t.Fatalf("expected absent blockchain progress, got %+v", resp)
	}
}
