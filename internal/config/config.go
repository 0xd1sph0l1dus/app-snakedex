// Package config loads server configuration from environment variables,
// falling back to sensible defaults for local development.
package config

import "os"

type Config struct {
	Port   string // PORT — default "8080"
	DBPath string // DB_PATH — default "snakedex.db"
}

func Load() Config {
	return Config{
		Port:   getEnv("PORT", "8080"),
		DBPath: getEnv("DB_PATH", "snakedex.db"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
