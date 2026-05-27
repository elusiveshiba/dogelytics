package config

import (
	"os"
	"path/filepath"
	"testing"
)

func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()
	old, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func TestLoadDotEnvIfPresent_LoadsKeyValues(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := "INDEXER_DBURL=postgres://user:pass@localhost:5432/indexer?sslmode=disable\n"

	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	unsetEnvForTest(t, "INDEXER_DBURL")
	loadDotEnvIfPresent(envPath)

	got := os.Getenv("INDEXER_DBURL")
	want := "postgres://user:pass@localhost:5432/indexer?sslmode=disable"
	if got != want {
		t.Fatalf("INDEXER_DBURL mismatch: got %q, want %q", got, want)
	}
}

func TestLoadDotEnvIfPresent_PreservesExistingEnv(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := "INDEXER_DBURL=postgres://fromfile@localhost:5432/indexer?sslmode=disable\n"

	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	t.Setenv("INDEXER_DBURL", "postgres://fromenv@localhost:5432/indexer?sslmode=disable")
	loadDotEnvIfPresent(envPath)

	got := os.Getenv("INDEXER_DBURL")
	want := "postgres://fromenv@localhost:5432/indexer?sslmode=disable"
	if got != want {
		t.Fatalf("INDEXER_DBURL should preserve existing env: got %q, want %q", got, want)
	}
}

func TestLoadDotEnvIfPresent_HandlesExportAndQuotes(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := "export INDEXER_DBURL='postgres://quoted@localhost:5432/indexer?sslmode=disable'\n"

	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	unsetEnvForTest(t, "INDEXER_DBURL")
	loadDotEnvIfPresent(envPath)

	got := os.Getenv("INDEXER_DBURL")
	want := "postgres://quoted@localhost:5432/indexer?sslmode=disable"
	if got != want {
		t.Fatalf("INDEXER_DBURL mismatch with export/quotes: got %q, want %q", got, want)
	}
}
