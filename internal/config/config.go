package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	IndexerDbURL    string
	DogelyticsDbURL string
	BindAddr        string
	CorsOrigin      string
	Confirmations   int64
	RateLimit       int
	APIKeyRateLimit int
	SessionSecret   string
	MaxKeysPerUser  int
	EnableUI        bool
	EnableSignups   bool
}

// getEnv returns environment variable value or default
func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// loadDotEnvIfPresent loads simple KEY=VALUE pairs from a .env file.
// Existing environment variables are preserved and not overwritten.
func loadDotEnvIfPresent(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		_ = os.Setenv(key, value)
	}
}

// getEnvInt returns environment variable as int or default
func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}

// getEnvInt64 returns environment variable as int64 or default
func getEnvInt64(key string, defaultValue int64) int64 {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

// getEnvBool returns environment variable as bool or default.
func getEnvBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultValue
}

// ParseConfig parses command-line flags and environment variables, then returns a populated Config.
// Environment variables take precedence over defaults, and command-line flags take precedence over env vars.
func ParseConfig() *Config {
	var config Config

	// Load .env for local development when present.
	// Existing OS environment variables still take precedence.
	loadDotEnvIfPresent(".env")

	// Parse environment variables first (as defaults for flags)
	envIndexerDbURL := getEnv("INDEXER_DBURL", "postgres://indexer:changeme@localhost:5432/indexer?sslmode=disable")
	envDogelyticsDbURL := getEnv("DOGELYTICS_DBURL", "postgres://dogelytics:changeme@localhost:5432/dogelytics?sslmode=disable")
	envBindAddr := getEnv("BIND", "localhost:4420")
	envCorsOrigin := getEnv("CORS", "*")
	envConfirmations := getEnvInt64("CONFIRMATIONS", 6)
	envRateLimit := getEnvInt("RATELIMIT", 10)
	envAPIKeyRateLimit := getEnvInt("API_KEY_RATELIMIT", 120)
	envSessionSecret := getEnv("SESSION_SECRET", "")
	envMaxKeysPerUser := getEnvInt("MAX_KEYS_PER_USER", 1)
	envEnableUI := getEnvBool("ENABLE_UI", true)
	envEnableSignups := getEnvBool("ENABLE_SIGNUPS", true)

	// Define flags with env vars as defaults
	flag.StringVar(&config.IndexerDbURL, "indexer-dburl", envIndexerDbURL, "PostgreSQL database URL for indexer data (env: INDEXER_DBURL)")
	flag.StringVar(&config.DogelyticsDbURL, "dogelytics-dburl", envDogelyticsDbURL, "PostgreSQL database URL for dogelytics data (users, keys, logs) (env: DOGELYTICS_DBURL)")
	flag.StringVar(&config.BindAddr, "bind", envBindAddr, "HTTP server bind address (env: BIND)")
	flag.StringVar(&config.CorsOrigin, "cors", envCorsOrigin, "CORS allowed origin (env: CORS)")
	flag.Int64Var(&config.Confirmations, "confirmations", envConfirmations, "Number of confirmations for available balance (env: CONFIRMATIONS)")
	flag.IntVar(&config.RateLimit, "ratelimit", envRateLimit, "Maximum requests per IP per minute (0 = disabled) (env: RATELIMIT)")
	flag.IntVar(&config.APIKeyRateLimit, "apikey-ratelimit", envAPIKeyRateLimit, "Maximum requests per API key per minute (0 = disabled) (env: API_KEY_RATELIMIT)")
	flag.StringVar(&config.SessionSecret, "session-secret", envSessionSecret, "Session HMAC secret (required for local email/password auth) (env: SESSION_SECRET)")
	flag.IntVar(&config.MaxKeysPerUser, "max-keys-per-user", envMaxKeysPerUser, "Maximum number of API keys per user (env: MAX_KEYS_PER_USER)")
	flag.BoolVar(&config.EnableUI, "enable-ui", envEnableUI, "Enable web UI endpoints (login, register, keys pages) (env: ENABLE_UI)")
	flag.BoolVar(&config.EnableSignups, "enable-signups", envEnableSignups, "Enable user registration through the UI (env: ENABLE_SIGNUPS)")
	flag.Parse()

	return &config
}
