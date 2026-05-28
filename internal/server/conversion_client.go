package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const coinGeckoBaseURL = "https://api.coingecko.com/api/v3"

var ErrUnsupportedCurrency = errors.New("unsupported currency")

// ConversionQuote is a DOGE conversion quote from the upstream price source.
type ConversionQuote struct {
	Currency           string
	Rate               string
	CoinGeckoUpdatedAt *time.Time
}

// ConversionClient fetches DOGE conversion rates from an upstream source.
type ConversionClient interface {
	GetRate(ctx context.Context, currency string) (ConversionQuote, error)
}

// CoinGeckoClient fetches DOGE conversion rates from CoinGecko.
type CoinGeckoClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewCoinGeckoClient creates a CoinGecko client with a sensible default timeout.
func NewCoinGeckoClient(httpClient *http.Client) *CoinGeckoClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &CoinGeckoClient{
		baseURL:    coinGeckoBaseURL,
		httpClient: httpClient,
	}
}

// GetRate fetches the current DOGE conversion rate for a single target currency.
func (c *CoinGeckoClient) GetRate(ctx context.Context, currency string) (ConversionQuote, error) {
	normalisedCurrency := strings.ToLower(strings.TrimSpace(currency))

	endpoint, err := url.Parse(c.baseURL + "/simple/price")
	if err != nil {
		return ConversionQuote{}, fmt.Errorf("build coingecko url: %w", err)
	}

	query := endpoint.Query()
	query.Set("ids", "dogecoin")
	query.Set("vs_currencies", normalisedCurrency)
	query.Set("include_last_updated_at", "true")
	query.Set("precision", "full")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return ConversionQuote{}, fmt.Errorf("create coingecko request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ConversionQuote{}, fmt.Errorf("request coingecko rate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ConversionQuote{}, fmt.Errorf("coingecko returned status %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	var payload map[string]map[string]json.Number
	if err := decoder.Decode(&payload); err != nil {
		return ConversionQuote{}, fmt.Errorf("decode coingecko response: %w", err)
	}

	entry, ok := payload["dogecoin"]
	if !ok {
		return ConversionQuote{}, ErrUnsupportedCurrency
	}

	rate, ok := entry[normalisedCurrency]
	if !ok {
		return ConversionQuote{}, ErrUnsupportedCurrency
	}

	quote := ConversionQuote{
		Currency: normalisedCurrency,
		Rate:     rate.String(),
	}

	if lastUpdated, ok := entry["last_updated_at"]; ok {
		unixTime, err := lastUpdated.Int64()
		if err != nil {
			return ConversionQuote{}, fmt.Errorf("parse coingecko last_updated_at: %w", err)
		}
		updatedAt := time.Unix(unixTime, 0).UTC()
		quote.CoinGeckoUpdatedAt = &updatedAt
	}

	return quote, nil
}
