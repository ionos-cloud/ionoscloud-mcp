// Package config provides configuration management for the IONOS Cloud MCP server.
package config

import (
	"os"
	"strings"
)

// Config holds the server configuration.
type Config struct {
	// EnableLogging enables request/response logging to stderr
	EnableLogging bool
	// EnableMetrics enables basic timing metrics
	EnableMetrics bool
	// ReadOnly if true, only exposes read-only tools
	ReadOnly bool
	// EnabledToolsets limits which toolsets are enabled (empty = all)
	EnabledToolsets []string
}

// LoadFromEnv loads configuration from environment variables.
func LoadFromEnv() *Config {
	cfg := &Config{}

	// IONOS_MCP_LOGGING enables logging
	if v := os.Getenv("IONOS_MCP_LOGGING"); v == "true" || v == "1" {
		cfg.EnableLogging = true
	}

	// IONOS_MCP_METRICS enables metrics
	if v := os.Getenv("IONOS_MCP_METRICS"); v == "true" || v == "1" {
		cfg.EnableMetrics = true
	}

	// IONOS_MCP_READ_ONLY enables read-only mode
	if v := os.Getenv("IONOS_MCP_READ_ONLY"); v == "true" || v == "1" {
		cfg.ReadOnly = true
	}

	// IONOS_MCP_TOOLSETS limits enabled toolsets (comma-separated)
	if v := os.Getenv("IONOS_MCP_TOOLSETS"); v != "" {
		cfg.EnabledToolsets = strings.Split(v, ",")
		// Trim whitespace
		for i := range cfg.EnabledToolsets {
			cfg.EnabledToolsets[i] = strings.TrimSpace(cfg.EnabledToolsets[i])
		}
	}

	return cfg
}

// Default returns a default configuration.
func Default() *Config {
	return &Config{
		EnableLogging:   false,
		EnableMetrics:   false,
		ReadOnly:        false,
		EnabledToolsets: nil, // All toolsets enabled
	}
}
