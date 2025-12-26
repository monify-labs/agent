package agent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/monify-labs/agent/internal/config"
	"github.com/monify-labs/agent/internal/sender"
	"github.com/monify-labs/agent/pkg/models"
)

// Agent is the main monitoring agent
type Agent struct {
	serverURL        string
	token            string
	debug            bool
	sender           sender.Sender
	staticCollector  *StaticCollector
	dynamicCollector *DynamicCollector

	// State
	mu             sync.RWMutex
	running        bool
	authFailed     bool // When true, authentication has failed permanently
	hostname       string
	startTime      time.Time
	lastCollection time.Time
	lastSend       time.Time
	metricsCount   uint64
	errorCount     uint64

	// Channels
	stopChan chan struct{}
}

// NewAgent creates a new monitoring agent
func NewAgent(serverURL, token string, debug bool) (*Agent, error) {
	// Initialize collectors
	staticCollector := NewStaticCollector()
	dynamicCollector := NewDynamicCollector()

	// Initialize sender
	httpSender := sender.NewHTTPSender(serverURL, token)

	return &Agent{
		serverURL:        serverURL,
		token:            token,
		debug:            debug,
		sender:           httpSender,
		staticCollector:  staticCollector,
		dynamicCollector: dynamicCollector,
		stopChan:         make(chan struct{}),
	}, nil
}

// Start starts the agent
func (a *Agent) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("agent is already running")
	}
	a.running = true
	a.startTime = time.Now()
	a.mu.Unlock()

	// Start background samplers
	a.dynamicCollector.Start()
	defer a.dynamicCollector.Stop()

	// Initial static collection to get hostname
	staticMetrics, err := a.staticCollector.Collect(ctx)
	if err != nil {
		log.Printf("WARN: %v - %s", err, "Failed to collect initial static metrics")
	} else {
		a.hostname = staticMetrics.Hostname
	}

	log.Printf("INFO: %s [%s=%v]", "Agent starting", "hostname", a.hostname)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start collection loop
	ticker := time.NewTicker(config.CollectionInterval)
	defer ticker.Stop()

	// Collect immediately on start
	a.collectAndSend(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Printf("INFO: %s", "Agent stopping: context cancelled")
			return a.Stop()

		case <-a.stopChan:
			log.Printf("INFO: %s", "Agent stopping: stop signal received")
			return nil

		case sig := <-sigChan:
			switch sig {
			case syscall.SIGHUP:
				log.Printf("INFO: %s", "Received SIGHUP (configuration reload not supported)")
			case syscall.SIGINT, syscall.SIGTERM:
				log.Printf("INFO: %s", "Received shutdown signal")
				return a.Stop()
			}

		case <-ticker.C:
			// Check if auth failed
			a.mu.RLock()
			isAuthFailed := a.authFailed
			a.mu.RUnlock()

			if isAuthFailed {
				log.Printf("ERROR: %s", "Authentication failed - stopping agent")
				log.Printf("ERROR: %s", "Agent stopped. Please login to restart:")
				log.Printf("ERROR: %s", "  sudo monify login")

				if err := a.Stop(); err != nil {
					log.Printf("ERROR: %v - %s", err, "Error during stop")
				}

				// Exit with special code to prevent systemd restart
				os.Exit(3)
			}

			a.collectAndSend(ctx)
		}
	}
}

// collectAndSend collects metrics and sends them to the server
func (a *Agent) collectAndSend(ctx context.Context) {
	// Single timeout for entire operation (collection + send)
	opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Check if static metrics need refreshing
	var staticMetrics *models.StaticMetrics
	if a.staticCollector.ShouldRefresh() {
		if a.debug {
			log.Printf("INFO: Refreshing static metrics")
		}
		static, err := a.staticCollector.Collect(opCtx)
		if err != nil {
			log.Printf("ERROR: Failed to collect static metrics: %v", err)
		} else {
			staticMetrics = static
			// Update hostname if changed
			if a.hostname == "" || a.hostname != static.Hostname {
				a.mu.Lock()
				a.hostname = static.Hostname
				a.mu.Unlock()
			}
		}
	}

	// Always collect dynamic metrics
	dynamicMetrics, err := a.dynamicCollector.Collect(opCtx)
	if err != nil {
		log.Printf("ERROR: Failed to collect dynamic metrics: %v", err)
		a.incrementErrorCount()
		return
	}

	// Create payload
	payload := &models.MetricPayload{
		Hostname:       a.hostname,
		Timestamp:      time.Now(),
		StaticMetrics:  staticMetrics, // nil if not refreshed
		DynamicMetrics: dynamicMetrics,
	}

	// Debug mode - log detailed payload
	if a.debug {
		cpuUsage := 0.0
		memUsage := 0.0
		if dynamicMetrics != nil {
			if dynamicMetrics.CPU != nil {
				cpuUsage = dynamicMetrics.CPU.UsagePercent
			}
			if dynamicMetrics.Memory != nil {
				memUsage = dynamicMetrics.Memory.UsedPercent
			}
		}
		log.Printf("DEBUG: Sending metrics [hostname=%s static=%v cpu=%.1f%% mem=%.1f%%]",
			payload.Hostname, staticMetrics != nil, cpuUsage, memUsage)
	}

	// Send to server
	serverResp, err := a.sender.Send(opCtx, payload)
	if err != nil {
		// Check if this is an authentication error
		if errors.Is(err, sender.ErrUnauthorized) {
			log.Printf("ERROR: Authentication failed - token invalid/expired")
			log.Printf("ERROR: Please login again: sudo monify login")

			// Mark auth as failed
			a.mu.Lock()
			a.authFailed = true
			a.mu.Unlock()

			return
		}

		log.Printf("ERROR: Failed to send metrics: %v", err)
		a.incrementErrorCount()
		return
	}

	// Update stats (single lock)
	now := time.Now()
	a.mu.Lock()
	a.lastCollection = now
	a.lastSend = now
	a.metricsCount++
	a.mu.Unlock()

	if a.debug {
		log.Printf("DEBUG: Metrics sent successfully")
	}

	// Process server commands if any
	if serverResp != nil && len(serverResp.Commands) > 0 {
		a.processServerCommands(ctx, serverResp.Commands)
	}
}

// Stop stops the agent gracefully
func (a *Agent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return fmt.Errorf("agent is not running")
	}

	log.Printf("INFO: %s", "Stopping agent")
	close(a.stopChan)
	a.running = false

	// Stop dynamic collectors
	a.dynamicCollector.Stop()

	// Close sender
	if err := a.sender.Close(); err != nil {
		log.Printf("ERROR: %v - %s", err, "Failed to close sender")
	}

	return nil
}

// GetStatus returns the current status of the agent
func (a *Agent) GetStatus() *models.AgentStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	status := "stopped"
	if a.running {
		status = "running"
	}

	uptime := uint64(0)
	if !a.startTime.IsZero() {
		uptime = uint64(time.Since(a.startTime).Seconds())
	}

	return &models.AgentStatus{
		Hostname:       a.hostname,
		Version:        config.Version,
		Uptime:         uptime,
		LastCollection: a.lastCollection,
		LastSend:       a.lastSend,
		MetricsCount:   a.metricsCount,
		ErrorCount:     a.errorCount,
		Status:         status,
	}
}

// processServerCommands processes commands received from server
func (a *Agent) processServerCommands(ctx context.Context, commands []models.ServerCommand) {
	for _, cmd := range commands {
		if a.debug {
			log.Printf("INFO: Processing server command [command=%s]", cmd.Command)
		}

		switch cmd.Command {
		case "uninstall":
			reason := "Server deleted"
			if r, ok := cmd.Params["reason"].(string); ok {
				reason = r
			}
			log.Printf("WARN: Received uninstall command [reason=%s]", reason)
			go func() {
				time.Sleep(2 * time.Second)
				a.runUninstallScript()
			}()

		default:
			if a.debug {
				log.Printf("DEBUG: Ignoring unsupported command [command=%s]", cmd.Command)
			}
		}
	}
}

// runUninstallScript executes the uninstall script to remove the agent
func (a *Agent) runUninstallScript() {
	log.Printf("INFO: Executing uninstall script")
	exec.Command("bash", "-c", "curl -sSL https://monify.cloud/uninstall.sh | sudo bash").Start()
}

// incrementErrorCount increments the error counter
func (a *Agent) incrementErrorCount() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.errorCount++
}
