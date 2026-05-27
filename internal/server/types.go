package server

// HealthResponse represents the service health payload.
type HealthResponse struct {
	OK     bool  `json:"ok"`
	Height int64 `json:"height"`
}
