package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	// Server settings
	ServerURL = "https://api.monify.cloud/v1/agent/metrics"
	Timeout   = 10 * time.Second

	// Collection settings
	CollectionInterval    = 15 * time.Second
	StaticRefreshInterval = 1 * time.Hour

	// Agent info (injected at build time via ldflags)
	Version   = "1.1.1"
	Commit    = "unknown"
	BuildDate = "unknown"

	// Environment file path
	EnvFilePath = "/etc/monify/env"
)

// LoadEnvFile loads environment variables from /etc/monify/env
func LoadEnvFile() error {
	data, err := os.ReadFile(EnvFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist is not an error
		}
		return err
	}

	// Parse each line as KEY=VALUE
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Only set if not already set in environment
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}

	return nil
}

// SaveEnvFile saves environment variables to /etc/monify/env
func SaveEnvFile(vars map[string]string) error {
	// Read existing file
	existing := make(map[string]string)
	if data, err := os.ReadFile(EnvFilePath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				existing[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Merge with new vars
	for k, v := range vars {
		existing[k] = v
	}

	// Create directory
	dir := "/etc/monify"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	var content strings.Builder
	for k, v := range existing {
		content.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	if err := os.WriteFile(EnvFilePath, []byte(content.String()), 0600); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	return nil
}

// GetServerURL returns server URL from env or default
func GetServerURL() string {
	if url := os.Getenv("MONIFY_SERVER_URL"); url != "" {
		return url
	}
	return ServerURL
}

// GetToken returns token from environment variable
func GetToken() (string, error) {
	token := os.Getenv("MONIFY_TOKEN")
	if token == "" {
		return "", fmt.Errorf("MONIFY_TOKEN environment variable not set")
	}
	return token, nil
}

// IsDebugMode checks if debug mode is enabled
func IsDebugMode() bool {
	debug := os.Getenv("MONIFY_DEBUG")
	return debug == "true" || debug == "1"
}
