package spec

import "flag"

// Config holds all configuration options for the dogelytics service
type Config struct {
	DbURL           string
	BindAddr        string
	CorsOrigin      string
	Confirmations   int64
	RateLimit       int    // Maximum requests per IP per minute (0 = disabled)
	APIKeyRateLimit int    // Maximum requests per API key per minute (0 = disabled)
	SessionSecret   string // HMAC secret for signing sessions (required for local auth)
	MaxKeysPerUser  int    // Max API keys per user
}

// ParseConfig parses command-line flags and returns a populated Config
func ParseConfig() *Config {
	var config Config
	flag.StringVar(&config.DbURL, "dburl", "postgres://indexer:changeme@localhost:5432/indexer?sslmode=disable", "PostgreSQL database URL (required - SQLite not supported due to concurrency issues)")
	flag.StringVar(&config.BindAddr, "bind", "localhost:4420", "HTTP server bind address")
	flag.StringVar(&config.CorsOrigin, "cors", "*", "CORS allowed origin")
	flag.Int64Var(&config.Confirmations, "confirmations", 6, "Number of confirmations for available balance")
	flag.IntVar(&config.RateLimit, "ratelimit", 10, "Maximum requests per IP per minute (0 = disabled)")
	flag.IntVar(&config.APIKeyRateLimit, "apikey-ratelimit", 120, "Maximum requests per API key per minute (0 = disabled)")
	flag.StringVar(&config.SessionSecret, "session-secret", "", "Session HMAC secret (required for local email/password auth)")
	flag.IntVar(&config.MaxKeysPerUser, "max-keys-per-user", 1, "Maximum number of API keys per user")
	flag.Parse()
	return &config
}
