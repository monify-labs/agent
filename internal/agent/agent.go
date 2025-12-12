package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/monify-labs/agent/internal/collector"
	"github.com/monify-labs/agent/internal/scanner"
	"github.com/monify-labs/agent/internal/sender"
	"github.com/monify-labs/agent/pkg/config"
	"github.com/monify-labs/agent/pkg/models"
	"github.com/sirupsen/logrus"
)

// Agent is the main monitoring agent
type Agent struct {
	config     *config.Config
	logger     *logrus.Logger
	collectors []collector.Collector
	sender     sender.Sender
	scanner    *scanner.PortScanner

	// State
	mu             sync.RWMutex
	running        bool
	paused         bool   // When true, agent is paused (server disabled)
	hostname       string // Cached hostname from system
	startTime      time.Time
	lastCollection time.Time
	lastSend       time.Time
	metricsCount   uint64
	errorCount     uint64
	staticSent     bool // Track if static info has been sent

	// Channels
	stopChan    chan struct{}
	refreshChan chan struct{}
}

// NewAgent creates a new monitoring agent
func NewAgent(cfg *config.Config) (*Agent, error) {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(getLogLevel(cfg.Logging.Level))

	if cfg.Logging.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Set log output
	if cfg.Logging.File != "" {
		file, err := os.OpenFile(cfg.Logging.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		logger.SetOutput(file)
	}

	// Initialize collectors
	collectors := []collector.Collector{
		collector.NewCPUCollector(cfg.Metrics.CPU),
		collector.NewMemoryCollector(cfg.Metrics.Memory),
		collector.NewDiskCollector(cfg.Metrics.Disk),
		collector.NewNetworkCollector(cfg.Metrics.Network),
		collector.NewSystemCollector(cfg.Metrics.System),
	}

	// Initialize sender
	httpSender := sender.NewHTTPSender(cfg)

	// Initialize port scanner
	portScanner := scanner.NewPortScanner(
		cfg.PortScanner.Timeout,
		cfg.PortScanner.MaxWorkers,
	)

	return &Agent{
		config:      cfg,
		logger:      logger,
		collectors:  collectors,
		sender:      httpSender,
		scanner:     portScanner,
		stopChan:    make(chan struct{}),
		refreshChan: make(chan struct{}, 1),
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

	// Get hostname from system
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	a.hostname = hostname

	a.logger.WithFields(logrus.Fields{
		"hostname": a.hostname,
		"interval": a.config.Collection.Interval,
	}).Info("Agent starting")

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start collection loop
	ticker := time.NewTicker(a.config.Collection.Interval)
	defer ticker.Stop()

	// Collect immediately on start
	a.collectAndSend(ctx)

	// Start status writer
	go a.startStatusWriter(ctx)

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Agent stopping: context cancelled")
			return a.Stop()

		case <-a.stopChan:
			a.logger.Info("Agent stopping: stop signal received")
			return nil

		case sig := <-sigChan:
			switch sig {
			case syscall.SIGHUP:
				a.logger.Info("Received SIGHUP, reloading configuration")
				if err := a.reloadConfig(); err != nil {
					a.logger.WithError(err).Error("Failed to reload configuration")
				}
			case syscall.SIGINT, syscall.SIGTERM:
				a.logger.Info("Received shutdown signal")
				return a.Stop()
			}

		case <-ticker.C:
			a.collectAndSend(ctx)

		case <-a.refreshChan:
			a.logger.Info("Manual refresh triggered")
			a.collectAndSend(ctx)
		}
	}
}

// collectAndSend collects metrics and sends them to the server
func (a *Agent) collectAndSend(ctx context.Context) {
	a.mu.RLock()
	isPaused := a.paused
	a.mu.RUnlock()

	// If paused, send lightweight ping to check server status
	if isPaused {
		a.logger.Debug("Agent is paused, sending status check")
		a.checkServerStatus(ctx)
		return
	}

	a.logger.Debug("Starting metric collection")

	collectionCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Collect static info only if not sent yet
	var staticInfo *models.StaticMetrics
	a.mu.RLock()
	shouldSendStatic := !a.staticSent
	a.mu.RUnlock()

	if shouldSendStatic {
		// Collect static info from system collector
		for _, col := range a.collectors {
			if col.Name() == "system" && col.Enabled() {
				if systemCol, ok := col.(*collector.SystemCollector); ok {
					static, err := systemCol.CollectStatic(collectionCtx)
					if err != nil {
						a.logger.WithError(err).Error("Failed to collect static info")
					} else {
						staticInfo = static
						a.logger.Debug("Static info collected (will be sent once)")
					}
				}
				break
			}
		}
	}

	// Collect dynamic metrics concurrently
	var wg sync.WaitGroup
	metricsChan := make(chan interface{}, len(a.collectors))

	for _, col := range a.collectors {
		if !col.Enabled() {
			continue
		}

		wg.Add(1)
		go func(c collector.Collector) {
			defer wg.Done()

			// For system collector, collect dynamic metrics only
			if c.Name() == "system" {
				if systemCol, ok := c.(*collector.SystemCollector); ok {
					dynamic, err := systemCol.CollectDynamic(collectionCtx)
					if err != nil {
						a.logger.WithError(err).WithField("collector", c.Name()).Error("Collection failed")
						a.incrementErrorCount()
						return
					}
					if dynamic != nil {
						metricsChan <- dynamic
					}
				}
			} else {
				// For other collectors, use normal Collect
				metric, err := c.Collect(collectionCtx)
				if err != nil {
					a.logger.WithError(err).WithField("collector", c.Name()).Error("Collection failed")
					a.incrementErrorCount()
					return
				}
				if metric != nil {
					metricsChan <- metric
				}
			}
		}(col)
	}

	// Wait for all collectors
	wg.Wait()
	close(metricsChan)

	// Build dynamic metrics data
	dynamicMetrics := &models.DynamicMetrics{}
	for metric := range metricsChan {
		switch m := metric.(type) {
		case *models.CPUMetrics:
			dynamicMetrics.CPU = m
		case *models.MemoryMetrics:
			// Legacy path - keep for backward compatibility
			dynamicMetrics.Memory = m
		case map[string]interface{}:
			// New memory collector returns map with "memory" and "swap"
			if mem, ok := m["memory"].(*models.MemoryMetrics); ok {
				dynamicMetrics.Memory = mem
			}
			if swap, ok := m["swap"].(*models.SwapMetrics); ok {
				dynamicMetrics.Swap = swap
			}
			// New disk collector returns map with "disk_space", "disk_io", and "disk"
			if diskSpace, ok := m["disk_space"].(*models.DiskSpaceMetrics); ok {
				dynamicMetrics.DiskSpace = diskSpace
			}
			if diskIO, ok := m["disk_io"].(*models.DiskIOMetrics); ok {
				dynamicMetrics.DiskIO = diskIO
			}
			// Network aggregated metrics
			if netPub, ok := m["network_public"].(*models.NetworkAggregateMetrics); ok {
				dynamicMetrics.NetworkPublic = netPub
			}
			if netPriv, ok := m["network_private"].(*models.NetworkAggregateMetrics); ok {
				dynamicMetrics.NetworkPrivate = netPriv
			}
			if netHealth, ok := m["network_health"].(*models.NetworkHealthMetrics); ok {
				dynamicMetrics.NetworkHealth = netHealth
			}
		case *models.SystemDynamic:
			dynamicMetrics.System = m
		}
	}

	// Create payload with optimized structure
	// Cache hostname from static info on first collection
	if staticInfo != nil && a.hostname == "" {
		a.mu.Lock()
		a.hostname = staticInfo.Hostname
		a.mu.Unlock()
	}

	payload := &models.MetricPayload{
		Hostname:   a.hostname,
		Timestamp:  time.Now(),
		StaticInfo: staticInfo, // Will be nil if already sent
		Metrics:    dynamicMetrics,
	}

	// Send to server
	sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	serverResp, err := a.sender.Send(sendCtx, payload)
	if err != nil {
		a.logger.WithError(err).Error("Failed to send metrics")
		a.incrementErrorCount()
		return
	}

	a.mu.Lock()
	a.lastCollection = time.Now()
	a.lastSend = time.Now()
	a.metricsCount++
	if staticInfo != nil {
		a.staticSent = true
		a.logger.Info("Static info sent to server (will not be sent again)")
	}
	a.mu.Unlock()

	// Update status file
	go a.writeStatusFile()

	a.logger.Debug("Metrics sent successfully")

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

	a.logger.Info("Stopping agent")
	close(a.stopChan)
	a.running = false

	// Stop all collectors
	for _, col := range a.collectors {
		if err := col.Stop(); err != nil {
			a.logger.WithError(err).WithField("collector", col.Name()).Error("Failed to stop collector")
		}
	}

	// Close sender
	if err := a.sender.Close(); err != nil {
		a.logger.WithError(err).Error("Failed to close sender")
	}

	return nil
}

// RefreshNow triggers an immediate metric collection
func (a *Agent) RefreshNow() error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.running {
		return fmt.Errorf("agent is not running")
	}

	select {
	case a.refreshChan <- struct{}{}:
		a.logger.Info("Refresh triggered")
		return nil
	default:
		return fmt.Errorf("refresh already pending")
	}
}

// ScanPorts performs a port scan
func (a *Agent) ScanPorts(ctx context.Context, target string, ports []int) (*models.PortScanResult, error) {
	if !a.config.PortScanner.Enabled {
		return nil, fmt.Errorf("port scanner is disabled")
	}

	a.logger.WithFields(logrus.Fields{
		"target": target,
		"ports":  len(ports),
	}).Info("Starting port scan")

	result, err := a.scanner.Scan(ctx, target, ports)
	if err != nil {
		a.logger.WithError(err).Error("Port scan failed")
		return nil, err
	}

	a.logger.WithFields(logrus.Fields{
		"target":     target,
		"open_ports": len(result.OpenPorts),
		"scan_time":  result.ScanTime,
	}).Info("Port scan completed")

	return result, nil
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
		Version:        a.config.Agent.Version,
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
		a.logger.WithField("command", cmd.Command).Info("Processing server command")

		switch cmd.Command {
		case "update_config":
			// Update configuration from server params
			a.mu.Lock()
			configUpdated := false
			
			// Collection interval
			if interval, ok := cmd.Params["collection_interval"].(string); ok {
				if d, err := time.ParseDuration(interval); err == nil {
					oldInterval := a.config.Collection.Interval
					a.config.Collection.Interval = d
					a.logger.WithFields(logrus.Fields{
						"old": oldInterval,
						"new": d,
					}).Info("Collection interval updated from server")
					configUpdated = true
				}
			}
			
			// Log level
			if level, ok := cmd.Params["log_level"].(string); ok {
				validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
				if validLevels[strings.ToLower(level)] {
					a.config.Logging.Level = level
					a.logger.SetLevel(getLogLevel(level))
					a.logger.WithField("level", level).Info("Log level updated from server")
					configUpdated = true
				}
			}
			
			// Server URL (for future migrations)
			if url, ok := cmd.Params["server_url"].(string); ok {
				if url != "" {
					a.config.Server.URL = url
					a.logger.WithField("url", url).Info("Server URL updated from server")
					configUpdated = true
				}
			}
			
			// Metrics toggles
			if metricsParams, ok := cmd.Params["metrics"].(map[string]interface{}); ok {
				if cpu, ok := metricsParams["cpu"].(bool); ok {
					a.config.Metrics.CPU = cpu
				}
				if memory, ok := metricsParams["memory"].(bool); ok {
					a.config.Metrics.Memory = memory
				}
				if disk, ok := metricsParams["disk"].(bool); ok {
					a.config.Metrics.Disk = disk
				}
				if network, ok := metricsParams["network"].(bool); ok {
					a.config.Metrics.Network = network
				}
				if system, ok := metricsParams["system"].(bool); ok {
					a.config.Metrics.System = system
				}
				if configUpdated {
					a.logger.Info("Metrics collection settings updated from server")
				}
			}
			
			a.mu.Unlock()
			
			if configUpdated {
				a.logger.Info("Configuration updated successfully from server")
			}

		case "refresh":
			// Trigger immediate metric collection
			a.RefreshNow()

		case "scan_ports":
			// Trigger port scan
			if target, ok := cmd.Params["target"].(string); ok {
				ports := []int{} // Parse ports from params
				if portsParam, ok := cmd.Params["ports"].([]any); ok {
					for _, p := range portsParam {
						if port, ok := p.(float64); ok {
							ports = append(ports, int(port))
						}
					}
				}
				if len(ports) > 0 {
					go a.ScanPorts(ctx, target, ports)
				}
			}

		case "stop":
			// Server has been disabled, pause the agent
			reason := "Server disabled"
			if r, ok := cmd.Params["reason"].(string); ok {
				reason = r
			}
			a.logger.WithField("reason", reason).Warn("Server disabled - entering paused mode")
			a.mu.Lock()
			a.paused = true
			a.mu.Unlock()
			a.logger.Info("Agent paused - will check server status every 60s")

		case "uninstall":
			// Server has been deleted, uninstall the agent
			reason := "Server deleted"
			if r, ok := cmd.Params["reason"].(string); ok {
				reason = r
			}
			a.logger.WithField("reason", reason).Warn("Received uninstall command from server")
			go func() {
				time.Sleep(2 * time.Second) // Give time to log
				a.logger.Info("Running uninstall script...")
				a.runUninstallScript()
				a.Stop()
			}()

		default:
			a.logger.WithField("command", cmd.Command).Warn("Unknown server command")
		}
	}
}



// reloadConfig reloads the configuration from file
func (a *Agent) reloadConfig() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Info("Reloading configuration")

	// Load new config
	newConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Preserve agent version
	newConfig.Agent.Version = a.config.Agent.Version

	// Update configuration
	oldInterval := a.config.Collection.Interval
	a.config = newConfig

	a.logger.WithFields(logrus.Fields{
		"old_interval": oldInterval,
		"new_interval": newConfig.Collection.Interval,
		"log_level":    newConfig.Logging.Level,
	}).Info("Configuration reloaded successfully")

	// Update logger level if changed
	a.logger.SetLevel(getLogLevel(newConfig.Logging.Level))

	return nil
}

// incrementErrorCount increments the error counter
func (a *Agent) incrementErrorCount() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.errorCount++
}

// getLogLevel converts string log level to logrus level
func getLogLevel(level string) logrus.Level {
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

// checkServerStatus sends a lightweight ping to check if server is re-enabled
func (a *Agent) checkServerStatus(ctx context.Context) {
	// Send empty payload to check server status
	payload := &models.MetricPayload{
		Hostname:  a.hostname,
		Timestamp: time.Now(),
		Metrics:   &models.DynamicMetrics{}, // Empty metrics
	}

	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	serverResp, err := a.sender.Send(sendCtx, payload)
	if err != nil {
		a.logger.WithError(err).Debug("Status check failed (expected if server still disabled)")
		return
	}

	// Check if server sent resume command or no stop command
	if serverResp != nil && len(serverResp.Commands) > 0 {
		hasStopCommand := false
		for _, cmd := range serverResp.Commands {
			if cmd.Command == "stop" {
				hasStopCommand = true
				break
			}
		}

		// If no stop command, server is re-enabled
		if !hasStopCommand {
			a.mu.Lock()
			a.paused = false
			a.mu.Unlock()
			a.logger.Info("Server re-enabled - resuming normal operation")
		}

		// Process other commands
		a.processServerCommands(ctx, serverResp.Commands)
	} else {
		// No commands = server is active again
		a.mu.Lock()
		wasPaused := a.paused
		a.paused = false
		a.mu.Unlock()

		if wasPaused {
			a.logger.Info("Server re-enabled - resuming normal operation")
		}
	}
}

// runUninstallScript executes the uninstall script to remove the agent
func (a *Agent) runUninstallScript() {
	// Check if running in Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		a.logger.Info("Running in Docker - container will stop")
		// In Docker, just stopping is enough as container can be removed
		return
	}

	// Check if uninstall script exists
	uninstallScript := "/usr/local/bin/monify-agent-uninstall.sh"
	if _, err := os.Stat(uninstallScript); err == nil {
		a.logger.WithField("script", uninstallScript).Info("Executing uninstall script")

		cmd := exec.Command("bash", uninstallScript)
		output, err := cmd.CombinedOutput()

		if err != nil {
			a.logger.WithError(err).WithField("output", string(output)).Error("Uninstall script failed")
		} else {
			a.logger.WithField("output", string(output)).Info("Uninstall script completed")
		}
	} else {
		a.logger.Warn("Uninstall script not found - agent will only stop")
	}
}
