package main

import (
	"context"
	"fmt"
	"os"

	"github.com/monify-labs/agent/internal/agent"
	"github.com/monify-labs/agent/pkg/config"
	"github.com/monify-labs/agent/pkg/lock"
)

var (
	version = "0.2.2"
	commit  = "dev"
	date    = "unknown"
)

const (
	binaryName        = "monify"
	defaultConfigPath = "/etc/monify/config.yaml"
	defaultEnvPath    = "/etc/monify/.env"
	lockFilePath      = "/var/run/monify.lock"
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
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		showHelp()
		os.Exit(1)
	}
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
  start      Start the monitoring agent (daemon mode)
  status     Show agent and server status
  version    Show version information
  help       Show this help message

Configuration:
  Config file: %s
  Env file:    %s
  
  Set your token in .env file:
    TOKEN_DEVICE=your_token_here

Examples:
  # Start agent (usually via systemd)
  sudo systemctl start monify

  # Check status
  %s status

  # View logs
  sudo journalctl -u monify -f

Documentation:
  https://github.com/monify-labs/agent

`, binaryName, defaultConfigPath, defaultEnvPath, binaryName)
}

func handleStatus() {
	// Load config
	cfg, err := config.LoadConfig(defaultConfigPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Check if authenticated
	authenticated := cfg.Server.Token != ""

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Monify Agent Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("Version:        %s\n", version)
	fmt.Printf("Config:         %s\n", defaultConfigPath)
	fmt.Printf("Authenticated:  %v\n", authenticated)

	if authenticated {
		// Mask token
		token := cfg.Server.Token
		if len(token) > 8 {
			token = token[:4] + "..." + token[len(token)-4:]
		}
		fmt.Printf("Token:          %s\n", token)
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	fmt.Printf("Hostname:       %s\n", hostname)
	fmt.Printf("Server URL:     %s\n", cfg.Server.URL)
	fmt.Printf("Interval:       %s\n", cfg.Collection.Interval)
	fmt.Println()

	// Check if service is running
	fmt.Println("Service Status:")
	fmt.Println("  Run: systemctl status monify")
	fmt.Println()

	if !authenticated {
		fmt.Println("⚠️  Not authenticated!")
		fmt.Println()
		fmt.Println("To authenticate:")
		fmt.Printf("  1. Edit %s\n", defaultEnvPath)
		fmt.Println("  2. Set TOKEN_DEVICE=your_token")
		fmt.Println("  3. Restart: sudo systemctl restart monify")
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

	// Acquire lock to ensure single instance
	agentLock := lock.NewLock("/var/run")
	if err := agentLock.Acquire(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Another instance may be running. Check: systemctl status monify\n")
		os.Exit(1)
	}
	defer agentLock.Release()

	// Load configuration
	cfg, err := config.LoadConfig(defaultConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		fmt.Fprintf(os.Stderr, "Config file: %s\n", defaultConfigPath)
		os.Exit(1)
	}

	// Check if authenticated
	if cfg.Server.Token == "" {
		fmt.Fprintf(os.Stderr, "Error: Not authenticated\n\n")
		fmt.Fprintf(os.Stderr, "Please set your token in: %s\n", defaultEnvPath)
		fmt.Fprintf(os.Stderr, "  TOKEN_DEVICE=your_token_here\n\n")
		fmt.Fprintf(os.Stderr, "Then restart: sudo systemctl restart monify\n")
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
