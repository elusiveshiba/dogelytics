package store

import (
	"context"
	"os"
	"testing"
)

func TestMigrationVersion(t *testing.T) {
	version, err := migrationVersion("001_initial.sql")
	if err != nil {
		t.Fatalf("parse migration version: %v", err)
	}
	if version != 1 {
		t.Fatalf("version mismatch: got %d, want 1", version)
	}
	for _, name := range []string{"initial.sql", "000_invalid.sql", "abc_invalid.sql"} {
		if _, err := migrationVersion(name); err == nil {
			t.Fatalf("expected %q to fail", name)
		}
	}
}

func TestMigrateIntegration(t *testing.T) {
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
		t.Fatalf("first migration: %v", err)
	}
	if err := storage.Migrate(ctx); err != nil {
		t.Fatalf("idempotent migration: %v", err)
	}

	var count int
	if err := storage.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dogelytics_schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count < 1 {
		t.Fatalf("expected at least one migration, got %d", count)
	}
}
