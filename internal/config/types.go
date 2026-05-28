package config

import (
	"time"

	"github.com/dogeorg/doge/koinu"
)

// BalanceResponse represents the balance information for a Dogecoin address
type BalanceResponse struct {
	Incoming  koinu.Koinu `json:"incoming"`  // takes N confirmations to become Available
	Available koinu.Koinu `json:"available"` // confirmed balance you can spend
	Outgoing  koinu.Koinu `json:"outgoing"`  // takes N confirmations to become fully Spent
	Current   koinu.Koinu `json:"current"`   // current balance: Incoming + Available
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ConversionResponse represents a DOGE conversion quote for a single currency.
type ConversionResponse struct {
	Currency           string     `json:"currency"`
	Rate               string     `json:"rate"`
	Source             string     `json:"source"`
	Cached             bool       `json:"cached"`
	FetchedAt          time.Time  `json:"fetched_at"`
	CoinGeckoUpdatedAt *time.Time `json:"coingecko_updated_at,omitempty"`
}
