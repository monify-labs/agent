package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const statusFile = "/var/log/monify/status.json"

// writeStatusFile writes the current agent status to a file
func (a *Agent) writeStatusFile() {
	status := a.GetStatus()
	
	// Create directory if not exists
	if err := os.MkdirAll(filepath.Dir(statusFile), 0755); err != nil {
		a.logger.WithError(err).Error("Failed to create status file directory")
		return
	}

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		a.logger.WithError(err).Error("Failed to marshal status")
		return
	}

	if err := os.WriteFile(statusFile, data, 0644); err != nil {
		a.logger.WithError(err).Error("Failed to write status file")
	}
}

// startStatusWriter starts a goroutine to periodically write status file
func (a *Agent) startStatusWriter(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial write
	a.writeStatusFile()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.writeStatusFile()
		}
	}
}
