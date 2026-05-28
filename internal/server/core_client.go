package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CoreClient reads chain state from Dogecoin Core.
type CoreClient interface {
	BlockchainHeight(ctx context.Context) (int64, error)
}

// CoreRPCClient talks to Dogecoin Core's JSON-RPC API.
type CoreRPCClient struct {
	url        string
	user       string
	password   string
	httpClient *http.Client
}

// NewCoreRPCClient creates a Core RPC client when a URL is configured.
func NewCoreRPCClient(url string, user string, password string, httpClient *http.Client) *CoreRPCClient {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &CoreRPCClient{
		url:        url,
		user:       user,
		password:   password,
		httpClient: httpClient,
	}
}

// BlockchainHeight returns the best known chain height from Core.
func (c *CoreRPCClient) BlockchainHeight(ctx context.Context) (int64, error) {
	requestBody := map[string]any{
		"jsonrpc": "1.0",
		"id":      "dogelytics",
		"method":  "getblockchaininfo",
		"params":  []any{},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return 0, fmt.Errorf("marshal core rpc request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create core rpc request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.user != "" || c.password != "" {
		req.SetBasicAuth(c.user, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request core rpc: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("core rpc returned status %d", resp.StatusCode)
	}

	var rpcResponse struct {
		Result struct {
			Blocks  int64 `json:"blocks"`
			Headers int64 `json:"headers"`
		} `json:"result"`
		Error any `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
		return 0, fmt.Errorf("decode core rpc response: %w", err)
	}
	if rpcResponse.Error != nil {
		return 0, fmt.Errorf("core rpc error: %v", rpcResponse.Error)
	}

	if rpcResponse.Result.Headers > rpcResponse.Result.Blocks {
		return rpcResponse.Result.Headers, nil
	}
	return rpcResponse.Result.Blocks, nil
}
