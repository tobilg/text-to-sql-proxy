package config

import (
	"os"
	"strconv"
)

const (
	defaultPort          = 4000
	defaultAllowedOrigin = "https://sql-workbench.com"
	defaultProvider      = "claude"
	defaultDatabase      = "DuckDB"
)

// Config holds the application configuration.
type Config struct {
	Port          int
	AllowedOrigin string
	Provider      string
	Database      string
}

// Load loads configuration from environment variables with sensible defaults.
func Load() Config {
	cfg := Config{
		Port:          defaultPort,
		AllowedOrigin: defaultAllowedOrigin,
		Provider:      defaultProvider,
		Database:      defaultDatabase,
	}

	if portStr := os.Getenv("TEXT_TO_SQL_PROXY_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port < 65536 {
			cfg.Port = port
		}
	}

	if origin := os.Getenv("TEXT_TO_SQL_PROXY_ALLOWED_ORIGIN"); origin != "" {
		cfg.AllowedOrigin = origin
	}

	if provider := os.Getenv("TEXT_TO_SQL_PROXY_PROVIDER"); provider != "" {
		cfg.Provider = provider
	}

	if database := os.Getenv("TEXT_TO_SQL_PROXY_DATABASE"); database != "" {
		cfg.Database = database
	}

	return cfg
}
