package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	AnthropicKey          string
	ElephantasmURL        string
	ElephantasmKey        string
	SupabaseURL           string
	ClassifierMode        string
	LumberModelDir        string
	ResendAPIKey          string
	NotificationFromEmail string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		AnthropicKey:   getEnv("ANTHROPIC_API_KEY", ""),
		ElephantasmURL: getEnv("ELEPHANTASM_URL", ""),
		ElephantasmKey: getEnv("ELEPHANTASM_API_KEY", ""),
		SupabaseURL:    getEnv("SUPABASE_URL", ""),
		ClassifierMode:        getEnv("CLASSIFIER_MODE", "fallback"),
		LumberModelDir:        getEnv("LUMBER_MODEL_DIR", "/opt/lumber/models"),
		ResendAPIKey:          getEnv("RESEND_API_KEY", ""),
		NotificationFromEmail: getEnv("NOTIFICATION_FROM_EMAIL", ""),
	}
}

func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.AnthropicKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is required")
	}
	if c.SupabaseURL == "" {
		return fmt.Errorf("SUPABASE_URL is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
