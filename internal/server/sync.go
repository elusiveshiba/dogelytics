package server

import (
	"context"
	"log"
)

type chainProgress struct {
	CoreHeight       int64
	BlockchainHeight int64
	IndexedPercent   *float64
}

func (s *Server) loadChainProgress(ctx context.Context, indexedHeight int64) (chainProgress, bool) {
	if s.coreClient == nil {
		return chainProgress{}, false
	}

	coreState, err := s.coreClient.ChainState(ctx)
	if err != nil {
		log.Printf("[Dogelytics] failed to load blockchain height: %v", err)
		return chainProgress{}, false
	}
	if coreState.BlockchainHeight <= 0 {
		return chainProgress{}, false
	}

	return chainProgress{
		CoreHeight:       coreState.CoreHeight,
		BlockchainHeight: coreState.BlockchainHeight,
		IndexedPercent:   percentOf(indexedHeight, coreState.BlockchainHeight),
	}, true
}

func percentOf(height int64, total int64) *float64 {
	if total <= 0 {
		return nil
	}

	percent := (float64(height) / float64(total)) * 100
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}
	return &percent
}
