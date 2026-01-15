package config

import (
	"os"
	"strconv"
)

const (
	defaultPort          = 4000
	defaultAllowedOrigin = "https://sql-workbench.com"
	defaultProvider      = "claude"
)

// Config holds the application configuration.
type Config struct {
	Port          int
	AllowedOrigin string
	Provider      string
}

// Load loads configuration from environment variables with sensible defaults.
func Load() Config {
	cfg := Config{
		Port:          defaultPort,
		AllowedOrigin: defaultAllowedOrigin,
		Provider:      defaultProvider,
	}

	if portStr := os.Getenv("AI_CLI_PROXY_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port < 65536 {
			cfg.Port = port
		}
	}

	if origin := os.Getenv("AI_CLI_PROXY_ALLOWED_ORIGIN"); origin != "" {
		cfg.AllowedOrigin = origin
	}

	if provider := os.Getenv("AI_CLI_PROXY_PROVIDER"); provider != "" {
		cfg.Provider = provider
	}

	return cfg
}
