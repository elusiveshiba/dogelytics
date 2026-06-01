package server

import "time"

// HealthResponse represents the service health payload.
type HealthResponse struct {
	OK                bool       `json:"ok"`
	IndexerHeight     int64      `json:"indexer_height"`
	CoreBlocksHeight  int64      `json:"core_blocks_height,omitempty"`
	CoreHeadersHeight int64      `json:"core_headers_height,omitempty"`
	CoreSyncUpdatedAt *time.Time `json:"core_sync_updated_at,omitempty"`
}

// DashboardStatsResponse represents the public dashboard metrics payload.
type DashboardStatsResponse struct {
	Available             bool       `json:"available"`
	Message               string     `json:"message,omitempty"`
	IndexerHeight         int64      `json:"indexer_height"`
	CoreBlocksHeight      int64      `json:"core_blocks_height,omitempty"`
	CoreHeadersHeight     int64      `json:"core_headers_height,omitempty"`
	CoreSyncUpdatedAt     *time.Time `json:"core_sync_updated_at,omitempty"`
	IndexedPercent        *float64   `json:"indexed_percent,omitempty"`
	TotalWalletsChecked   int        `json:"total_wallets_checked"`
	WalletsCheckedLast24h int        `json:"wallets_checked_last_24h"`
	UniqueWalletsChecked  int        `json:"unique_wallets_checked"`
	UniqueWalletsLast24h  int        `json:"unique_wallets_last_24h"`
}
