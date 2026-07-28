package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ConversionRate stores a cached DOGE conversion rate for a target currency.
type ConversionRate struct {
	Currency           string
	Rate               string
	FetchedAt          time.Time
	CoinGeckoUpdatedAt *time.Time
}

// GetConversionRate returns the most recently cached rate regardless of age.
func (s *Store) GetConversionRate(ctx context.Context, currency string) (ConversionRate, bool, error) {
	const query = `
		SELECT currency, rate, coingecko_updated_at, fetched_at
		FROM dogelytics_conversion_rates
		WHERE currency = $1
	`
	return s.scanConversionRate(ctx, query, strings.ToLower(strings.TrimSpace(currency)))
}

// GetFreshConversionRate returns a cached rate when it is newer than maxAge.
func (s *Store) GetFreshConversionRate(ctx context.Context, currency string, maxAge time.Duration) (ConversionRate, bool, error) {
	const query = `
		SELECT currency, rate, coingecko_updated_at, fetched_at
		FROM dogelytics_conversion_rates
		WHERE currency = $1
		AND fetched_at >= $2
	`

	normalisedCurrency := strings.ToLower(strings.TrimSpace(currency))
	threshold := time.Now().Add(-maxAge)

	return s.scanConversionRate(ctx, query, normalisedCurrency, threshold)
}

func (s *Store) scanConversionRate(ctx context.Context, query string, args ...any) (ConversionRate, bool, error) {
	var rate ConversionRate
	var coingeckoUpdatedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&rate.Currency,
		&rate.Rate,
		&coingeckoUpdatedAt,
		&rate.FetchedAt,
	)
	if err == sql.ErrNoRows {
		return ConversionRate{}, false, nil
	}
	if err != nil {
		return ConversionRate{}, false, fmt.Errorf("get conversion rate: %w", err)
	}
	if coingeckoUpdatedAt.Valid {
		rate.CoinGeckoUpdatedAt = &coingeckoUpdatedAt.Time
	}

	return rate, true, nil
}

// UpsertConversionRate inserts or updates a cached conversion rate.
func (s *Store) UpsertConversionRate(ctx context.Context, currency string, rate string, fetchedAt time.Time, coingeckoUpdatedAt *time.Time) (ConversionRate, error) {
	const query = `
		INSERT INTO dogelytics_conversion_rates (currency, rate, coingecko_updated_at, fetched_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (currency) DO UPDATE
		SET rate = EXCLUDED.rate,
			coingecko_updated_at = EXCLUDED.coingecko_updated_at,
			fetched_at = EXCLUDED.fetched_at
		RETURNING currency, rate, coingecko_updated_at, fetched_at
	`

	normalisedCurrency := strings.ToLower(strings.TrimSpace(currency))

	var cachedRate ConversionRate
	var updatedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, query, normalisedCurrency, rate, coingeckoUpdatedAt, fetchedAt).Scan(
		&cachedRate.Currency,
		&cachedRate.Rate,
		&updatedAt,
		&cachedRate.FetchedAt,
	)
	if err != nil {
		return ConversionRate{}, fmt.Errorf("upsert conversion rate: %w", err)
	}
	if updatedAt.Valid {
		cachedRate.CoinGeckoUpdatedAt = &updatedAt.Time
	}

	return cachedRate, nil
}
