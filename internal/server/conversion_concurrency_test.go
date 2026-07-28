package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dogeorg/dogelytics/internal/config"
	"github.com/dogeorg/dogelytics/internal/store"
)

type concurrentConversionStore struct {
	upserts atomic.Int32
}

func (*concurrentConversionStore) GetConversionRate(context.Context, string) (store.ConversionRate, bool, error) {
	return store.ConversionRate{}, false, nil
}

func (*concurrentConversionStore) GetFreshConversionRate(context.Context, string, time.Duration) (store.ConversionRate, bool, error) {
	return store.ConversionRate{}, false, nil
}

func (s *concurrentConversionStore) UpsertConversionRate(_ context.Context, currency, rate string, fetchedAt time.Time, updatedAt *time.Time) (store.ConversionRate, error) {
	s.upserts.Add(1)
	return store.ConversionRate{Currency: currency, Rate: rate, FetchedAt: fetchedAt, CoinGeckoUpdatedAt: updatedAt}, nil
}

type concurrentConversionClient struct {
	calls atomic.Int32
}

func (c *concurrentConversionClient) GetRate(context.Context, string) (ConversionQuote, error) {
	c.calls.Add(1)
	time.Sleep(25 * time.Millisecond)
	return ConversionQuote{Currency: "usd", Rate: "0.2"}, nil
}

func TestHandleConversionDeduplicatesConcurrentRefreshes(t *testing.T) {
	cache := &concurrentConversionStore{}
	client := &concurrentConversionClient{}
	srv := newTestServer(&fakeBalanceStore{}, &config.Config{CorsOrigin: "*"})
	srv.conversionStore = cache
	srv.conversionClient = client

	const requests = 16
	start := make(chan struct{})
	results := make(chan int, requests)
	var workers sync.WaitGroup
	for range requests {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			req := httptest.NewRequest(http.MethodGet, "/conversion?currency=usd", nil)
			rec := httptest.NewRecorder()
			srv.APIHandler().ServeHTTP(rec, req)
			results <- rec.Code
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	for status := range results {
		if status != http.StatusOK {
			t.Fatalf("unexpected response status: %d", status)
		}
	}
	if got := client.calls.Load(); got != 1 {
		t.Fatalf("expected one upstream refresh, got %d", got)
	}
	if got := cache.upserts.Load(); got != 1 {
		t.Fatalf("expected one cache update, got %d", got)
	}
}

func BenchmarkRenderCachedTemplate(b *testing.B) {
	const source = `<html><body><h1>{{.}}</h1></body></html>`
	for b.Loop() {
		renderNamedTemplate(httptest.NewRecorder(), "benchmark", source, "Dogelytics")
	}
}
