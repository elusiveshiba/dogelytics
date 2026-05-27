package server

import (
	"context"

	"github.com/dogeorg/doge"
	"github.com/dogeorg/dogelytics/internal/config"
	"github.com/dogeorg/dogelytics/internal/indexer"
	"github.com/dogeorg/dogelytics/internal/store"
)

// BalanceStore defines the indexer operations needed by the HTTP API.
type BalanceStore interface {
	GetBalance(ctx context.Context, scriptType doge.ScriptType, address []byte, confirmations int64) (indexer.Balance, error)
	CurrentHeight(ctx context.Context) (int64, error)
}

// Server handles HTTP requests for balance and health endpoints.
type Server struct {
	indexerStore BalanceStore
	authStore    *store.Store
	config       *config.Config
	ipLimiter    *RateLimiter
	apiLimiter   *RateLimiter
}

// NewServer creates a new Server instance.
func NewServer(indexerStore BalanceStore, authStore *store.Store, cfg *config.Config, ipLimiter *RateLimiter, apiLimiter *RateLimiter) *Server {
	return &Server{
		indexerStore: indexerStore,
		authStore:    authStore,
		config:       cfg,
		ipLimiter:    ipLimiter,
		apiLimiter:   apiLimiter,
	}
}
