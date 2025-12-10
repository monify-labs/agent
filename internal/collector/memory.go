package collector

import (
	"context"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v3/mem"
)

// MemoryCollector collects memory metrics
type MemoryCollector struct {
	*BaseCollector
}

// NewMemoryCollector creates a new memory collector
func NewMemoryCollector(enabled bool) *MemoryCollector {
	return &MemoryCollector{
		BaseCollector: NewBaseCollector("memory", enabled),
	}
}

// Collect collects memory metrics
func (m *MemoryCollector) Collect(ctx context.Context) (interface{}, error) {
	if !m.Enabled() {
		return nil, nil
	}

	metrics := &models.MemoryMetrics{}

	// Get virtual memory stats
	vmStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	metrics.Total = vmStat.Total
	metrics.Used = vmStat.Used
	metrics.Free = vmStat.Free
	metrics.Available = vmStat.Available
	metrics.UsedPercent = vmStat.UsedPercent

	// Platform-specific fields
	if vmStat.Cached > 0 {
		metrics.Cached = vmStat.Cached
	}
	if vmStat.Buffers > 0 {
		metrics.Buffers = vmStat.Buffers
	}

	// Get swap memory stats
	swapStat, err := mem.SwapMemoryWithContext(ctx)
	if err == nil {
		metrics.SwapTotal = swapStat.Total
		metrics.SwapUsed = swapStat.Used
		metrics.SwapFree = swapStat.Free
	}

	return metrics, nil
}
