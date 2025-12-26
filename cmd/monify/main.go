package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/monify-labs/agent/internal/agent"
	"github.com/monify-labs/agent/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Load environment file
	if err := config.LoadEnvFile(); err != nil {
		fmt.Printf("Warning: Failed to load env file: %v\n", err)
	}

	command := os.Args[1]

	switch command {
	case "run":
		runAgent()
	case "status":
		showStatus()
	case "login":
		handleLogin()
	case "logout":
		handleLogout()
	case "update":
		handleUpdate()
	case "version":
		showVersion()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Monify Agent - Server Monitoring Agent

Usage:
  monify <command>

Commands:
  run       Start the monitoring agent
  status    Show agent status
  login     Login and save authentication token
  logout    Remove token and stop agent
  update    Update agent to latest version
  version   Show version information
  help      Show this help message

Environment Variables:
  MONIFY_TOKEN       Authentication token (required for run)
  MONIFY_SERVER_URL  Server URL (optional, default: https://api.monify.cloud/v1/agent/metrics)
  MONIFY_DEBUG       Enable debug logging (true/1)

Configuration File:
  /etc/monify/env    Environment variables file

Examples:
  sudo monify login YOUR_TOKEN
  sudo monify update
  monify status
  monify version`)
}

func runAgent() {
	// Check if running as root (required for some metrics)
	if os.Geteuid() != 0 {
		fmt.Println("Warning: Running without root privileges. Some metrics may not be available.")
	}

	// Get token
	token, err := config.GetToken()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Please run 'sudo monify login' to configure the agent.")
		os.Exit(1)
	}

	// Get server URL
	serverURL := config.GetServerURL()

	// Check debug mode
	debug := config.IsDebugMode()

	// Create agent
	a, err := agent.NewAgent(serverURL, token, debug)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		os.Exit(1)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal...")
		cancel()
	}()

	// Start agent
	fmt.Printf("Starting Monify Agent v%s\n", config.Version)
	fmt.Printf("Server: %s\n", serverURL)
	if debug {
		fmt.Println("Debug mode: enabled")
	}

	if err := a.Start(ctx); err != nil {
		fmt.Printf("Agent error: %v\n", err)
		os.Exit(1)
	}
}

func showStatus() {
	fmt.Println("Monify Agent Status")
	fmt.Println("-------------------")

	// Check if service is running using systemctl
	status, exitCode := getServiceStatus()
	fmt.Printf("Service: %s\n", status)

	// Check configuration
	token, tokenErr := config.GetToken()
	if tokenErr != nil {
		fmt.Println("Token: not configured")
	} else if token == "" {
		fmt.Println("Token: empty")
	} else {
		fmt.Println("Token: configured")
	}

	fmt.Printf("Server URL: %s\n", config.GetServerURL())
	fmt.Printf("Version: %s\n", config.Version)

	// Show troubleshooting hints if service is not running
	if status != "running" {
		fmt.Println("")
		fmt.Println("Troubleshooting:")

		if tokenErr != nil || token == "" {
			fmt.Println("  → Token not configured. Run: sudo monify login")
		} else if exitCode == 3 {
			fmt.Println("  → Authentication failed (invalid token).")
			fmt.Println("    Run: sudo monify login")
			fmt.Println("    Then: sudo systemctl start monify")
		} else {
			fmt.Println("  → Check logs: journalctl -u monify --no-pager -n 20")
			fmt.Println("  → Start service: sudo systemctl start monify")
		}
	}
}

func getServiceStatus() (string, int) {
	// Try systemctl first
	cmd := exec.Command("systemctl", "is-active", "monify")
	output, err := cmd.Output()
	
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	status := strings.TrimSpace(string(output))
	if status == "" {
		status = "unknown"
	}

	// Map systemctl status to friendly names
	switch status {
	case "active":
		return "running", 0
	case "inactive":
		// Check why it's inactive - get last exit code
		cmd := exec.Command("systemctl", "show", "monify", "--property=ExecMainStatus")
		output, _ := cmd.Output()
		if strings.Contains(string(output), "=3") {
			return "stopped (auth failed)", 3
		}
		return "stopped", exitCode
	case "failed":
		return "failed", exitCode
	default:
		return status, exitCode
	}
}

func handleLogin() {
	// Check if running as root
	if os.Geteuid() != 0 {
		fmt.Println("Error: login requires root privileges.")
		fmt.Println("Please run: sudo monify login [TOKEN]")
		os.Exit(1)
	}

	var token string

	// Check if token is passed as argument
	if len(os.Args) >= 3 {
		token = os.Args[2]
	} else {
		// Interactive mode
		fmt.Println("Monify Agent Login")
		fmt.Println("------------------")
		fmt.Print("Enter your server token: ")

		_, err := fmt.Scanln(&token)
		if err != nil {
			fmt.Println("Error reading token")
			os.Exit(1)
		}
	}

	if token == "" {
		fmt.Println("Error: Token cannot be empty")
		os.Exit(1)
	}

	// Save token to env file
	err := config.SaveEnvFile(map[string]string{
		"MONIFY_TOKEN": token,
	})
	if err != nil {
		fmt.Printf("Error saving token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Token saved successfully!")
	fmt.Println("")
	fmt.Println("To start the agent, run:")
	fmt.Println("  sudo systemctl start monify")
}

func handleLogout() {
	// Check if running as root
	if os.Geteuid() != 0 {
		fmt.Println("Error: logout requires root privileges.")
		fmt.Println("Please run: sudo monify logout")
		os.Exit(1)
	}

	fmt.Println("Logging out...")

	// Stop service first
	cmd := exec.Command("systemctl", "stop", "monify")
	cmd.Run() // Ignore error if service not running

	// Remove token from env file
	err := config.SaveEnvFile(map[string]string{
		"MONIFY_TOKEN": "",
	})
	if err != nil {
		fmt.Printf("Error removing token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Service stopped")
	fmt.Println("✓ Token removed")
	fmt.Println("")
	fmt.Println("To login again: sudo monify login [TOKEN]")
}

func handleUpdate() {
	// Check if running as root
	if os.Geteuid() != 0 {
		fmt.Println("Error: update requires root privileges.")
		fmt.Println("Please run: sudo monify update")
		os.Exit(1)
	}

	fmt.Println("Updating Monify Agent...")
	fmt.Printf("Current version: %s\n", config.Version)
	fmt.Println("")

	// Run install script without token (it will use existing token)
	cmd := exec.Command("bash", "-c", "curl -sSL https://monify.cloud/install.sh | bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Update failed: %v\n", err)
		os.Exit(1)
	}
}

func showVersion() {
	fmt.Printf("Monify Agent v%s\n", config.Version)
	fmt.Printf("Commit: %s\n", config.Commit)
	fmt.Printf("Build Date: %s\n", config.BuildDate)
	fmt.Println("https://monify.cloud")
}
