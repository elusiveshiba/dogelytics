package server

import (
	"context"
	"log"
	"time"
)

type chainProgress struct {
	IndexerHeight     int64
	CoreBlocksHeight  int64
	CoreHeadersHeight int64
	CoreSyncUpdatedAt *time.Time
	IndexedPercent    *float64
}

func (s *Server) loadChainProgress(ctx context.Context) (chainProgress, bool) {
	if s.syncSource == nil {
		return chainProgress{}, false
	}

	syncHeights, err := s.syncSource.SyncHeights(ctx)
	if err != nil {
		log.Printf("[Dogelytics] failed to load indexer sync heights: %v", err)
		return chainProgress{}, false
	}

	progress := chainProgress{
		IndexerHeight:     syncHeights.IndexerHeight,
		CoreSyncUpdatedAt: syncHeights.CoreSyncUpdatedAt,
	}

	if syncHeights.CoreBlocksHeight != nil {
		progress.CoreBlocksHeight = *syncHeights.CoreBlocksHeight
	}
	if syncHeights.CoreHeadersHeight != nil {
		progress.CoreHeadersHeight = *syncHeights.CoreHeadersHeight
		progress.IndexedPercent = percentOf(progress.IndexerHeight, progress.CoreHeadersHeight)
	}

	return progress, true
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
