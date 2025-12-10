package collector

import (
	"context"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
)

// CPUCollector collects CPU metrics
type CPUCollector struct {
	*BaseCollector
}

// NewCPUCollector creates a new CPU collector
func NewCPUCollector(enabled bool) *CPUCollector {
	return &CPUCollector{
		BaseCollector: NewBaseCollector("cpu", enabled),
	}
}

// Collect collects CPU metrics
func (c *CPUCollector) Collect(ctx context.Context) (interface{}, error) {
	if !c.Enabled() {
		return nil, nil
	}

	metrics := &models.CPUMetrics{}

	// Get overall CPU usage (average over 1 second)
	percentages, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		return nil, err
	}
	if len(percentages) > 0 {
		metrics.UsagePercent = percentages[0]
	}

	// Get per-core CPU usage
	perCore, err := cpu.PercentWithContext(ctx, time.Second, true)
	if err != nil {
		return nil, err
	}
	metrics.PerCore = perCore

	// Get load average (only available on Unix-like systems)
	loadAvg, err := load.AvgWithContext(ctx)
	if err == nil {
		metrics.LoadAvg = []float64{loadAvg.Load1, loadAvg.Load5, loadAvg.Load15}
	}

	return metrics, nil
}
