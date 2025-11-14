package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/dogeorg/doge"
	"github.com/dogeorg/doge/koinu"
)

// Store provides read-only database access for dogelytics
type Store struct {
	db  *sql.DB
	ctx context.Context
}

// Balance represents address balance information
type Balance struct {
	Incoming  koinu.Koinu
	Available koinu.Koinu
	Outgoing  koinu.Koinu
	Current   koinu.Koinu
}

// NewStore creates a new database connection
func NewStore(dbURL string, ctx context.Context) (*Store, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{
		db:  db,
		ctx: ctx,
	}, nil
}

// WithCtx returns a store with a new context
func (s *Store) WithCtx(ctx context.Context) *Store {
	return &Store{
		db:  s.db,
		ctx: ctx,
	}
}

// GetBalance returns the balance for an address
func (s *Store) GetBalance(scriptType doge.ScriptType, address []byte, confirmations int64) (Balance, error) {
	query := `
		SELECT
			(SELECT COALESCE(SUM(u.value),0) FROM utxo u 
			 INNER JOIN tx t ON u.txid = t.txid 
			 WHERE u.script=$1 AND u.kind=$2 
			 AND t.height < (SELECT height FROM resume LIMIT 1)-$3 
			 AND u.spent IS NULL) as available,
			(SELECT COALESCE(SUM(u.value),0) FROM utxo u 
			 INNER JOIN tx t ON u.txid = t.txid 
			 WHERE u.script=$1 AND u.kind=$2 
			 AND t.height >= (SELECT height FROM resume LIMIT 1)-$3 
			 AND u.spent IS NULL) as incoming,
			(SELECT COALESCE(SUM(u.value),0) FROM utxo u 
			 INNER JOIN tx t ON u.txid = t.txid 
			 WHERE u.script=$1 AND u.kind=$2 
			 AND u.spent >= (SELECT height FROM resume LIMIT 1)-$3) as outgoing
	`

	var available, incoming, outgoing int64
	err := s.db.QueryRowContext(s.ctx, query, address, scriptType, confirmations).Scan(&available, &incoming, &outgoing)
	if err != nil {
		return Balance{}, fmt.Errorf("failed to get balance: %w", err)
	}

	return Balance{
		Available: koinu.Koinu(available),
		Incoming:  koinu.Koinu(incoming),
		Outgoing:  koinu.Koinu(outgoing),
		Current:   koinu.Koinu(available + incoming),
	}, nil
}

// GetCurrentHeight returns the current indexed block height
func (s *Store) GetCurrentHeight() (int64, error) {
	var height int64
	err := s.db.QueryRowContext(s.ctx, "SELECT height FROM resume LIMIT 1").Scan(&height)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get current height: %w", err)
	}
	return height, nil
}

// GetResumePoint returns the resume point hash (for health checks)
func (s *Store) GetResumePoint() ([]byte, error) {
	var hash []byte
	err := s.db.QueryRowContext(s.ctx, "SELECT hash FROM resume LIMIT 1").Scan(&hash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get resume point: %w", err)
	}
	return hash, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// QueryContext executes a query that returns rows
func (s *Store) QueryContext(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.QueryContext(s.ctx, query, args...)
}

