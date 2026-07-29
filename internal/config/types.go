package config

import (
	"encoding/json"
	"errors"
	"regexp"
	"time"
)

var dogeAmountPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]{1,8})?$`)

// DogeAmount is an exact decimal DOGE amount returned by the indexer.
// It deliberately remains a string because address-level totals can exceed int64.
type DogeAmount string

// UnmarshalJSON accepts the indexer's quoted decimal representation.
func (amount *DogeAmount) UnmarshalJSON(data []byte) error {
	if amount == nil {
		return errors.New("cannot unmarshal DOGE amount into nil receiver")
	}
	if len(data) == 0 || data[0] != '"' {
		return errors.New("DOGE amount must be a JSON string")
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if !dogeAmountPattern.MatchString(value) {
		return errors.New("invalid DOGE amount")
	}

	*amount = DogeAmount(value)
	return nil
}

// BalanceResponse represents the balance information for a Dogecoin address
type BalanceResponse struct {
	Incoming  DogeAmount `json:"incoming"`  // takes six confirmations to become Available
	Available DogeAmount `json:"available"` // confirmed balance you can spend
	Outgoing  DogeAmount `json:"outgoing"`  // takes six confirmations to become fully Spent
	Current   DogeAmount `json:"current"`   // current balance: Incoming + Available
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
