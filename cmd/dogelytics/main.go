package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dogeorg/dogelytics/internal/config"
	"github.com/dogeorg/dogelytics/internal/indexer"
	"github.com/dogeorg/dogelytics/internal/server"
	"github.com/dogeorg/dogelytics/internal/store"
)

var dogeboxMetricsClient = &http.Client{Timeout: 10 * time.Second}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		log.Fatal(err)
	}
}

func runServer(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(args)
	if err != nil {
		return err
	}

	indexerClient := indexer.NewSyncClient(cfg.IndexerAPIURL, nil)
	if indexerClient == nil {
		return errors.New("INDEXER_API_URL is required")
	}

	authStore, err := openAuthStore(ctx, cfg)
	if err != nil {
		return err
	}
	if authStore != nil {
		defer func() {
			if closeErr := authStore.Close(); closeErr != nil {
				log.Printf("close dogelytics database: %v", closeErr)
			}
		}()
	}

	if authStore != nil {
		startMetricsReporter(ctx, authStore)
		startAnalyticsMaintenance(ctx, authStore)
	}

	srv := server.NewServer(
		indexerClient,
		indexerClient,
		authStore,
		cfg,
		server.NewRateLimiter(cfg.RateLimit),
		server.NewRateLimiter(cfg.APIKeyRateLimit),
	)

	servers := []namedHTTPServer{
		{
			name:   "api",
			server: newHTTPServer(cfg.BindAddr, srv.APIHandler()),
		},
	}

	if cfg.EnableAdminUI {
		servers = append(servers, namedHTTPServer{
			name:   "admin-ui",
			server: newHTTPServer(listenerAddrForPort(cfg.BindAddr, cfg.AdminUIPort), srv.AdminHandler()),
		})
	}

	if cfg.EnableDashboardUI {
		servers = append(servers, namedHTTPServer{
			name:   "dashboard-ui",
			server: newHTTPServer(listenerAddrForPort(cfg.BindAddr, cfg.DashboardUIPort), srv.DashboardHandler()),
		})
	}

	return runHTTPServers(ctx, servers)
}

func newHTTPServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}

func startMetricsReporter(ctx context.Context, authStore *store.Store) {
	host := os.Getenv("DBX_HOST")
	port := os.Getenv("DBX_PORT")
	if host == "" || port == "" {
		return
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			submitDogeboxMetrics(ctx, authStore, host, port)

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func submitDogeboxMetrics(ctx context.Context, authStore *store.Store, host string, port string) {
	stats, err := authStore.GetDashboardStats(ctx)
	if err != nil {
		log.Printf("Dogelytics metrics unavailable: %v", err)
		return
	}

	metrics := map[string]map[string]int{
		"total_wallets_checked": {
			"value": stats.TotalWalletsChecked,
		},
		"wallets_checked_last_24h": {
			"value": stats.WalletsCheckedLast24h,
		},
		"unique_wallets_checked": {
			"value": stats.UniqueWalletsChecked,
		},
		"unique_wallets_last_24h": {
			"value": stats.UniqueWalletsLast24h,
		},
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Marshal Dogelytics metrics: %v", err)
		return
	}

	url := fmt.Sprintf("http://%s:%s/dbx/metrics", host, port)
	resp, err := dogeboxMetricsClient.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Submit Dogelytics metrics: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Dogelytics metrics returned status %d: %s", resp.StatusCode, string(body))
	}
}

type namedHTTPServer struct {
	name   string
	server *http.Server
}

type serverResult struct {
	name string
	err  error
}

func openAuthStore(ctx context.Context, cfg *config.Config) (*store.Store, error) {
	authStore, err := store.NewStore(cfg.DogelyticsDbURL, ctx)
	if err != nil {
		return nil, err
	}
	if err := authStore.Migrate(ctx); err != nil {
		_ = authStore.Close()
		return nil, err
	}
	if cfg.EnableAnalytics {
		authStore.ConfigureAnalytics(cfg.AnalyticsSecret, cfg.AnalyticsRetention)
		if err := authStore.MigrateLegacyAnalytics(ctx); err != nil {
			_ = authStore.Close()
			return nil, err
		}
	}
	return authStore, nil
}

func startAnalyticsMaintenance(ctx context.Context, authStore *store.Store) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			if err := authStore.PurgeExpiredAnalytics(ctx); err != nil && ctx.Err() == nil {
				log.Printf("Purge expired analytics: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func runHTTPServers(ctx context.Context, servers []namedHTTPServer) error {
	errCh := make(chan serverResult, len(servers))

	for _, namedServer := range servers {
		go func(namedServer namedHTTPServer) {
			log.Printf("Dogelytics %s listening on %s", namedServer.name, namedServer.server.Addr)
			errCh <- serverResult{name: namedServer.name, err: namedServer.server.ListenAndServe()}
		}(namedServer)
	}

	select {
	case result := <-errCh:
		if errors.Is(result.err, http.ErrServerClosed) {
			return nil
		}
		_ = shutdownHTTPServers(servers)
		return fmt.Errorf("%s listen and serve: %w", result.name, result.err)
	case <-ctx.Done():
		if err := shutdownHTTPServers(servers); err != nil {
			return err
		}
		for i := 0; i < len(servers); i++ {
			result := <-errCh
			if !errors.Is(result.err, http.ErrServerClosed) {
				return fmt.Errorf("%s listen and serve: %w", result.name, result.err)
			}
		}
		return nil
	}
}

func shutdownHTTPServers(servers []namedHTTPServer) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, namedServer := range servers {
		if err := namedServer.server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("shutdown %s server: %w", namedServer.name, err)
		}
	}

	return nil
}

func listenerAddrForPort(bindAddr string, port int) string {
	host := ""
	if parsedHost, _, err := net.SplitHostPort(bindAddr); err == nil {
		host = parsedHost
	} else if strings.HasPrefix(bindAddr, ":") {
		host = ""
	} else if idx := strings.LastIndex(bindAddr, ":"); idx > -1 {
		host = bindAddr[:idx]
	} else {
		host = bindAddr
	}

	return net.JoinHostPort(host, strconv.Itoa(port))
}
