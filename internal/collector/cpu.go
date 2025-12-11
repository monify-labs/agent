package collector

import (
	"context"
	"sync"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
)

// CPUCollector collects CPU metrics
type CPUCollector struct {
	*BaseCollector
	mu           sync.Mutex
	samples      [][]float64 // List of per-core usage samples
	cancelFunc   context.CancelFunc
	sampleTicker *time.Ticker
	wg           sync.WaitGroup
}

// NewCPUCollector creates a new CPU collector
func NewCPUCollector(enabled bool) *CPUCollector {
	c := &CPUCollector{
		BaseCollector: NewBaseCollector("cpu", enabled),
		samples:       make([][]float64, 0, 60),
	}

	if enabled {
		ctx, cancel := context.WithCancel(context.Background())
		c.cancelFunc = cancel
		c.wg.Add(1)
		go c.startSampling(ctx)
	}

	return c
}

// startSampling runs in a background goroutine to collect CPU metrics every second
func (c *CPUCollector) startSampling(ctx context.Context) {
	defer c.wg.Done()

	// Initial collection to prime the stats
	// We allow 1s for the measurement
	cpu.PercentWithContext(ctx, time.Second, true)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Measure CPU usage over the next 1 second
			// This blocks for 1 second effectively acting as our ticker
			perCore, err := cpu.PercentWithContext(ctx, time.Second, true)
			if err != nil {
				// If context was cancelled, we exit
				if ctx.Err() != nil {
					return
				}
				// Otherwise, just skip this sample (maybe log if we had a logger here)
				// Small sleep to avoid tight loop on persistent error
				time.Sleep(time.Second)
				continue
			}

			c.mu.Lock()
			// Keep a rolling buffer if somehow it's not being drained, limit to 600 (10 mins)
			if len(c.samples) > 600 {
				// Drop oldest samples
				copy(c.samples, c.samples[len(c.samples)-600:])
				c.samples = c.samples[:600]
			}
			c.samples = append(c.samples, perCore)
			c.mu.Unlock()
		}
	}
}

// Stop stops the background sampling
func (c *CPUCollector) Stop() error {
	if c.cancelFunc != nil {
		c.cancelFunc()
		c.wg.Wait()
	}
	return nil
}

// Collect collects CPU metrics
func (c *CPUCollector) Collect(ctx context.Context) (interface{}, error) {
	if !c.Enabled() {
		return nil, nil
	}

	metrics := &models.CPUMetrics{}

	// Calculate averages from collected samples
	c.mu.Lock()
	samples := c.samples
	// Reset sample buffer for next period
	c.samples = make([][]float64, 0, 60)
	c.mu.Unlock()

	if len(samples) > 0 {
		numCores := len(samples[0])
		avgPerCore := make([]float64, numCores)
		var totalSum float64

		// Sum up all samples
		for _, sample := range samples {
			// Handle case where core count might change (unlikely but safe)
			limit := len(sample)
			if limit > numCores {
				limit = numCores
			}
			for i := 0; i < limit; i++ {
				avgPerCore[i] += sample[i]
			}
		}

		// Calculate averages
		for i := range avgPerCore {
			avgPerCore[i] = avgPerCore[i] / float64(len(samples))
			totalSum += avgPerCore[i]
		}

		metrics.PerCore = avgPerCore
		if numCores > 0 {
			metrics.UsagePercent = totalSum / float64(numCores)
		}
	} else {
		// Fallback if no samples (e.g. just started) - take a quick snapshot
		// Note: This takes 1s, blocking the collection
		perCore, err := cpu.PercentWithContext(ctx, time.Second, true)
		if err != nil {
			return nil, err
		}
		metrics.PerCore = perCore
		
		var total float64
		for _, p := range perCore {
			total += p
		}
		if len(perCore) > 0 {
			metrics.UsagePercent = total / float64(len(perCore))
		}
	}

	// Get load average (only available on Unix-like systems)
	loadAvg, err := load.AvgWithContext(ctx)
	if err == nil {
		metrics.LoadAvg = []float64{loadAvg.Load1, loadAvg.Load5, loadAvg.Load15}
	}

	return metrics, nil
}
