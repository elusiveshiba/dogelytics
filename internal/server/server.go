package server

import (
	"context"
	"sync"

	"github.com/dogeorg/dogelytics/internal/config"
	"github.com/dogeorg/dogelytics/internal/indexer"
	"github.com/dogeorg/dogelytics/internal/store"
)

// BalanceStore defines the indexer operations needed by the HTTP API.
type BalanceStore interface {
	GetBalance(ctx context.Context, address string) (indexer.Balance, error)
	CurrentHeight(ctx context.Context) (int64, error)
}

// SyncSource provides indexer-owned sync heights over HTTP.
type SyncSource interface {
	SyncHeights(ctx context.Context) (indexer.SyncHeights, error)
}

// Server handles HTTP requests for the API and optional UI surfaces.
type Server struct {
	indexerStore     BalanceStore
	syncSource       SyncSource
	authStore        *store.Store
	conversionStore  ConversionStore
	conversionClient ConversionClient
	config           *config.Config
	ipLimiter        *RateLimiter
	apiLimiter       *RateLimiter
	authLimiter      *RateLimiter
	conversionMu     sync.Mutex
	conversionFlight map[string]*conversionRefresh
}

// NewServer creates a new Server instance.
func NewServer(indexerStore BalanceStore, syncSource SyncSource, authStore *store.Store, cfg *config.Config, ipLimiter *RateLimiter, apiLimiter *RateLimiter) *Server {
	return &Server{
		indexerStore:     indexerStore,
		syncSource:       syncSource,
		authStore:        authStore,
		conversionStore:  authStore,
		conversionClient: NewCoinGeckoClient(nil),
		config:           cfg,
		ipLimiter:        ipLimiter,
		apiLimiter:       apiLimiter,
		authLimiter:      NewRateLimiter(10),
		conversionFlight: make(map[string]*conversionRefresh),
	}
}

func (s *Server) hasAuthStore() bool {
	return s.authStore != nil
}
