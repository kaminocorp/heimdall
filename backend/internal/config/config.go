package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port      string
	LogFormat string // "text" (default) or "json"
	// DatabaseURL is the runtime app-pool URL. Today resolves to the
	// `postgres` superuser; Phase 6 of the RLS role-split flips it to
	// `app_user` (per-tenant CRUD with RLS enforced).
	DatabaseURL string
	// CronDatabaseURL is the runtime cron-pool URL used for cross-tenant
	// enumeration paths (monitor's app list, scheduler's enabled list,
	// the log_buffer pruner). Today resolves to the same role as
	// DatabaseURL; Phase 6 flips it to `cron_user` (BYPASSRLS, narrow
	// grants). Empty → fall back to DatabaseURL with a WARN at startup.
	CronDatabaseURL string
	// DirectURL is the superuser URL used by `make migrate-up` /
	// `make migrate-down` and any other path that needs DDL or
	// privilege-altering SQL. Phase 6's flip will downgrade
	// DatabaseURL to a non-superuser, but DirectURL stays on
	// `postgres`. Empty → fall back to DatabaseURL.
	DirectURL string
	// Environment gates production-only invariants. Currently only
	// guards the role-split startup check (refuse to launch when
	// DatabaseURL and CronDatabaseURL resolve to the same role).
	// Unset → development mode; "production" → enforce the invariant.
	Environment           string
	AnthropicKey          string
	OpenRouterKey         string
	ElephantasmURL        string
	ElephantasmKey        string
	SupabaseURL           string
	ClassifierMode        string
	LumberModelDir        string
	ResendAPIKey          string
	NotificationFromEmail string
	GitHubAppID           string
	GitHubPrivateKey      string
	GitHubClientID        string
	GitHubAppSlug         string
	GitHubWebhookSecret   string
	FrontendURL           string
	SyslogTLSCert         string
	SyslogTLSKey          string
	DisableBackgroundJobs bool
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		LogFormat:       getEnv("LOG_FORMAT", "text"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		CronDatabaseURL: getEnv("CRON_DATABASE_URL", ""),
		DirectURL:       getEnv("DIRECT_URL", ""),
		Environment:     getEnv("HEIMDALL_ENV", ""),
		AnthropicKey:    getEnv("ANTHROPIC_API_KEY", ""),
		OpenRouterKey:  getEnv("OPENROUTER_API_KEY", ""),
		ElephantasmURL: getEnv("ELEPHANTASM_URL", ""),
		ElephantasmKey: getEnv("ELEPHANTASM_API_KEY", ""),
		SupabaseURL:    getEnv("SUPABASE_URL", ""),
		ClassifierMode:        getEnv("CLASSIFIER_MODE", "fallback"),
		LumberModelDir:        getEnv("LUMBER_MODEL_DIR", "/opt/lumber/models"),
		ResendAPIKey:          getEnv("RESEND_API_KEY", ""),
		NotificationFromEmail: getEnv("NOTIFICATION_FROM_EMAIL", ""),
		GitHubAppID:           getEnv("GITHUB_APP_ID", ""),
		GitHubPrivateKey:      getEnv("GITHUB_PRIVATE_KEY", ""),
		GitHubClientID:        getEnv("GITHUB_CLIENT_ID", ""),
		GitHubAppSlug:         getEnv("GITHUB_APP_SLUG", "heimdall-agent"),
		GitHubWebhookSecret:   getEnv("GITHUB_WEBHOOK_SECRET", ""),
		FrontendURL:           getEnv("FRONTEND_URL", "http://localhost:5173"),
		SyslogTLSCert:         getEnv("SYSLOG_TLS_CERT", ""),
		SyslogTLSKey:          getEnv("SYSLOG_TLS_KEY", ""),
		DisableBackgroundJobs: getEnv("DISABLE_BACKGROUND_JOBS", "") != "",
	}
}

func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.AnthropicKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is required")
	}
	// Note: SupabaseURL is loaded but not validated as required. It is
	// consumed by the Supabase log connector (when configured) but is not
	// needed for core server operation.
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
