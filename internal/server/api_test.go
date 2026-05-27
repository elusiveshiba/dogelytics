package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dogeorg/doge"
	"github.com/dogeorg/doge/koinu"
	"github.com/dogeorg/dogelytics/internal/config"
	"github.com/dogeorg/dogelytics/internal/indexer"
)

type fakeBalanceStore struct {
	height      int64
	heightErr   error
	balance     indexer.Balance
	balanceErr  error
	lastScript  doge.ScriptType
	lastAddress []byte
}

func (f *fakeBalanceStore) GetBalance(_ context.Context, scriptType doge.ScriptType, address []byte, _ int64) (indexer.Balance, error) {
	f.lastScript = scriptType
	f.lastAddress = append([]byte(nil), address...)
	return f.balance, f.balanceErr
}

func (f *fakeBalanceStore) CurrentHeight(context.Context) (int64, error) {
	return f.height, f.heightErr
}

func TestHandlerRoutesAPIOnly(t *testing.T) {
	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*", EnableUI: false})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for removed UI routes, got %d", rec.Code)
	}
}

func TestHandleHealthSuccess(t *testing.T) {
	store := &fakeBalanceStore{height: 123}
	srv := newTestServer(store, &config.Config{CorsOrigin: "*"})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if !resp.OK || resp.Height != 123 {
		t.Fatalf("unexpected health response: %+v", resp)
	}
}

func TestHandleHealthStoreError(t *testing.T) {
	store := &fakeBalanceStore{heightErr: errors.New("boom")}
	srv := newTestServer(store, &config.Config{CorsOrigin: "*"})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestHandleBalanceMissingAddress(t *testing.T) {
	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})

	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var resp config.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error != "missing-parameter" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestHandleBalanceInvalidAddress(t *testing.T) {
	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})

	req := httptest.NewRequest(http.MethodGet, "/balance?address=not-a-doge-address", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleBalanceSuccess(t *testing.T) {
	store := &fakeBalanceStore{
		balance: indexer.Balance{
			Incoming:  koinu.Koinu(100),
			Available: koinu.Koinu(200),
			Outgoing:  koinu.Koinu(50),
			Current:   koinu.Koinu(300),
		},
	}
	srv := newTestServer(store, &config.Config{CorsOrigin: "*", Confirmations: 6})

	req := httptest.NewRequest(http.MethodGet, "/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if store.lastScript != doge.ScriptTypeP2PKH {
		t.Fatalf("expected P2PKH script type, got %v", store.lastScript)
	}
	if len(store.lastAddress) != 20 {
		t.Fatalf("expected 20-byte address hash, got %d bytes", len(store.lastAddress))
	}

	var resp config.BalanceResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode balance response: %v", err)
	}
	if resp.Current != koinu.Koinu(300) {
		t.Fatalf("unexpected current balance: %v", resp.Current)
	}
}

func newTestServer(store BalanceStore, cfg *config.Config) *Server {
	return NewServer(store, nil, cfg, nil, nil)
}
