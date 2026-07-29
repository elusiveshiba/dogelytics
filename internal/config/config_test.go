package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadValidatesRequiredConfiguration(t *testing.T) {
	for _, key := range []string{
		"DOGELYTICS_DBURL", "DOGELYTICS_DBURL_FILE", "SESSION_SECRET", "SESSION_SECRET_FILE",
		"ENABLE_ADMIN_UI", "ENABLE_DASHBOARD_UI", "RATELIMIT", "API_KEY_RATELIMIT", "ANALYTICS_SECRET_FILE",
	} {
		t.Setenv(key, "")
	}

	if _, err := Load(nil); err == nil || !strings.Contains(err.Error(), "DOGELYTICS_DBURL") {
		t.Fatalf("expected required database error, got %v", err)
	}

	t.Setenv("DOGELYTICS_DBURL", "postgres://dogelytics:secret@localhost:5432/dogelytics?sslmode=disable")
	t.Setenv("ANALYTICS_SECRET", strings.Repeat("a", 32))
	cfg, err := Load([]string{"-bind", "127.0.0.1:5000", "-ratelimit", "25"})
	if err != nil {
		t.Fatalf("load valid configuration: %v", err)
	}
	if cfg.BindAddr != "127.0.0.1:5000" || cfg.RateLimit != 25 {
		t.Fatalf("unexpected configuration: %+v", cfg)
	}
}

func TestLoadReadsSecretFilesAndRejectsInvalidEnvironment(t *testing.T) {
	directory := t.TempDir()
	databaseFile := filepath.Join(directory, "database-url")
	sessionFile := filepath.Join(directory, "session-secret")
	if err := os.WriteFile(databaseFile, []byte("postgres://dogelytics:secret@localhost:5432/dogelytics?sslmode=disable\n"), 0o600); err != nil {
		t.Fatalf("write database secret: %v", err)
	}
	if err := os.WriteFile(sessionFile, []byte(strings.Repeat("s", 32)+"\n"), 0o600); err != nil {
		t.Fatalf("write session secret: %v", err)
	}

	t.Setenv("DOGELYTICS_DBURL", "")
	t.Setenv("DOGELYTICS_DBURL_FILE", databaseFile)
	t.Setenv("SESSION_SECRET", "")
	t.Setenv("SESSION_SECRET_FILE", sessionFile)
	t.Setenv("ANALYTICS_SECRET", strings.Repeat("a", 32))
	t.Setenv("ENABLE_ADMIN_UI", "true")
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("load file-backed secrets: %v", err)
	}
	if len(cfg.SessionSecret) != 32 {
		t.Fatalf("unexpected session secret length: %d", len(cfg.SessionSecret))
	}

	t.Setenv("RATELIMIT", "many")
	if _, err := Load(nil); err == nil || !strings.Contains(err.Error(), "RATELIMIT") {
		t.Fatalf("expected invalid rate-limit error, got %v", err)
	}
}

func TestConfigValidateRejectsListenerConflicts(t *testing.T) {
	cfg := &Config{
		IndexerAPIURL:     "http://indexer:8000",
		DogelyticsDbURL:   "postgres://dogelytics:secret@database:5432/dogelytics",
		BindAddr:          "0.0.0.0:4420",
		MaxKeysPerUser:    1,
		AdminUIPort:       4420,
		DashboardUIPort:   4422,
		EnableAdminUI:     true,
		SessionSecret:     strings.Repeat("s", 32),
		EnableDashboardUI: true,
	}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "ADMIN_UI_PORT") {
		t.Fatalf("expected listener conflict, got %v", err)
	}
}
