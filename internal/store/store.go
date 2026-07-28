package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

const (
	defaultMaxOpenConnections = 10
	defaultMaxIdleConnections = 5
	defaultConnectionLifetime = 30 * time.Minute
)

// Store provides PostgreSQL-backed application storage.
type Store struct {
	db                 *sql.DB
	analyticsSecret    []byte
	analyticsRetention time.Duration
}

// NewStore opens and verifies a bounded PostgreSQL connection pool.
func NewStore(dbURL string, ctx context.Context) (*Store, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(defaultMaxOpenConnections)
	db.SetMaxIdleConns(defaultMaxIdleConnections)
	db.SetConnMaxLifetime(defaultConnectionLifetime)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Store{db: db}, nil
}

// Ping verifies that the database is reachable.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close closes the database connection pool.
func (s *Store) Close() error {
	return s.db.Close()
}
