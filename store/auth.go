package store

import (
	"database/sql"
	"fmt"
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

// EnsureAuthSchema creates auth-related tables if they do not exist
func (s *Store) EnsureAuthSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS dogelytics_users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS dogelytics_api_keys (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES dogelytics_users(id) ON DELETE CASCADE,
			kid TEXT UNIQUE NOT NULL,
			secret_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			expires_at TIMESTAMPTZ,
			revoked_at TIMESTAMPTZ
		);
		CREATE INDEX IF NOT EXISTS idx_dgl_users_email ON dogelytics_users(email);
		CREATE INDEX IF NOT EXISTS idx_dgl_apikeys_user_id ON dogelytics_api_keys(user_id);
		CREATE INDEX IF NOT EXISTS idx_dgl_apikeys_expires_at ON dogelytics_api_keys(expires_at);
	`
	_, err := s.db.ExecContext(s.ctx, schema)
	if err != nil {
		return fmt.Errorf("ensure auth schema: %w", err)
	}
	return nil
}

// CreateUser inserts a new user with email and password hash
func (s *Store) CreateUser(id string, email string, passwordHash string) (User, error) {
	query := `
		INSERT INTO dogelytics_users (id, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, created_at
	`
	var u User
	err := s.db.QueryRowContext(s.ctx, query, id, email, passwordHash).
		Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// GetUserByEmail fetches a user for login
func (s *Store) GetUserByEmail(email string) (User, string, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM dogelytics_users
		WHERE email = $1
	`
	var u User
	var hash string
	err := s.db.QueryRowContext(s.ctx, query, email).
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
func (s *Store) GetUserByID(id string) (User, error) {
	query := `
		SELECT id, email, created_at
		FROM dogelytics_users
		WHERE id = $1
	`
	var u User
	err := s.db.QueryRowContext(s.ctx, query, id).
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
func (s *Store) CreateAPIKey(id string, userID string, kid string, secretHash string, expiresAt *time.Time) (APIKey, error) {
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
	err := s.db.QueryRowContext(s.ctx, query, id, userID, kid, secretHash, exp).
		Scan(&k.ID, &k.UserID, &k.KID, &k.SecretHash, &k.CreatedAt, &k.ExpiresAt, &k.RevokedAt)
	if err != nil {
		return APIKey{}, fmt.Errorf("create api key: %w", err)
	}
	return k, nil
}

// GetAPIKeysByUserID lists keys for a user
func (s *Store) GetAPIKeysByUserID(userID string) ([]APIKey, error) {
	query := `
		SELECT id, user_id, kid, secret_hash, created_at, expires_at, revoked_at
		FROM dogelytics_api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(s.ctx, query, userID)
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
func (s *Store) GetAPIKeyByKID(kid string) (APIKey, error) {
	query := `
		SELECT id, user_id, kid, secret_hash, created_at, expires_at, revoked_at
		FROM dogelytics_api_keys
		WHERE kid = $1
	`
	var k APIKey
	err := s.db.QueryRowContext(s.ctx, query, kid).
		Scan(&k.ID, &k.UserID, &k.KID, &k.SecretHash, &k.CreatedAt, &k.ExpiresAt, &k.RevokedAt)
	if err == sql.ErrNoRows {
		return APIKey{}, nil
	}
	if err != nil {
		return APIKey{}, fmt.Errorf("get api key by kid: %w", err)
	}
	return k, nil
}

// RevokeAPIKey soft-deletes a key
func (s *Store) RevokeAPIKey(userID string, kid string, revokedAt time.Time) error {
	query := `
		UPDATE dogelytics_api_keys
		SET revoked_at = $1
		WHERE user_id = $2 AND kid = $3 AND revoked_at IS NULL
	`
	res, err := s.db.ExecContext(s.ctx, query, revokedAt, userID, kid)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	_, _ = res.RowsAffected()
	return nil
}

// UpdateAPIKeyExpiry sets the expiry for a key
func (s *Store) UpdateAPIKeyExpiry(userID string, kid string, expiresAt *time.Time) error {
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
	_, err := s.db.ExecContext(s.ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update api key expiry: %w", err)
	}
	return nil
}


