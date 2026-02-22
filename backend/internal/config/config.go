package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port              string
	DatabaseURL       string
	AnthropicKey      string
	ElephantasmURL    string
	ElephantasmKey    string
	SupabaseJWTSecret string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		AnthropicKey:   getEnv("ANTHROPIC_API_KEY", ""),
		ElephantasmURL:    getEnv("ELEPHANTASM_URL", ""),
		ElephantasmKey:    getEnv("ELEPHANTASM_API_KEY", ""),
		SupabaseJWTSecret: getEnv("SUPABASE_JWT_SECRET", ""),
	}
}

func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.AnthropicKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is required")
	}
	if c.SupabaseJWTSecret == "" {
		return fmt.Errorf("SUPABASE_JWT_SECRET is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
