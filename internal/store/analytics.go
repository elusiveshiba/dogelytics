package store

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"
)

// ConfigureAnalytics enables HMAC-fingerprinted analytics and raw-event retention.
func (s *Store) ConfigureAnalytics(secret string, retentionDays int) {
	s.analyticsSecret = []byte(secret)
	s.analyticsRetention = time.Duration(retentionDays) * 24 * time.Hour
}

func (s *Store) analyticsFingerprint(value string) []byte {
	if value == "" || len(s.analyticsSecret) == 0 {
		return nil
	}
	digest := hmac.New(sha256.New, s.analyticsSecret)
	_, _ = digest.Write([]byte(value))
	return digest.Sum(nil)
}

// MigrateLegacyAnalytics fingerprints and clears legacy raw identifiers once.
func (s *Store) MigrateLegacyAnalytics(ctx context.Context) error {
	if len(s.analyticsSecret) == 0 {
		return nil
	}
	transaction, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin legacy analytics migration: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	var complete bool
	if err := transaction.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM dogelytics_analytics_state WHERE key = 'legacy_fingerprints_complete')
	`).Scan(&complete); err != nil {
		return fmt.Errorf("check legacy analytics migration: %w", err)
	}
	if complete {
		return transaction.Commit()
	}

	type legacyEvent struct {
		id     int64
		client sql.NullString
		wallet sql.NullString
	}
	rows, err := transaction.QueryContext(ctx, `
		SELECT id, client_ip, wallet_address
		FROM dogelytics_request_logs
		WHERE client_ip IS NOT NULL OR wallet_address IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("read legacy analytics: %w", err)
	}
	var events []legacyEvent
	for rows.Next() {
		var event legacyEvent
		if err := rows.Scan(&event.id, &event.client, &event.wallet); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scan legacy analytics: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close legacy analytics rows: %w", err)
	}

	for _, event := range events {
		if _, err := transaction.ExecContext(ctx, `
			UPDATE dogelytics_request_logs
			SET client_fingerprint = $1,
				wallet_fingerprint = $2,
				client_ip = NULL,
				wallet_address = NULL
			WHERE id = $3
		`, s.analyticsFingerprint(event.client.String), s.analyticsFingerprint(event.wallet.String), event.id); err != nil {
			return fmt.Errorf("fingerprint legacy analytics: %w", err)
		}
	}

	if _, err := transaction.ExecContext(ctx, `
		UPDATE dogelytics_analytics_totals
		SET successful_requests = (SELECT COUNT(*) FROM dogelytics_request_logs WHERE success = true)
		WHERE id = 1;
		TRUNCATE dogelytics_analytics_wallets;
		INSERT INTO dogelytics_analytics_wallets (wallet_fingerprint, first_seen, last_seen)
		SELECT wallet_fingerprint, MIN(timestamp), MAX(timestamp)
		FROM dogelytics_request_logs
		WHERE success = true AND wallet_fingerprint IS NOT NULL
		GROUP BY wallet_fingerprint;
		TRUNCATE dogelytics_analytics_hourly;
		INSERT INTO dogelytics_analytics_hourly (bucket, api_key, successful_requests)
		SELECT date_trunc('hour', timestamp), COALESCE(api_key, ''), COUNT(*)
		FROM dogelytics_request_logs
		WHERE success = true
		GROUP BY date_trunc('hour', timestamp), COALESCE(api_key, '');
		INSERT INTO dogelytics_analytics_state (key, value)
		VALUES ('legacy_fingerprints_complete', 'true')
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now();
	`); err != nil {
		return fmt.Errorf("rebuild private analytics: %w", err)
	}
	if _, err := transaction.ExecContext(ctx, `
		ALTER TABLE dogelytics_request_logs
			DROP COLUMN client_ip,
			DROP COLUMN wallet_address
	`); err != nil {
		return fmt.Errorf("remove legacy analytics identifiers: %w", err)
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit legacy analytics migration: %w", err)
	}
	return nil
}

// LogRequest records a fingerprinted request and updates durable aggregates.
func (s *Store) LogRequest(ctx context.Context, clientIP, apiKey, walletAddress string, success bool) error {
	if len(s.analyticsSecret) == 0 {
		return nil
	}
	clientFingerprint := s.analyticsFingerprint(clientIP)
	walletFingerprint := s.analyticsFingerprint(walletAddress)
	transaction, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin request log: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	if _, err := transaction.ExecContext(ctx, `
		INSERT INTO dogelytics_request_logs
			(client_fingerprint, api_key, wallet_fingerprint, success)
		VALUES ($1, NULLIF($2, ''), $3, $4)
	`, clientFingerprint, apiKey, walletFingerprint, success); err != nil {
		return fmt.Errorf("log request: %w", err)
	}
	if success {
		if _, err := transaction.ExecContext(ctx, `
			UPDATE dogelytics_analytics_totals
			SET successful_requests = successful_requests + 1
			WHERE id = 1
		`); err != nil {
			return fmt.Errorf("update analytics total: %w", err)
		}
		if _, err := transaction.ExecContext(ctx, `
			INSERT INTO dogelytics_analytics_hourly (bucket, api_key, successful_requests)
			VALUES (date_trunc('hour', now()), $1, 1)
			ON CONFLICT (bucket, api_key) DO UPDATE
			SET successful_requests = dogelytics_analytics_hourly.successful_requests + 1
		`, apiKey); err != nil {
			return fmt.Errorf("update hourly analytics: %w", err)
		}
		if len(walletFingerprint) != 0 {
			if _, err := transaction.ExecContext(ctx, `
				INSERT INTO dogelytics_analytics_wallets (wallet_fingerprint, first_seen, last_seen)
				VALUES ($1, now(), now())
				ON CONFLICT (wallet_fingerprint) DO UPDATE SET last_seen = EXCLUDED.last_seen
			`, walletFingerprint); err != nil {
				return fmt.Errorf("update unique wallet analytics: %w", err)
			}
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit request log: %w", err)
	}
	return nil
}

// PurgeExpiredAnalytics removes fingerprinted request events after their retention period.
func (s *Store) PurgeExpiredAnalytics(ctx context.Context) error {
	if s.analyticsRetention <= 0 {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM dogelytics_request_logs WHERE timestamp < $1`, time.Now().Add(-s.analyticsRetention))
	if err != nil {
		return fmt.Errorf("purge expired analytics: %w", err)
	}
	return nil
}
