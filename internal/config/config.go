package config

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	IndexerAPIURL        string
	DogelyticsDbURL      string
	BindAddr             string
	CorsOrigin           string
	PublicURL            string
	TrustedProxies       string
	EnableAnalytics      bool
	AnalyticsSecret      string
	AnalyticsRetention   int
	RateLimit            int
	APIKeyRateLimit      int
	SessionSecret        string
	MaxKeysPerUser       int
	EnableAdminUI        bool
	AdminUIPort          int
	EnableDashboardUI    bool
	DashboardUIPort      int
	EnableDashboardTips  bool
	DashboardTipsAddress string
	EnableSignups        bool
}

// loadDotEnvIfPresent loads simple KEY=VALUE pairs from a .env file.
// Existing environment variables are preserved and not overwritten.
func loadDotEnvIfPresent(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, rawLine := range strings.Split(string(data), "\n") {
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

		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		_ = os.Setenv(key, value)
	}
}

// Load parses and validates configuration using the provided command arguments.
func Load(args []string) (*Config, error) {
	loadDotEnvIfPresent(".env")

	databaseURL, err := envOrFile("DOGELYTICS_DBURL")
	if err != nil {
		return nil, err
	}
	sessionSecret, err := envOrFile("SESSION_SECRET")
	if err != nil {
		return nil, err
	}
	analyticsSecret, err := envOrFile("ANALYTICS_SECRET")
	if err != nil {
		return nil, err
	}
	rateLimit, err := envInt("RATELIMIT", 10)
	if err != nil {
		return nil, err
	}
	apiKeyRateLimit, err := envInt("API_KEY_RATELIMIT", 120)
	if err != nil {
		return nil, err
	}
	maxKeysPerUser, err := envInt("MAX_KEYS_PER_USER", 1)
	if err != nil {
		return nil, err
	}
	adminUIPort, err := envInt("ADMIN_UI_PORT", 4421)
	if err != nil {
		return nil, err
	}
	dashboardUIPort, err := envInt("DASHBOARD_UI_PORT", 4422)
	if err != nil {
		return nil, err
	}
	enableAdminUI, err := envBool("ENABLE_ADMIN_UI", false)
	if err != nil {
		return nil, err
	}
	enableDashboardUI, err := envBool("ENABLE_DASHBOARD_UI", false)
	if err != nil {
		return nil, err
	}
	enableDashboardTips, err := envBool("ENABLE_DASHBOARD_TIPS", true)
	if err != nil {
		return nil, err
	}
	enableSignups, err := envBool("ENABLE_SIGNUPS", false)
	if err != nil {
		return nil, err
	}
	enableAnalytics, err := envBool("ENABLE_ANALYTICS", true)
	if err != nil {
		return nil, err
	}
	analyticsRetention, err := envInt("ANALYTICS_RETENTION_DAYS", 30)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	flags := flag.NewFlagSet("dogelytics", flag.ContinueOnError)
	flags.StringVar(&cfg.IndexerAPIURL, "indexer-api-url", env("INDEXER_API_URL", "http://localhost:8000"), "Indexer API base URL (env: INDEXER_API_URL)")
	flags.StringVar(&cfg.DogelyticsDbURL, "dogelytics-dburl", databaseURL, "PostgreSQL database URL (env: DOGELYTICS_DBURL or DOGELYTICS_DBURL_FILE)")
	flags.StringVar(&cfg.BindAddr, "bind", env("BIND", "localhost:4420"), "HTTP server bind address (env: BIND)")
	flags.StringVar(&cfg.CorsOrigin, "cors", env("CORS", "*"), "CORS allowed origin (env: CORS)")
	flags.StringVar(&cfg.PublicURL, "public-url", env("PUBLIC_URL", ""), "Public HTTPS origin used for cookies and CSRF checks (env: PUBLIC_URL)")
	flags.StringVar(&cfg.TrustedProxies, "trusted-proxies", env("TRUSTED_PROXIES", ""), "Comma-separated trusted proxy IPs or CIDRs (env: TRUSTED_PROXIES)")
	flags.BoolVar(&cfg.EnableAnalytics, "enable-analytics", enableAnalytics, "Enable privacy-preserving request analytics (env: ENABLE_ANALYTICS)")
	flags.StringVar(&cfg.AnalyticsSecret, "analytics-secret", analyticsSecret, "Analytics HMAC secret (env: ANALYTICS_SECRET or ANALYTICS_SECRET_FILE)")
	flags.IntVar(&cfg.AnalyticsRetention, "analytics-retention-days", analyticsRetention, "Recent analytics retention in days (env: ANALYTICS_RETENTION_DAYS)")
	flags.IntVar(&cfg.RateLimit, "ratelimit", rateLimit, "Maximum requests per IP per minute (env: RATELIMIT)")
	flags.IntVar(&cfg.APIKeyRateLimit, "apikey-ratelimit", apiKeyRateLimit, "Maximum requests per API key per minute (env: API_KEY_RATELIMIT)")
	flags.StringVar(&cfg.SessionSecret, "session-secret", sessionSecret, "Session HMAC secret (env: SESSION_SECRET or SESSION_SECRET_FILE)")
	flags.IntVar(&cfg.MaxKeysPerUser, "max-keys-per-user", maxKeysPerUser, "Maximum active API keys per user (env: MAX_KEYS_PER_USER)")
	flags.BoolVar(&cfg.EnableAdminUI, "enable-admin-ui", enableAdminUI, "Enable the admin UI (env: ENABLE_ADMIN_UI)")
	flags.IntVar(&cfg.AdminUIPort, "admin-ui-port", adminUIPort, "Admin UI port (env: ADMIN_UI_PORT)")
	flags.BoolVar(&cfg.EnableDashboardUI, "enable-dashboard-ui", enableDashboardUI, "Enable the dashboard UI (env: ENABLE_DASHBOARD_UI)")
	flags.IntVar(&cfg.DashboardUIPort, "dashboard-ui-port", dashboardUIPort, "Dashboard UI port (env: DASHBOARD_UI_PORT)")
	flags.BoolVar(&cfg.EnableDashboardTips, "enable-dashboard-tips", enableDashboardTips, "Enable dashboard tips (env: ENABLE_DASHBOARD_TIPS)")
	flags.StringVar(&cfg.DashboardTipsAddress, "dashboard-tips-address", env("DASHBOARD_TIPS_ADDRESS", "DChPB3HbQgNYgWRrpeRKqNT6939rRLceNz"), "Dashboard tips address (env: DASHBOARD_TIPS_ADDRESS)")
	flags.BoolVar(&cfg.EnableSignups, "enable-signups", enableSignups, "Enable user registration (env: ENABLE_SIGNUPS)")
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	if flags.NArg() != 0 {
		return nil, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ParseConfig loads configuration for the current process.
func ParseConfig() (*Config, error) {
	return Load(os.Args[1:])
}

// Validate rejects ambiguous or unsafe runtime configuration.
func (cfg *Config) Validate() error {
	if cfg == nil {
		return errors.New("configuration is required")
	}
	if err := validateHTTPURL("INDEXER_API_URL", cfg.IndexerAPIURL); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.DogelyticsDbURL) == "" {
		return errors.New("DOGELYTICS_DBURL or DOGELYTICS_DBURL_FILE is required")
	}
	if err := validatePostgresURL(cfg.DogelyticsDbURL); err != nil {
		return err
	}
	_, apiPort, err := net.SplitHostPort(cfg.BindAddr)
	if err != nil {
		return fmt.Errorf("invalid BIND: %w", err)
	}
	parsedAPIPort, err := strconv.Atoi(apiPort)
	if err != nil || parsedAPIPort < 1 || parsedAPIPort > 65535 {
		return errors.New("BIND must contain a valid port")
	}
	if cfg.RateLimit < 0 || cfg.APIKeyRateLimit < 0 {
		return errors.New("rate limits cannot be negative")
	}
	if cfg.PublicURL != "" {
		parsed, err := url.ParseRequestURI(cfg.PublicURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Path != "" {
			return errors.New("PUBLIC_URL must be an HTTP or HTTPS origin without a path")
		}
	}
	if err := validateTrustedProxies(cfg.TrustedProxies); err != nil {
		return err
	}
	if cfg.MaxKeysPerUser < 1 {
		return errors.New("MAX_KEYS_PER_USER must be at least 1")
	}
	if cfg.EnableAnalytics {
		if len(cfg.AnalyticsSecret) < 32 {
			return errors.New("ANALYTICS_SECRET must be at least 32 characters when ENABLE_ANALYTICS=true")
		}
		if cfg.AnalyticsRetention < 1 {
			return errors.New("ANALYTICS_RETENTION_DAYS must be at least 1")
		}
	}
	if err := validatePort("ADMIN_UI_PORT", cfg.AdminUIPort); err != nil {
		return err
	}
	if err := validatePort("DASHBOARD_UI_PORT", cfg.DashboardUIPort); err != nil {
		return err
	}
	if cfg.EnableAdminUI {
		if len(cfg.SessionSecret) < 32 {
			return errors.New("SESSION_SECRET must be at least 32 characters when ENABLE_ADMIN_UI=true")
		}
		if cfg.AdminUIPort == parsedAPIPort {
			return errors.New("ADMIN_UI_PORT must differ from the API port")
		}
	}
	if cfg.EnableDashboardUI {
		if cfg.DashboardUIPort == parsedAPIPort {
			return errors.New("DASHBOARD_UI_PORT must differ from the API port")
		}
		if cfg.EnableAdminUI && cfg.DashboardUIPort == cfg.AdminUIPort {
			return errors.New("DASHBOARD_UI_PORT must differ from ADMIN_UI_PORT")
		}
	}
	return nil
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func envOrFile(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value, nil
	}
	path := strings.TrimSpace(os.Getenv(key + "_FILE"))
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s_FILE: %w", key, err)
	}
	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", fmt.Errorf("%s_FILE is empty", key)
	}
	return value, nil
}

func envInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}
	return parsed, nil
}

func envBool(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", key)
	}
	return parsed, nil
}

func validateHTTPURL(name, value string) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("%s must be an absolute HTTP or HTTPS URL", name)
	}
	return nil
}

func validatePostgresURL(value string) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Path == "" {
		return errors.New("DOGELYTICS_DBURL must be a PostgreSQL URL with a database name")
	}
	return nil
}

func validatePort(name string, port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", name)
	}
	return nil
}

func validateTrustedProxies(value string) error {
	for _, entry := range strings.Split(value, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			if _, _, err := net.ParseCIDR(entry); err != nil {
				return fmt.Errorf("TRUSTED_PROXIES contains invalid CIDR %q", entry)
			}
			continue
		}
		if net.ParseIP(entry) == nil {
			return fmt.Errorf("TRUSTED_PROXIES contains invalid IP %q", entry)
		}
	}
	return nil
}
