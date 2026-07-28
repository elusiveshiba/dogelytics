package store

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrate applies each embedded schema migration exactly once.
func (s *Store) Migrate(ctx context.Context) error {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS dogelytics_schema_migrations (
			version BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version, err := migrationVersion(entry.Name())
		if err != nil {
			return err
		}
		contents, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		if err := s.applyMigration(ctx, version, entry.Name(), string(contents)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) applyMigration(ctx context.Context, version int64, name, statement string) error {
	transaction, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}
	defer func() { _ = transaction.Rollback() }()

	var applied bool
	if err := transaction.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM dogelytics_schema_migrations WHERE version = $1)
	`, version).Scan(&applied); err != nil {
		return fmt.Errorf("check migration %s: %w", name, err)
	}
	if applied {
		return transaction.Commit()
	}
	if _, err := transaction.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("apply migration %s: %w", name, err)
	}
	if _, err := transaction.ExecContext(ctx, `
		INSERT INTO dogelytics_schema_migrations (version, name) VALUES ($1, $2)
	`, version, name); err != nil {
		return fmt.Errorf("record migration %s: %w", name, err)
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}
	return nil
}

func migrationVersion(name string) (int64, error) {
	prefix, _, ok := strings.Cut(name, "_")
	if !ok {
		return 0, fmt.Errorf("migration %q must start with a numeric version", name)
	}
	version, err := strconv.ParseInt(prefix, 10, 64)
	if err != nil || version < 1 {
		return 0, fmt.Errorf("migration %q has an invalid version", name)
	}
	return version, nil
}
