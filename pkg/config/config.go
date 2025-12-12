package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	DefaultServerURL       = "https://api.monify.cloud/v1/agent/metrics"
	DefaultCollectionInterval = 30 * time.Second
	DefaultLogLevel        = "info"
	DefaultTimeout         = 10 * time.Second
	TokenFilePath          = "/etc/monify/token"
)

// Config represents the agent configuration
type Config struct {
	Server      ServerConfig      `json:"server"`
	Agent       AgentConfig       `json:"agent"`
	Collection  CollectionConfig  `json:"collection"`
	Metrics     MetricsConfig     `json:"metrics"`
	PortScanner PortScannerConfig `json:"port_scanner"`
	Logging     LoggingConfig     `json:"logging"`
}

// ServerConfig contains server connection settings
type ServerConfig struct {
	URL     string        `json:"url"`
	Token   string        `json:"token"`
	Timeout time.Duration `json:"timeout"`
	TLS     TLSConfig     `json:"tls"`
}

// TLSConfig contains TLS settings
type TLSConfig struct {
	Enabled            bool `json:"enabled"`
	InsecureSkipVerify bool `json:"insecure_skip_verify"`
}

// AgentConfig contains agent identification settings
type AgentConfig struct {
	Version string `json:"version"`
}

// CollectionConfig contains metric collection settings
type CollectionConfig struct {
	Interval time.Duration `json:"interval"`
}

// MetricsConfig controls which metrics to collect
type MetricsConfig struct {
	CPU     bool `json:"cpu"`
	Memory  bool `json:"memory"`
	Disk    bool `json:"disk"`
	Network bool `json:"network"`
	System  bool `json:"system"`
}

// PortScannerConfig contains port scanner settings
type PortScannerConfig struct {
	Enabled    bool          `json:"enabled"`
	Timeout    time.Duration `json:"timeout"`
	MaxWorkers int           `json:"max_workers"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Format string `json:"format"` // json, text
	File   string `json:"file"`   // empty for stdout
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			URL:     DefaultServerURL,
			Token:   "",
			Timeout: DefaultTimeout,
			TLS: TLSConfig{
				Enabled:            true,
				InsecureSkipVerify: false,
			},
		},
		Agent: AgentConfig{
			Version: "1.0.0",
		},
		Collection: CollectionConfig{
			Interval: DefaultCollectionInterval,
		},
		Metrics: MetricsConfig{
			CPU:     true,
			Memory:  true,
			Disk:    true,
			Network: true,
			System:  true,
		},
		PortScanner: PortScannerConfig{
			Enabled:    true,
			Timeout:    5 * time.Second,
			MaxWorkers: 100,
		},
		Logging: LoggingConfig{
			Level:  DefaultLogLevel,
			Format: "text",
			File:   "",
		},
	}
}

// LoadConfig loads configuration with environment variable overrides
// Token is read from /etc/monify/token file
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Read token from file
	token, err := readToken()
	if err != nil {
		// Token file doesn't exist or can't be read - this is OK
		// User needs to run `monify login` first
		config.Server.Token = ""
	} else {
		config.Server.Token = token
	}

	// Apply environment variable overrides
	applyEnvOverrides(config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// readToken reads the authentication token from file
func readToken() (string, error) {
	data, err := os.ReadFile(TokenFilePath)
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("token file is empty")
	}

	return token, nil
}

// SaveToken saves the authentication token to file
func SaveToken(token string) error {
	// Ensure directory exists
	dir := "/etc/monify"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write token to file with restricted permissions
	if err := os.WriteFile(TokenFilePath, []byte(token+"\n"), 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// applyEnvOverrides applies environment variable overrides to the configuration
// Only used for development mode - production uses token file and server config
func applyEnvOverrides(config *Config) {
	// Token override for development mode only
	// Production should use: sudo monify login
	if token := os.Getenv("MONIFY_TOKEN"); token != "" {
		config.Server.Token = token
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.URL == "" {
		return fmt.Errorf("server URL is required")
	}

	if c.Collection.Interval < 1*time.Second {
		return fmt.Errorf("collection interval must be at least 1 second")
	}

	if c.PortScanner.MaxWorkers < 1 {
		return fmt.Errorf("port scanner max workers must be at least 1")
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[strings.ToLower(c.Logging.Level)] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.Logging.Level)
	}

	return nil
}
