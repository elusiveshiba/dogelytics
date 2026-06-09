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
	content := "INDEXER_API_URL=http://indexer.local:8000\n"

	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	unsetEnvForTest(t, "INDEXER_API_URL")
	loadDotEnvIfPresent(envPath)

	got := os.Getenv("INDEXER_API_URL")
	want := "http://indexer.local:8000"
	if got != want {
		t.Fatalf("INDEXER_API_URL mismatch: got %q, want %q", got, want)
	}
}

func TestLoadDotEnvIfPresent_PreservesExistingEnv(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := "INDEXER_API_URL=http://from-file:8000\n"

	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	t.Setenv("INDEXER_API_URL", "http://from-env:8000")
	loadDotEnvIfPresent(envPath)

	got := os.Getenv("INDEXER_API_URL")
	want := "http://from-env:8000"
	if got != want {
		t.Fatalf("INDEXER_API_URL should preserve existing env: got %q, want %q", got, want)
	}
}

func TestLoadDotEnvIfPresent_HandlesExportAndQuotes(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := "export INDEXER_API_URL='http://quoted:8000'\n"

	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	unsetEnvForTest(t, "INDEXER_API_URL")
	loadDotEnvIfPresent(envPath)

	got := os.Getenv("INDEXER_API_URL")
	want := "http://quoted:8000"
	if got != want {
		t.Fatalf("INDEXER_API_URL mismatch with export/quotes: got %q, want %q", got, want)
	}
}
