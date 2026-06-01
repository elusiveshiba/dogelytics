package server

// HealthResponse represents the service health payload.
type HealthResponse struct {
	OK            bool  `json:"ok"`
	IndexerHeight int64 `json:"indexer_height"`
	BlocksHeight  int64 `json:"blocks_height,omitempty"`
	HeadersHeight int64 `json:"headers_height,omitempty"`
}

// DashboardStatsResponse represents the public dashboard metrics payload.
type DashboardStatsResponse struct {
	Available             bool     `json:"available"`
	Message               string   `json:"message,omitempty"`
	Height                int64    `json:"height"`
	BlocksHeight          int64    `json:"blocks_height,omitempty"`
	HeadersHeight         int64    `json:"headers_height,omitempty"`
	IndexedPercent        *float64 `json:"indexed_percent,omitempty"`
	TotalWalletsChecked   int      `json:"total_wallets_checked"`
	WalletsCheckedLast24h int      `json:"wallets_checked_last_24h"`
	UniqueWalletsChecked  int      `json:"unique_wallets_checked"`
	UniqueWalletsLast24h  int      `json:"unique_wallets_last_24h"`
}
