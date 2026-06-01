package server

import (
	"context"
	"log"
)

type chainProgress struct {
	IndexedHeight    int64
	BlocksHeight     int64
	HeadersHeight    int64
	IndexedPercent   *float64
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
		IndexedHeight: syncHeights.IndexedHeight,
	}

	if syncHeights.BlocksHeight != nil {
		progress.BlocksHeight = *syncHeights.BlocksHeight
	}
	if syncHeights.HeadersHeight != nil {
		progress.HeadersHeight = *syncHeights.HeadersHeight
		progress.IndexedPercent = percentOf(progress.IndexedHeight, progress.HeadersHeight)
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
