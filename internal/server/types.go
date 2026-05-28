package server

// HealthResponse represents the service health payload.
type HealthResponse struct {
	OK     bool  `json:"ok"`
	Height int64 `json:"height"`
}

// DashboardStatsResponse represents the public dashboard metrics payload.
type DashboardStatsResponse struct {
	Available             bool   `json:"available"`
	Message               string `json:"message,omitempty"`
	Height                int64  `json:"height"`
	TotalWalletsChecked   int    `json:"total_wallets_checked"`
	WalletsCheckedLast24h int    `json:"wallets_checked_last_24h"`
	UniqueWalletsLast24h  int    `json:"unique_wallets_last_24h"`
}
