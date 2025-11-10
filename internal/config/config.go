package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds application configuration
type Config struct {
	// Data directory for storing app info and updates
	DataDir string

	// Apps to track (bundle IDs)
	Apps []string

	// Check interval for polling
	CheckInterval time.Duration

	// Log level (debug, info, warn, error)
	LogLevel string

	// Server settings (for future HTTP API)
	ServerPort int
	ServerHost string

	// Apprise notification URL
	AppriseURL string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		DataDir:       getEnv("MAVT_DATA_DIR", "./data"),
		CheckInterval: parseDuration(getEnv("MAVT_CHECK_INTERVAL", "1h"), 1*time.Hour),
		LogLevel:      getEnv("MAVT_LOG_LEVEL", "info"),
		ServerPort:    parseInt(getEnv("MAVT_SERVER_PORT", "8080"), 8080),
		ServerHost:    getEnv("MAVT_SERVER_HOST", "0.0.0.0"),
		AppriseURL:    getEnv("MAVT_APPRISE_URL", ""),
	}

	// Parse apps list from environment
	appsEnv := getEnv("MAVT_APPS", "")
	if appsEnv != "" {
		config.Apps = parseAppsList(appsEnv)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.DataDir == "" {
		return fmt.Errorf("data directory cannot be empty")
	}

	if c.CheckInterval < 1*time.Minute {
		return fmt.Errorf("check interval must be at least 1 minute")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[strings.ToLower(c.LogLevel)] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseInt parses an integer from a string, returning default on error
func parseInt(s string, defaultValue int) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultValue
}

// parseDuration parses a duration string, returning default on error
func parseDuration(s string, defaultValue time.Duration) time.Duration {
	if dur, err := time.ParseDuration(s); err == nil {
		return dur
	}
	return defaultValue
}

// parseAppsList parses a comma-separated list of app bundle IDs
func parseAppsList(s string) []string {
	parts := strings.Split(s, ",")
	var apps []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			apps = append(apps, trimmed)
		}
	}
	return apps
}
