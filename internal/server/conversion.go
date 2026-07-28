package server

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dogeorg/dogelytics/internal/config"
	"github.com/dogeorg/dogelytics/internal/store"
)

var conversionCurrencyPattern = regexp.MustCompile(`^[a-z0-9-]{2,16}$`)

const conversionCacheMaxAge = time.Hour

// ConversionStore defines the cache operations needed for currency conversion rates.
type ConversionStore interface {
	GetFreshConversionRate(ctx context.Context, currency string, maxAge time.Duration) (store.ConversionRate, bool, error)
	UpsertConversionRate(ctx context.Context, currency string, rate string, fetchedAt time.Time, coingeckoUpdatedAt *time.Time) (store.ConversionRate, error)
}

// HandleConversion responds with a cached or refreshed DOGE conversion rate.
func (s *Server) HandleConversion(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w, r)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "method-not-allowed", "Only GET and OPTIONS methods are allowed")
		return
	}

	currency, errCode, message := normaliseConversionCurrency(r.URL.Query().Get("currency"))
	if errCode != "" {
		s.sendError(w, http.StatusBadRequest, errCode, message)
		return
	}
	if currency == "doge" {
		s.sendJSON(w, http.StatusOK, staticDogeConversionResponse())
		return
	}
	if s.conversionStore == nil {
		s.sendError(w, http.StatusServiceUnavailable, "conversion-cache-unavailable", "Conversion cache is unavailable")
		return
	}
	if s.conversionClient == nil {
		s.sendError(w, http.StatusBadGateway, "conversion-source-error", "Failed to fetch conversion rate")
		return
	}

	cachedRate, found, err := s.conversionStore.GetFreshConversionRate(r.Context(), currency, conversionCacheMaxAge)
	if err != nil {
		s.sendError(w, http.StatusServiceUnavailable, "conversion-cache-unavailable", "Conversion cache is unavailable")
		return
	}
	if found {
		s.sendJSON(w, http.StatusOK, conversionResponseFromCache(cachedRate))
		return
	}

	quote, err := s.conversionClient.GetRate(r.Context(), currency)
	if err != nil {
		if errors.Is(err, ErrUnsupportedCurrency) {
			s.sendError(w, http.StatusBadRequest, "unsupported-currency", "Unsupported currency code")
			return
		}
		s.sendError(w, http.StatusBadGateway, "conversion-source-error", "Failed to fetch conversion rate")
		return
	}

	refreshedRate, err := s.conversionStore.UpsertConversionRate(
		r.Context(),
		quote.Currency,
		quote.Rate,
		time.Now().UTC(),
		quote.CoinGeckoUpdatedAt,
	)
	if err != nil {
		s.sendError(w, http.StatusServiceUnavailable, "conversion-cache-unavailable", "Conversion cache is unavailable")
		return
	}

	s.sendJSON(w, http.StatusOK, conversionResponseFromQuote(refreshedRate))
}

func normaliseConversionCurrency(raw string) (string, string, string) {
	currency := strings.ToLower(strings.TrimSpace(raw))
	if currency == "" {
		return "", "missing-parameter", "Missing 'currency' parameter in query string"
	}
	if !conversionCurrencyPattern.MatchString(currency) {
		return "", "invalid-currency", "Invalid currency code"
	}
	return currency, "", ""
}

func staticDogeConversionResponse() config.ConversionResponse {
	return config.ConversionResponse{
		Currency:  "doge",
		Rate:      "1",
		Source:    "static",
		Cached:    false,
		FetchedAt: time.Now().UTC(),
	}
}

func conversionResponseFromCache(rate store.ConversionRate) config.ConversionResponse {
	return config.ConversionResponse{
		Currency:           rate.Currency,
		Rate:               rate.Rate,
		Source:             "cache",
		Cached:             true,
		FetchedAt:          rate.FetchedAt,
		CoinGeckoUpdatedAt: rate.CoinGeckoUpdatedAt,
	}
}

func conversionResponseFromQuote(rate store.ConversionRate) config.ConversionResponse {
	return config.ConversionResponse{
		Currency:           rate.Currency,
		Rate:               rate.Rate,
		Source:             "coingecko",
		Cached:             false,
		FetchedAt:          rate.FetchedAt,
		CoinGeckoUpdatedAt: rate.CoinGeckoUpdatedAt,
	}
}
