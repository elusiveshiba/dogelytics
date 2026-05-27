package indexer

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dogeorg/doge"
	"github.com/dogeorg/doge/koinu"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Store provides read-only access to the Dogecoin Indexer database.
type Store struct {
	db *sql.DB
}

// Balance represents address balance information.
type Balance struct {
	Incoming  koinu.Koinu
	Available koinu.Koinu
	Outgoing  koinu.Koinu
	Current   koinu.Koinu
}

// NewStore creates a new indexer database connection.
func NewStore(ctx context.Context, dbURL string) (*Store, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("open indexer database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping indexer database: %w", err)
	}

	return &Store{db: db}, nil
}

// GetBalance returns the balance for an address.
func (s *Store) GetBalance(ctx context.Context, scriptType doge.ScriptType, address []byte, confirmations int64) (Balance, error) {
	query := `
		WITH cutoff AS (
			SELECT height - $3 AS min_confirmed_height
			FROM resume
			LIMIT 1
		)
		SELECT
			(
				SELECT COALESCE(SUM(u.value), 0)
				FROM utxo u
				INNER JOIN tx t ON u.txid = t.txid
				CROSS JOIN cutoff c
				WHERE u.script = $1
					AND u.kind = $2
					AND t.height < c.min_confirmed_height
					AND u.spent IS NULL
			) AS available,
			(
				SELECT COALESCE(SUM(u.value), 0)
				FROM utxo u
				INNER JOIN tx t ON u.txid = t.txid
				CROSS JOIN cutoff c
				WHERE u.script = $1
					AND u.kind = $2
					AND t.height >= c.min_confirmed_height
					AND u.spent IS NULL
			) AS incoming,
			(
				SELECT COALESCE(SUM(u.value), 0)
				FROM utxo u
				CROSS JOIN cutoff c
				WHERE u.script = $1
					AND u.kind = $2
					AND u.spent >= c.min_confirmed_height
			) AS outgoing
	`

	var available, incoming, outgoing int64
	if err := s.db.QueryRowContext(ctx, query, address, scriptType, confirmations).Scan(&available, &incoming, &outgoing); err != nil {
		return Balance{}, fmt.Errorf("get balance: %w", err)
	}

	return Balance{
		Available: koinu.Koinu(available),
		Incoming:  koinu.Koinu(incoming),
		Outgoing:  koinu.Koinu(outgoing),
		Current:   koinu.Koinu(available + incoming),
	}, nil
}

// CurrentHeight returns the current indexed block height.
func (s *Store) CurrentHeight(ctx context.Context) (int64, error) {
	var height int64
	err := s.db.QueryRowContext(ctx, "SELECT height FROM resume LIMIT 1").Scan(&height)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get current height: %w", err)
	}
	return height, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
