package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the agent configuration
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Agent       AgentConfig       `yaml:"agent"`
	Collection  CollectionConfig  `yaml:"collection"`
	Metrics     MetricsConfig     `yaml:"metrics"`
	PortScanner PortScannerConfig `yaml:"port_scanner"`
	Logging     LoggingConfig     `yaml:"logging"`
}

// ServerConfig contains server connection settings
type ServerConfig struct {
	URL     string        `yaml:"url"`
	Token   string        `yaml:"token"`
	Timeout time.Duration `yaml:"timeout"`
	TLS     TLSConfig     `yaml:"tls"`
}

// TLSConfig contains TLS settings
type TLSConfig struct {
	Enabled            bool   `yaml:"enabled"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	CertFile           string `yaml:"cert_file"`
	KeyFile            string `yaml:"key_file"`
	CAFile             string `yaml:"ca_file"`
}

// AgentConfig contains agent identification settings
type AgentConfig struct {
	Version string `yaml:"version"`
}

// CollectionConfig contains metric collection settings
type CollectionConfig struct {
	Interval time.Duration `yaml:"interval"`
}

// MetricsConfig controls which metrics to collect
type MetricsConfig struct {
	CPU     bool `yaml:"cpu"`
	Memory  bool `yaml:"memory"`
	Disk    bool `yaml:"disk"`
	Network bool `yaml:"network"`
	System  bool `yaml:"system"`
}

// PortScannerConfig contains port scanner settings
type PortScannerConfig struct {
	Enabled    bool          `yaml:"enabled"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxWorkers int           `yaml:"max_workers"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, text
	File   string `yaml:"file"`   // empty for stdout
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	return &Config{
		Server: ServerConfig{
			URL:     "https://api.monify.cloud/v1/metrics",
			Token:   "",
			Timeout: 10 * time.Second,
			TLS: TLSConfig{
				Enabled:            true,
				InsecureSkipVerify: false,
			},
		},
		Agent: AgentConfig{
			Version: "1.0.0",
		},
		Collection: CollectionConfig{
			Interval: 30 * time.Second,
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
			Level:  "info",
			Format: "text",
			File:   "",
		},
	}
}

// LoadConfig loads configuration from a YAML file with environment variable overrides
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	// Read the config file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, use defaults
			applyEnvOverrides(config)
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables in the config file
	expanded := os.ExpandEnv(string(data))

	// Parse YAML
	if err := yaml.Unmarshal([]byte(expanded), config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// applyEnvOverrides applies environment variable overrides to the configuration
func applyEnvOverrides(config *Config) {
	if url := os.Getenv("MONIFY_SERVER_URL"); url != "" {
		config.Server.URL = url
	}
	if apiKey := os.Getenv("TOKEN_DEVICE"); apiKey != "" {
		config.Server.Token = apiKey
	}
	if level := os.Getenv("MONIFY_LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}
	if interval := os.Getenv("MONIFY_COLLECTION_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			config.Collection.Interval = d
		}
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

// SaveConfig saves the configuration to a YAML file
func (c *Config) SaveConfig(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
