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
	IndexedHeight int64
	BlocksHeight  *int64
	HeadersHeight *int64
	UpdatedAt     *time.Time
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
		IndexedHeight *int64     `json:"indexed_height"`
		BlocksHeight  *int64     `json:"blocks_height"`
		HeadersHeight *int64     `json:"headers_height"`
		UpdatedAt     *time.Time `json:"sync_updated_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return SyncHeights{}, fmt.Errorf("decode indexer sync heights: %w", err)
	}
	if payload.IndexedHeight == nil {
		return SyncHeights{}, fmt.Errorf("indexer sync heights missing indexed_height")
	}

	return SyncHeights{
		IndexedHeight: *payload.IndexedHeight,
		BlocksHeight:  payload.BlocksHeight,
		HeadersHeight: payload.HeadersHeight,
		UpdatedAt:     payload.UpdatedAt,
	}, nil
}
