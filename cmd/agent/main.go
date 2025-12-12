package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/monify-labs/agent/internal/agent"
	"github.com/monify-labs/agent/pkg/config"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const (
	binaryName = "monify"
)

func main() {
	// No arguments = show help
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	// Parse command
	command := os.Args[1]

	switch command {
	case "start":
		runDaemon()
	case "status":
		handleStatus()
	case "version", "-v", "--version":
		showVersion()
	case "help", "-h", "--help":
		showHelp()
	case "login":
		handleLogin()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		showHelp()
		os.Exit(1)
	}
}

func handleLogin() {
	// Check if running as root
	if os.Geteuid() != 0 {
		fmt.Printf("Error: login command requires root privileges.\n")
		fmt.Printf("Please run: sudo %s login\n", binaryName)
		os.Exit(1)
	}

	var token string

	// Check if token provided as argument
	if len(os.Args) >= 3 {
		token = strings.TrimSpace(os.Args[2])
	} else {
		// Prompt for token
		fmt.Println("Enter your Monify agent token:")
		fmt.Print("> ")
		
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}
		token = strings.TrimSpace(input)
	}

	// Validate token
	if token == "" {
		fmt.Println("Error: Token cannot be empty")
		os.Exit(1)
	}

	if len(token) < 10 {
		fmt.Println("Error: Token appears to be invalid (too short)")
		os.Exit(1)
	}

	// Save token
	if err := config.SaveToken(token); err != nil {
		fmt.Printf("Error saving token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Token saved successfully to %s\n", config.TokenFilePath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Start agent:  sudo systemctl start %s\n", binaryName)
	fmt.Printf("  2. Enable auto-start: sudo systemctl enable %s\n", binaryName)
	fmt.Printf("  3. Check status: %s status\n", binaryName)
}

func showVersion() {
	fmt.Printf("%s version %s\n", binaryName, version)
	fmt.Printf("Commit: %s\n", commit)
	fmt.Printf("Built: %s\n", date)
	fmt.Printf("Platform: Linux\n")
}

func showHelp() {
	fmt.Printf(`Monify Agent - Linux Server Monitoring

Usage:
  %s <command>

Commands:
  start          Start the monitoring agent (daemon mode)
  login          Configure authentication token
  status         Show agent and server status
  version        Show version information
  help           Show this help message

Examples:
  # Authenticate (interactive)
  sudo %s login

  # Authenticate (with token)
  sudo %s login your_token_here

  # Start agent (usually via systemd)
  sudo systemctl start monify

  # Check status
  %s status

  # View logs
  sudo journalctl -u monify -f

Documentation:
  https://github.com/monify-labs/agent

`, binaryName, binaryName, binaryName, binaryName)
}

func handleStatus() {
	// 1. Try to read status file first
	statusFile := "/var/log/monify/status.json"
	var agentStatus struct {
		Hostname       string    `json:"hostname"`
		Version        string    `json:"version"`
		Uptime         uint64    `json:"uptime"`
		LastCollection time.Time `json:"last_collection"`
		LastSend       time.Time `json:"last_send"`
		MetricsCount   uint64    `json:"metrics_count"`
		ErrorCount     uint64    `json:"error_count"`
		Status         string    `json:"status"`
	}

	statusBytes, err := os.ReadFile(statusFile)
	isRunning := err == nil

	if isRunning {
		if err := json.Unmarshal(statusBytes, &agentStatus); err != nil {
			isRunning = false // Corrupt file
		} else {
			// Check if file is stale (older than 2 minutes)
			info, _ := os.Stat(statusFile)
			if time.Since(info.ModTime()) > 2*time.Minute {
				agentStatus.Status = "stale (not running?)"
			}
		}
	}

	// 2. Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Check if authenticated
	authenticated := cfg.Server.Token != ""

	// Display status
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Monify Agent Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if isRunning && agentStatus.Hostname != "" {
		fmt.Printf("Status:          %s\n", agentStatus.Status)
		fmt.Printf("Version:         %s\n", agentStatus.Version)
		fmt.Printf("Hostname:        %s\n", agentStatus.Hostname)
		fmt.Printf("Uptime:          %ds\n", agentStatus.Uptime)
		fmt.Println()
		fmt.Printf("Last Collection: %s\n", agentStatus.LastCollection.Format(time.RFC3339))
		fmt.Printf("Last Send:       %s\n", agentStatus.LastSend.Format(time.RFC3339))
		fmt.Printf("Metrics Sent:    %d\n", agentStatus.MetricsCount)
		fmt.Printf("Errors:          %d\n", agentStatus.ErrorCount)
	} else {
		fmt.Printf("Status:          STOPPED\n")
		fmt.Printf("Version:         %s\n", version)
		hostname, _ := os.Hostname()
		fmt.Printf("Hostname:        %s\n", hostname)
	}

	fmt.Println()
	fmt.Printf("Server URL:      %s\n", cfg.Server.URL)
	fmt.Printf("Authenticated:   %v\n", authenticated)

	if authenticated {
		// Mask token
		token := cfg.Server.Token
		if len(token) > 8 {
			token = token[:4] + "..." + token[len(token)-4:]
		}
		fmt.Printf("Token:           %s\n", token)
	}
	fmt.Println()

	if !authenticated {
		fmt.Println("⚠️  Not authenticated!")
		fmt.Println()
		fmt.Println("To authenticate:")
		fmt.Printf("  sudo %s login\n", binaryName)
		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func runDaemon() {
	// Check if running as root
	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "Error: Agent must run as root\n")
		fmt.Fprintf(os.Stderr, "Please run: sudo systemctl start monify\n")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Check if authenticated
	if cfg.Server.Token == "" {
		fmt.Fprintf(os.Stderr, "Error: Not authenticated\n\n")
		fmt.Fprintf(os.Stderr, "Please set your token:\n")
		fmt.Fprintf(os.Stderr, "  sudo %s login\n\n", binaryName)
		fmt.Fprintf(os.Stderr, "Then start: sudo systemctl start monify\n")
		os.Exit(1)
	}

	// Set version from build info
	cfg.Agent.Version = version

	// Create agent
	ag, err := agent.NewAgent(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create agent: %v\n", err)
		os.Exit(1)
	}

	// Start agent
	fmt.Printf("Starting Monify Agent v%s...\n", version)
	fmt.Printf("Server: %s\n", cfg.Server.URL)
	fmt.Printf("Interval: %s\n", cfg.Collection.Interval)

	ctx := context.Background()
	if err := ag.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Agent error: %v\n", err)
		os.Exit(1)
	}
}
