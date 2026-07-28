package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// User represents a locally authenticated user
type User struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

// APIKey represents an API key record
type APIKey struct {
	ID         string
	UserID     string
	KID        string
	SecretHash string
	CreatedAt  time.Time
	ExpiresAt  sql.NullTime
	RevokedAt  sql.NullTime
}

// CreateUser inserts a new user with email and password hash
func (s *Store) CreateUser(ctx context.Context, id string, email string, passwordHash string) (User, error) {
	query := `
		INSERT INTO dogelytics_users (id, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, created_at
	`
	var u User
	err := s.db.QueryRowContext(ctx, query, id, email, passwordHash).
		Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// GetUserByEmail fetches a user for login
func (s *Store) GetUserByEmail(ctx context.Context, email string) (User, string, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM dogelytics_users
		WHERE email = $1
	`
	var u User
	var hash string
	err := s.db.QueryRowContext(ctx, query, email).
		Scan(&u.ID, &u.Email, &hash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return User{}, "", nil
	}
	if err != nil {
		return User{}, "", fmt.Errorf("get user by email: %w", err)
	}
	return u, hash, nil
}

// GetUserByID returns a user by id
func (s *Store) GetUserByID(ctx context.Context, id string) (User, error) {
	query := `
		SELECT id, email, created_at
		FROM dogelytics_users
		WHERE id = $1
	`
	var u User
	err := s.db.QueryRowContext(ctx, query, id).
		Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return User{}, nil
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// CreateAPIKey inserts a new API key record
func (s *Store) CreateAPIKey(ctx context.Context, id string, userID string, kid string, secretHash string, expiresAt *time.Time) (APIKey, error) {
	query := `
		INSERT INTO dogelytics_api_keys (id, user_id, kid, secret_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, kid, secret_hash, created_at, expires_at, revoked_at
	`
	var k APIKey
	var exp sql.NullTime
	if expiresAt != nil {
		exp = sql.NullTime{Time: *expiresAt, Valid: true}
	}
	err := s.db.QueryRowContext(ctx, query, id, userID, kid, secretHash, exp).
		Scan(&k.ID, &k.UserID, &k.KID, &k.SecretHash, &k.CreatedAt, &k.ExpiresAt, &k.RevokedAt)
	if err != nil {
		return APIKey{}, fmt.Errorf("create api key: %w", err)
	}
	return k, nil
}

// GetAPIKeysByUserID lists keys for a user
func (s *Store) GetAPIKeysByUserID(ctx context.Context, userID string) ([]APIKey, error) {
	query := `
		SELECT id, user_id, kid, secret_hash, created_at, expires_at, revoked_at
		FROM dogelytics_api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.KID, &k.SecretHash, &k.CreatedAt, &k.ExpiresAt, &k.RevokedAt); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list api keys rows: %w", err)
	}
	return keys, nil
}

// GetAPIKeyByKID fetches a key for auth
func (s *Store) GetAPIKeyByKID(ctx context.Context, kid string) (APIKey, error) {
	query := `
		SELECT id, user_id, kid, secret_hash, created_at, expires_at, revoked_at
		FROM dogelytics_api_keys
		WHERE kid = $1
	`
	var k APIKey
	err := s.db.QueryRowContext(ctx, query, kid).
		Scan(&k.ID, &k.UserID, &k.KID, &k.SecretHash, &k.CreatedAt, &k.ExpiresAt, &k.RevokedAt)
	if err == sql.ErrNoRows {
		return APIKey{}, nil
	}
	if err != nil {
		return APIKey{}, fmt.Errorf("get api key by kid: %w", err)
	}
	return k, nil
}

// UpdateAPIKeySecretHash replaces an API key hash if it has not changed since it was read.
func (s *Store) UpdateAPIKeySecretHash(ctx context.Context, kid, previousHash, replacementHash string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE dogelytics_api_keys
		SET secret_hash = $1
		WHERE kid = $2 AND secret_hash = $3
	`, replacementHash, kid, previousHash)
	if err != nil {
		return fmt.Errorf("update api key secret hash: %w", err)
	}
	return nil
}

// RevokeAPIKey soft-deletes a key
func (s *Store) RevokeAPIKey(ctx context.Context, userID string, kid string, revokedAt time.Time) error {
	query := `
		UPDATE dogelytics_api_keys
		SET revoked_at = $1
		WHERE user_id = $2 AND kid = $3 AND revoked_at IS NULL
	`
	res, err := s.db.ExecContext(ctx, query, revokedAt, userID, kid)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	_, _ = res.RowsAffected()
	return nil
}

// UpdateAPIKeyExpiry sets the expiry for a key
func (s *Store) UpdateAPIKeyExpiry(ctx context.Context, userID string, kid string, expiresAt *time.Time) error {
	var (
		query string
		args  []interface{}
	)
	if expiresAt == nil {
		query = `
			UPDATE dogelytics_api_keys
			SET expires_at = NULL
			WHERE user_id = $1 AND kid = $2
		`
		args = []interface{}{userID, kid}
	} else {
		query = `
			UPDATE dogelytics_api_keys
			SET expires_at = $1
			WHERE user_id = $2 AND kid = $3
		`
		args = []interface{}{*expiresAt, userID, kid}
	}
	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update api key expiry: %w", err)
	}
	return nil
}

// GetTotalUsers returns the total number of registered users
func (s *Store) GetTotalUsers(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM dogelytics_users").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get total users: %w", err)
	}
	return count, nil
}

// GetTotalAPIKeys returns the total number of API keys (including revoked/expired)
func (s *Store) GetTotalAPIKeys(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM dogelytics_api_keys").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get total api keys: %w", err)
	}
	return count, nil
}

// GetActiveAPIKeys returns the number of active API keys (not revoked and not expired)
func (s *Store) GetActiveAPIKeys(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) FROM dogelytics_api_keys
		WHERE revoked_at IS NULL
		AND (expires_at IS NULL OR expires_at > now())
	`
	var count int
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get active api keys: %w", err)
	}
	return count, nil
}

// GetRevokedAPIKeys returns the number of revoked API keys
func (s *Store) GetRevokedAPIKeys(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM dogelytics_api_keys WHERE revoked_at IS NOT NULL`
	var count int
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get revoked api keys: %w", err)
	}
	return count, nil
}

// GetExpiredAPIKeys returns the number of expired API keys (not revoked but expired)
func (s *Store) GetExpiredAPIKeys(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) FROM dogelytics_api_keys
		WHERE revoked_at IS NULL
		AND expires_at IS NOT NULL
		AND expires_at <= now()
	`
	var count int
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get expired api keys: %w", err)
	}
	return count, nil
}

// LogRequest logs an API request
func (s *Store) LogRequest(ctx context.Context, clientIP string, apiKey string, walletAddress string, success bool) error {
	query := `
		INSERT INTO dogelytics_request_logs (client_ip, api_key, wallet_address, success)
		VALUES ($1, $2, $3, $4)
	`
	_, err := s.db.ExecContext(ctx, query, clientIP, apiKey, walletAddress, success)
	if err != nil {
		return fmt.Errorf("log request: %w", err)
	}
	return nil
}

// UsageStats holds usage statistics for a time period
type UsageStats struct {
	WalletsChecked int `json:"wallets_checked"`
	UniqueIPs      int `json:"unique_ips"`
	UniqueKeys     int `json:"unique_keys"`
}

// DashboardStats holds the public dashboard summary metrics.
type DashboardStats struct {
	TotalWalletsChecked   int `json:"total_wallets_checked"`
	WalletsCheckedLast24h int `json:"wallets_checked_last_24h"`
	UniqueWalletsChecked  int `json:"unique_wallets_checked"`
	UniqueWalletsLast24h  int `json:"unique_wallets_last_24h"`
}

// TimeSeriesPoint represents a single data point in time
type TimeSeriesPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	WalletsChecked int       `json:"wallets_checked"`
	UniqueIPs      int       `json:"unique_ips"`
	UniqueKeys     int       `json:"unique_keys"`
}

// GetUsageStats returns usage statistics for a given time period
func (s *Store) GetUsageStats(ctx context.Context, hours int, filterType string, filterValues []string) (UsageStats, error) {
	var stats UsageStats
	var query string
	var args []interface{}

	whereClause := "WHERE success = true"
	argIdx := 1

	// Add filter conditions
	if filterType == "keys" && len(filterValues) > 0 {
		placeholders := make([]string, len(filterValues))
		for i, val := range filterValues {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, val)
			argIdx++
		}
		whereClause += fmt.Sprintf(" AND api_key IN (%s)", strings.Join(placeholders, ","))
	}

	// Specific time period
	query = fmt.Sprintf(`
		SELECT 
			COUNT(*) as wallets_checked,
			COUNT(DISTINCT client_ip) as unique_ips,
			COUNT(DISTINCT api_key) FILTER (WHERE api_key IS NOT NULL AND api_key != '') as unique_keys
		FROM dogelytics_request_logs
		%s
		AND timestamp >= now() - interval '1 hour' * $%d
	`, whereClause, argIdx)
	args = append(args, hours)

	err := s.db.QueryRowContext(ctx, query, args...).Scan(&stats.WalletsChecked, &stats.UniqueIPs, &stats.UniqueKeys)
	if err != nil {
		return stats, fmt.Errorf("get usage stats: %w", err)
	}

	return stats, nil
}

// GetDashboardStats returns the public dashboard summary metrics.
func (s *Store) GetDashboardStats(ctx context.Context) (DashboardStats, error) {
	const query = `
		SELECT
			COUNT(*) FILTER (WHERE success = true) AS total_wallets_checked,
			COUNT(*) FILTER (
				WHERE success = true
				AND timestamp >= now() - interval '24 hours'
			) AS wallets_checked_last_24h,
			COUNT(DISTINCT wallet_address) FILTER (
				WHERE success = true
				AND wallet_address IS NOT NULL
				AND wallet_address != ''
			) AS unique_wallets_checked,
			COUNT(DISTINCT wallet_address) FILTER (
				WHERE success = true
				AND timestamp >= now() - interval '24 hours'
				AND wallet_address IS NOT NULL
				AND wallet_address != ''
			) AS unique_wallets_last_24h
		FROM dogelytics_request_logs
	`

	var stats DashboardStats
	err := s.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalWalletsChecked,
		&stats.WalletsCheckedLast24h,
		&stats.UniqueWalletsChecked,
		&stats.UniqueWalletsLast24h,
	)
	if err != nil {
		return DashboardStats{}, fmt.Errorf("get dashboard stats: %w", err)
	}

	return stats, nil
}

// GetUsageTimeSeries returns time-series data for charts
func (s *Store) GetUsageTimeSeries(ctx context.Context, hours int, filterType string, filterValues []string) ([]TimeSeriesPoint, error) {
	var points []TimeSeriesPoint
	var query string
	var args []interface{}
	var interval string

	// Determine interval based on timeframe
	switch hours {
	case 1: // Hour
		interval = "1 minute"
	case 24: // Day
		interval = "1 hour"
	case 168: // Week
		interval = "6 hours"
	case 720: // Month
		interval = "1 day"
	case 8760: // Year
		interval = "1 week"
	default:
		interval = "1 hour"
	}

	// Build time series query
	additionalFilter := ""
	argIdx := 1

	// First arg is always hours
	args = append(args, hours)
	argIdx++

	// Add filter if present
	if filterType == "keys" && len(filterValues) > 0 {
		placeholders := make([]string, len(filterValues))
		for i, val := range filterValues {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, val)
			argIdx++
		}
		additionalFilter = fmt.Sprintf("AND r.api_key IN (%s)", strings.Join(placeholders, ","))
	}

	query = fmt.Sprintf(`
		WITH time_series AS (
			SELECT generate_series(
				date_trunc('hour', now() - interval '1 hour' * $1),
				now(),
				interval '%s'
			) as time_bucket
		)
		SELECT 
			ts.time_bucket,
			COALESCE(COUNT(r.*), 0) as wallets_checked,
			COALESCE(COUNT(DISTINCT r.client_ip), 0) as unique_ips,
			COALESCE(COUNT(DISTINCT r.api_key) FILTER (WHERE r.api_key IS NOT NULL AND r.api_key != ''), 0) as unique_keys
		FROM time_series ts
		LEFT JOIN dogelytics_request_logs r 
			ON r.timestamp >= ts.time_bucket 
			AND r.timestamp < ts.time_bucket + interval '%s'
			AND r.success = true
			%s
		GROUP BY ts.time_bucket
		ORDER BY ts.time_bucket ASC
	`, interval, interval, additionalFilter)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get usage time series: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var point TimeSeriesPoint
		if err := rows.Scan(&point.Timestamp, &point.WalletsChecked, &point.UniqueIPs, &point.UniqueKeys); err != nil {
			return nil, fmt.Errorf("scan time series point: %w", err)
		}
		points = append(points, point)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("time series rows error: %w", err)
	}

	return points, nil
}
