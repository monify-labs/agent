package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/monify-labs/agent/internal/agent"
	"github.com/monify-labs/agent/internal/config"
)

func main() {
	// Parse command line flags
	tokenFlag := flag.String("token", "", "Authentication token")
	urlFlag := flag.String("url", "", "Server URL")
	debugFlag := flag.Bool("debug", false, "Enable debug logging")
	versionFlag := flag.Bool("version", false, "Show version information")
	helpFlag := flag.Bool("help", false, "Show help message")

	flag.Usage = func() {
		fmt.Printf("Monify Agent v%s - Server Monitoring Agent\n\n", config.Version)
		fmt.Println("Usage:")
		fmt.Println("  monify [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  --token string    Authentication token (or set MONIFY_TOKEN)")
		fmt.Println("  --url string      Server URL (or set MONIFY_SERVER_URL)")
		fmt.Println("  --debug           Enable debug logging (or set MONIFY_DEBUG=true)")
		fmt.Println("  --version         Show version information")
		fmt.Println("  --help            Show this help message")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  monify --token=YOUR_TOKEN")
		fmt.Println("  monify --token=YOUR_TOKEN --debug")
		fmt.Println("  MONIFY_TOKEN=xxx monify")
	}

	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Printf("Monify Agent v%s\n", config.Version)
		fmt.Printf("Commit: %s\n", config.Commit)
		fmt.Printf("Build Date: %s\n", config.BuildDate)
		os.Exit(0)
	}

	// Handle help flag
	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	// Load environment file
	if err := config.LoadEnvFile(); err != nil {
		fmt.Printf("Warning: Failed to load env file: %v\n", err)
	}

	// Check if running as root (required for some metrics)
	if os.Geteuid() != 0 {
		fmt.Println("Warning: Running without root privileges. Some metrics may not be available.")
	}

	// Get token (flag overrides env)
	var token string
	if *tokenFlag != "" {
		token = *tokenFlag
	} else {
		var err error
		token, err = config.GetToken()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("Please set MONIFY_TOKEN environment variable or use --token flag.")
			os.Exit(1)
		}
	}

	// Get server URL (flag overrides env)
	var serverURL string
	if *urlFlag != "" {
		serverURL = *urlFlag
	} else {
		serverURL = config.GetServerURL()
	}

	// Check debug mode (flag overrides env)
	debug := *debugFlag || config.IsDebugMode()

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
