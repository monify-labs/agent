package dynamic

import (
	"context"
	"sync"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

const maxSamples = 600 // 10 minutes at 1 second interval

// cpuSample represents a single CPU usage sample
type cpuSample struct {
	usagePercent float64
	timestamp    time.Time
}

// CPUCollector samples CPU usage in background
type CPUCollector struct {
	mu      sync.Mutex
	samples []cpuSample
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewCPUCollector creates a new CPU collector
func NewCPUCollector() *CPUCollector {
	return &CPUCollector{
		samples: make([]cpuSample, 0, maxSamples),
	}
}

// Start begins background sampling
func (c *CPUCollector) Start() {
	c.ctx, c.cancel = context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.ctx.Done():
				return
			case <-ticker.C:
				c.sample()
			}
		}
	}()
}

// Stop halts background sampling
func (c *CPUCollector) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

// sample takes a single CPU usage measurement
func (c *CPUCollector) sample() {
	// Get overall CPU usage (not per-core for cleaner averaging)
	percentages, err := cpu.Percent(0, false)
	if err != nil || len(percentages) == 0 {
		return
	}

	sample := cpuSample{
		usagePercent: percentages[0],
		timestamp:    time.Now(),
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Add sample to buffer
	c.samples = append(c.samples, sample)

	// Keep only last maxSamples
	if len(c.samples) > maxSamples {
		c.samples = c.samples[len(c.samples)-maxSamples:]
	}
}

// Collect drains samples and returns averaged metrics
func (c *CPUCollector) Collect(ctx context.Context) (*models.CPUMetrics, error) {
	// Get load averages (instant, no sampling needed)
	loadAvg, err := load.AvgWithContext(ctx)
	if err != nil {
		return nil, err
	}

	// Drain samples
	c.mu.Lock()
	samples := make([]cpuSample, len(c.samples))
	copy(samples, c.samples)
	c.samples = c.samples[:0] // Clear buffer
	c.mu.Unlock()

	// Calculate average CPU usage from samples
	avgUsage := 0.0
	if len(samples) > 0 {
		sum := 0.0
		for _, s := range samples {
			sum += s.usagePercent
		}
		avgUsage = sum / float64(len(samples))
	}

	return &models.CPUMetrics{
		UsagePercent: avgUsage,
		LoadAvg1m:    loadAvg.Load1,
		LoadAvg5m:    loadAvg.Load5,
		LoadAvg15m:   loadAvg.Load15,
	}, nil
}
