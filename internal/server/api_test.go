package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dogeorg/dogelytics/internal/config"
	"github.com/dogeorg/dogelytics/internal/indexer"
	"github.com/dogeorg/dogelytics/internal/store"
)

type fakeBalanceStore struct {
	height      int64
	heightErr   error
	balance     indexer.Balance
	balanceErr  error
	lastAddress string
}

type fakeSyncSource struct {
	syncHeights indexer.SyncHeights
	err         error
}

func (f *fakeBalanceStore) GetBalance(_ context.Context, address string) (indexer.Balance, error) {
	f.lastAddress = address
	return f.balance, f.balanceErr
}

func (f *fakeBalanceStore) CurrentHeight(context.Context) (int64, error) {
	return f.height, f.heightErr
}

func (f fakeSyncSource) SyncHeights(context.Context) (indexer.SyncHeights, error) {
	if f.err != nil {
		return indexer.SyncHeights{}, f.err
	}
	return f.syncHeights, nil
}

type fakeConversionStore struct {
	getRate                  store.ConversionRate
	getFound                 bool
	getErr                   error
	staleRate                store.ConversionRate
	staleFound               bool
	staleErr                 error
	upsertRate               store.ConversionRate
	upsertErr                error
	getCalls                 int
	upsertCalls              int
	lastCurrency             string
	lastMaxAge               time.Duration
	lastUpsertCurrency       string
	lastUpsertRate           string
	lastUpsertFetchedAt      time.Time
	lastUpsertCoinGeckoStamp *time.Time
}

func (f *fakeConversionStore) GetConversionRate(_ context.Context, _ string) (store.ConversionRate, bool, error) {
	return f.staleRate, f.staleFound, f.staleErr
}

func (f *fakeConversionStore) GetFreshConversionRate(_ context.Context, currency string, maxAge time.Duration) (store.ConversionRate, bool, error) {
	f.getCalls++
	f.lastCurrency = currency
	f.lastMaxAge = maxAge
	return f.getRate, f.getFound, f.getErr
}

func (f *fakeConversionStore) UpsertConversionRate(_ context.Context, currency string, rate string, fetchedAt time.Time, coingeckoUpdatedAt *time.Time) (store.ConversionRate, error) {
	f.upsertCalls++
	f.lastUpsertCurrency = currency
	f.lastUpsertRate = rate
	f.lastUpsertFetchedAt = fetchedAt
	f.lastUpsertCoinGeckoStamp = coingeckoUpdatedAt
	if f.upsertErr != nil {
		return store.ConversionRate{}, f.upsertErr
	}
	if f.upsertRate.Currency == "" {
		f.upsertRate = store.ConversionRate{
			Currency:           currency,
			Rate:               rate,
			FetchedAt:          fetchedAt,
			CoinGeckoUpdatedAt: coingeckoUpdatedAt,
		}
	}
	return f.upsertRate, nil
}

type fakeConversionClient struct {
	quote        ConversionQuote
	err          error
	calls        int
	lastCurrency string
}

func (f *fakeConversionClient) GetRate(_ context.Context, currency string) (ConversionQuote, error) {
	f.calls++
	f.lastCurrency = currency
	if f.err != nil {
		return ConversionQuote{}, f.err
	}
	return f.quote, nil
}

func TestHandlerRoutesAPIOnly(t *testing.T) {
	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.APIHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for removed UI routes, got %d", rec.Code)
	}
}

func TestHandleHealthSuccess(t *testing.T) {
	store := &fakeBalanceStore{height: 123}
	srv := newTestServer(store, &config.Config{CorsOrigin: "*"})
	srv.syncSource = fakeSyncSource{
		syncHeights: indexer.SyncHeights{
			IndexerHeight:     123,
			CoreBlocksHeight:  int64Ptr(900),
			CoreHeadersHeight: int64Ptr(1000),
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.Bytes()
	var resp HealthResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if !resp.OK || resp.IndexerHeight != 123 {
		t.Fatalf("unexpected health response: %+v", resp)
	}
	if resp.CoreBlocksHeight != 900 || resp.CoreHeadersHeight != 1000 {
		t.Fatalf("unexpected health heights: %+v", resp)
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("decode raw health response: %v", err)
	}
	for _, field := range []string{"height", "indexed_percent", "core_synced_percent", "core_height", "blockchain_height", "blocks_height", "headers_height"} {
		if _, ok := raw[field]; ok {
			t.Fatalf("did not expect %q in health response: %+v", field, raw)
		}
	}
}

func TestHandleHealthStoreError(t *testing.T) {
	store := &fakeBalanceStore{heightErr: errors.New("boom")}
	srv := newTestServer(store, &config.Config{CorsOrigin: "*"})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
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
			Incoming:  "100",
			Available: "200",
			Outgoing:  "50",
			Current:   "300",
		},
	}
	srv := newTestServer(store, &config.Config{CorsOrigin: "*"})

	req := httptest.NewRequest(http.MethodGet, "/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if store.lastAddress != "DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8" {
		t.Fatalf("expected balance lookup address to be preserved, got %q", store.lastAddress)
	}

	var resp config.BalanceResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode balance response: %v", err)
	}
	if resp.Current != "300" {
		t.Fatalf("unexpected current balance: %v", resp.Current)
	}
}

func TestHandleConversionUsesFreshCache(t *testing.T) {
	cachedAt := time.Now().UTC()
	updatedAt := cachedAt.Add(-2 * time.Minute)
	cache := &fakeConversionStore{
		getFound: true,
		getRate: store.ConversionRate{
			Currency:           "usd",
			Rate:               "0.25",
			FetchedAt:          cachedAt,
			CoinGeckoUpdatedAt: &updatedAt,
		},
	}
	client := &fakeConversionClient{}

	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})
	srv.conversionStore = cache
	srv.conversionClient = client

	req := httptest.NewRequest(http.MethodGet, "/conversion?currency=USD", nil)
	rec := httptest.NewRecorder()
	srv.APIHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if client.calls != 0 {
		t.Fatalf("expected no upstream calls, got %d", client.calls)
	}
	if cache.lastCurrency != "usd" {
		t.Fatalf("expected lowercase currency lookup, got %q", cache.lastCurrency)
	}
	if cache.lastMaxAge != time.Hour {
		t.Fatalf("expected 1 hour cache max age, got %v", cache.lastMaxAge)
	}

	var resp config.ConversionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode conversion response: %v", err)
	}
	if !resp.Cached || resp.Source != "cache" || resp.Rate != "0.25" {
		t.Fatalf("unexpected cached conversion response: %+v", resp)
	}
}

func TestHandleConversionRefreshesCache(t *testing.T) {
	updatedAt := time.Now().UTC().Add(-30 * time.Second)
	cache := &fakeConversionStore{}
	client := &fakeConversionClient{
		quote: ConversionQuote{
			Currency:           "usd",
			Rate:               "0.2345",
			CoinGeckoUpdatedAt: &updatedAt,
		},
	}

	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})
	srv.conversionStore = cache
	srv.conversionClient = client

	req := httptest.NewRequest(http.MethodGet, "/conversion?currency=usd", nil)
	rec := httptest.NewRecorder()
	srv.APIHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if client.calls != 1 {
		t.Fatalf("expected one upstream call, got %d", client.calls)
	}
	if cache.upsertCalls != 1 {
		t.Fatalf("expected one cache upsert, got %d", cache.upsertCalls)
	}
	if cache.lastUpsertCurrency != "usd" || cache.lastUpsertRate != "0.2345" {
		t.Fatalf("unexpected upsert inputs: currency=%q rate=%q", cache.lastUpsertCurrency, cache.lastUpsertRate)
	}

	var resp config.ConversionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode conversion response: %v", err)
	}
	if resp.Cached || resp.Source != "coingecko" || resp.Rate != "0.2345" {
		t.Fatalf("unexpected refreshed conversion response: %+v", resp)
	}
}

func TestHandleConversionInvalidCurrency(t *testing.T) {
	cache := &fakeConversionStore{}
	client := &fakeConversionClient{}

	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})
	srv.conversionStore = cache
	srv.conversionClient = client

	req := httptest.NewRequest(http.MethodGet, "/conversion?currency=us$", nil)
	rec := httptest.NewRecorder()
	srv.APIHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if cache.getCalls != 0 || client.calls != 0 {
		t.Fatalf("expected no cache or upstream calls for invalid currency")
	}

	var resp config.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error != "invalid-currency" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestHandleConversionUnsupportedCurrency(t *testing.T) {
	cache := &fakeConversionStore{}
	client := &fakeConversionClient{err: ErrUnsupportedCurrency}

	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})
	srv.conversionStore = cache
	srv.conversionClient = client

	req := httptest.NewRequest(http.MethodGet, "/conversion?currency=zzz", nil)
	rec := httptest.NewRecorder()
	srv.APIHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var resp config.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error != "unsupported-currency" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestHandleConversionUpstreamFailure(t *testing.T) {
	cache := &fakeConversionStore{}
	client := &fakeConversionClient{err: errors.New("boom")}

	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})
	srv.conversionStore = cache
	srv.conversionClient = client

	req := httptest.NewRequest(http.MethodGet, "/conversion?currency=usd", nil)
	rec := httptest.NewRecorder()
	srv.APIHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}

	var resp config.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error != "conversion-source-error" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestHandleConversionFallsBackToStaleCache(t *testing.T) {
	cachedAt := time.Now().UTC().Add(-2 * time.Hour)
	cache := &fakeConversionStore{
		staleFound: true,
		staleRate: store.ConversionRate{
			Currency:  "usd",
			Rate:      "0.21",
			FetchedAt: cachedAt,
		},
	}
	client := &fakeConversionClient{err: errors.New("boom")}
	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})
	srv.conversionStore = cache
	srv.conversionClient = client

	req := httptest.NewRequest(http.MethodGet, "/conversion?currency=usd", nil)
	rec := httptest.NewRecorder()
	srv.APIHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected stale-cache response, got %d", rec.Code)
	}
	var resp config.ConversionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode stale response: %v", err)
	}
	if resp.Source != "stale-cache" || !resp.Cached || resp.Rate != "0.21" {
		t.Fatalf("unexpected stale response: %+v", resp)
	}
}

func TestHandleConversionWithoutCache(t *testing.T) {
	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})
	srv.conversionStore = nil
	srv.conversionClient = &fakeConversionClient{}

	req := httptest.NewRequest(http.MethodGet, "/conversion?currency=usd", nil)
	rec := httptest.NewRecorder()
	srv.APIHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var resp config.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error != "conversion-cache-unavailable" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestHandleConversionSupportsDogeCurrency(t *testing.T) {
	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})
	srv.conversionStore = nil
	srv.conversionClient = nil

	req := httptest.NewRequest(http.MethodGet, "/conversion?currency=DOGE", nil)
	rec := httptest.NewRecorder()
	srv.APIHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp config.ConversionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode conversion response: %v", err)
	}
	if resp.Currency != "doge" || resp.Rate != "1" || resp.Source != "static" || resp.Cached {
		t.Fatalf("unexpected DOGE conversion response: %+v", resp)
	}
	if resp.FetchedAt.IsZero() {
		t.Fatalf("expected DOGE conversion response timestamp: %+v", resp)
	}
}

func TestDashboardConversionRouteUsesConversionHandler(t *testing.T) {
	fetchedAt := time.Date(2026, 5, 28, 1, 0, 0, 0, time.UTC)
	cache := &fakeConversionStore{
		getRate: store.ConversionRate{
			Currency:  "usd",
			Rate:      "0.10",
			FetchedAt: fetchedAt,
		},
		getFound: true,
	}
	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})
	srv.conversionStore = cache
	srv.conversionClient = &fakeConversionClient{}

	req := httptest.NewRequest(http.MethodGet, "/api/conversion?currency=usd", nil)
	rec := httptest.NewRecorder()
	srv.DashboardHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp config.ConversionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode conversion response: %v", err)
	}
	if resp.Currency != "usd" || resp.Rate != "0.10" {
		t.Fatalf("unexpected dashboard conversion response: %+v", resp)
	}
}

func newTestServer(store BalanceStore, cfg *config.Config) *Server {
	return NewServer(store, nil, nil, cfg, nil, nil)
}

func int64Ptr(v int64) *int64 {
	return &v
}
