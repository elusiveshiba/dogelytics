package store

import (
	"bytes"
	"context"
	"os"
	"testing"
)

func TestAnalyticsFingerprintIsKeyedAndDeterministic(t *testing.T) {
	storage := &Store{}
	storage.ConfigureAnalytics("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 30)
	first := storage.analyticsFingerprint("203.0.113.10")
	second := storage.analyticsFingerprint("203.0.113.10")
	if !bytes.Equal(first, second) {
		t.Fatal("expected deterministic fingerprint")
	}
	if bytes.Contains(first, []byte("203.0.113.10")) || len(first) != 32 {
		t.Fatalf("unexpected fingerprint representation: %x", first)
	}

	other := &Store{}
	other.ConfigureAnalytics("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", 30)
	if bytes.Equal(first, other.analyticsFingerprint("203.0.113.10")) {
		t.Fatal("expected different secrets to produce different fingerprints")
	}
}

func TestPrivateAnalyticsIntegration(t *testing.T) {
	databaseURL := os.Getenv("DOGELYTICS_TEST_DBURL")
	if databaseURL == "" {
		t.Skip("DOGELYTICS_TEST_DBURL is not configured")
	}
	ctx := context.Background()
	storage, err := NewStore(databaseURL, ctx)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer storage.Close()
	if err := storage.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	storage.ConfigureAnalytics("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 1)
	if _, err := storage.db.ExecContext(ctx, `
		TRUNCATE dogelytics_request_logs, dogelytics_analytics_wallets, dogelytics_analytics_hourly, dogelytics_analytics_state RESTART IDENTITY;
		UPDATE dogelytics_analytics_totals SET successful_requests = 0 WHERE id = 1;
		INSERT INTO dogelytics_request_logs (client_ip, wallet_address, success)
		VALUES ('203.0.113.10', 'DLegacyWallet', true);
	`); err != nil {
		t.Fatalf("prepare legacy analytics: %v", err)
	}
	if err := storage.MigrateLegacyAnalytics(ctx); err != nil {
		t.Fatalf("migrate legacy analytics: %v", err)
	}

	var clientFingerprint, walletFingerprint []byte
	if err := storage.db.QueryRowContext(ctx, `
		SELECT client_fingerprint, wallet_fingerprint
		FROM dogelytics_request_logs LIMIT 1
	`).Scan(&clientFingerprint, &walletFingerprint); err != nil {
		t.Fatalf("read migrated analytics: %v", err)
	}
	if len(clientFingerprint) != 32 || len(walletFingerprint) != 32 {
		t.Fatalf("legacy identifiers were not privately migrated")
	}
	var legacyColumns int
	if err := storage.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_name = 'dogelytics_request_logs'
			AND column_name IN ('client_ip', 'wallet_address')
	`).Scan(&legacyColumns); err != nil {
		t.Fatalf("check legacy columns: %v", err)
	}
	if legacyColumns != 0 {
		t.Fatalf("expected legacy identifier columns to be removed")
	}

	if err := storage.LogRequest(ctx, "198.51.100.4", "kid-one", "DWalletOne", true); err != nil {
		t.Fatalf("log first request: %v", err)
	}
	if err := storage.LogRequest(ctx, "198.51.100.5", "kid-one", "DWalletOne", true); err != nil {
		t.Fatalf("log second request: %v", err)
	}
	stats, err := storage.GetDashboardStats(ctx)
	if err != nil {
		t.Fatalf("dashboard stats: %v", err)
	}
	if stats.TotalWalletsChecked != 3 || stats.UniqueWalletsChecked != 2 {
		t.Fatalf("unexpected dashboard stats: %+v", stats)
	}

	if _, err := storage.db.ExecContext(ctx, `UPDATE dogelytics_request_logs SET timestamp = now() - interval '2 days'`); err != nil {
		t.Fatalf("age events: %v", err)
	}
	if err := storage.PurgeExpiredAnalytics(ctx); err != nil {
		t.Fatalf("purge analytics: %v", err)
	}
	var remaining int
	if err := storage.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dogelytics_request_logs`).Scan(&remaining); err != nil {
		t.Fatalf("count retained events: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected expired events to be purged, got %d", remaining)
	}
}
