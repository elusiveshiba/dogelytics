package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SyncHeights describes the indexer-owned sync heights exposed over HTTP.
type SyncHeights struct {
	IndexerHeight     int64
	CoreBlocksHeight  *int64
	CoreHeadersHeight *int64
	CoreSyncUpdatedAt *time.Time
}

// SyncClient fetches sync-height data from the indexer HTTP API.
type SyncClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewSyncClient(baseURL string, httpClient *http.Client) *SyncClient {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &SyncClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *SyncClient) SyncHeights(ctx context.Context) (SyncHeights, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/height", nil)
	if err != nil {
		return SyncHeights{}, fmt.Errorf("create indexer sync request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return SyncHeights{}, fmt.Errorf("request indexer sync heights: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SyncHeights{}, fmt.Errorf("indexer sync heights returned status %d", resp.StatusCode)
	}

	var payload struct {
		Height            int64      `json:"height"`
		CoreBlocksHeight  *int64     `json:"core_blocks_height"`
		CoreHeadersHeight *int64     `json:"core_headers_height"`
		CoreSyncUpdatedAt *time.Time `json:"core_sync_updated_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return SyncHeights{}, fmt.Errorf("decode indexer sync heights: %w", err)
	}

	return SyncHeights{
		IndexerHeight:     payload.Height,
		CoreBlocksHeight:  payload.CoreBlocksHeight,
		CoreHeadersHeight: payload.CoreHeadersHeight,
		CoreSyncUpdatedAt: payload.CoreSyncUpdatedAt,
	}, nil
}
