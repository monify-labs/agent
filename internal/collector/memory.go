package collector

import (
	"context"
	"sync"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v3/mem"
)

// MemoryCollector collects memory metrics
type MemoryCollector struct {
	*BaseCollector
	mu           sync.Mutex
	samples      []*models.MemoryMetrics // Buffered samples
	cancelFunc   context.CancelFunc
	wg           sync.WaitGroup
}

// NewMemoryCollector creates a new memory collector
func NewMemoryCollector(enabled bool) *MemoryCollector {
	c := &MemoryCollector{
		BaseCollector: NewBaseCollector("memory", enabled),
		samples:       make([]*models.MemoryMetrics, 0, 60),
	}

	if enabled {
		ctx, cancel := context.WithCancel(context.Background())
		c.cancelFunc = cancel
		c.wg.Add(1)
		go c.startSampling(ctx)
	}

	return c
}

// startSampling collects memory metrics every second
func (c *MemoryCollector) startSampling(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Initial sample
	c.collectSample(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collectSample(ctx)
		}
	}
}

// collectSample captures a single memory snapshot
func (c *MemoryCollector) collectSample(ctx context.Context) {
	metrics := &models.MemoryMetrics{}

	// Get virtual memory stats
	vmStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return
	}

	metrics.Total = vmStat.Total
	metrics.Used = vmStat.Used
	metrics.Free = vmStat.Free
	metrics.Available = vmStat.Available
	metrics.UsedPercent = vmStat.UsedPercent
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

	c.mu.Lock()
	if len(c.samples) > 600 {
		// Drop oldest samples
		copy(c.samples, c.samples[len(c.samples)-600:])
		c.samples = c.samples[:600]
	}
	c.samples = append(c.samples, metrics)
	c.mu.Unlock()
}

// Stop stops the background sampling
func (c *MemoryCollector) Stop() error {
	if c.cancelFunc != nil {
		c.cancelFunc()
		c.wg.Wait()
	}
	return nil
}

// Collect collects memory metrics (averaging samples)
func (c *MemoryCollector) Collect(ctx context.Context) (interface{}, error) {
	if !c.Enabled() {
		return nil, nil
	}

	c.mu.Lock()
	samples := c.samples
	c.samples = make([]*models.MemoryMetrics, 0, 60)
	c.mu.Unlock()

	// If no samples, take one immediately
	if len(samples) == 0 {
		c.collectSample(ctx)
		c.mu.Lock()
		if len(c.samples) > 0 {
			samples = append(samples, c.samples[0])
			c.samples = make([]*models.MemoryMetrics, 0, 60)
		}
		c.mu.Unlock()
	}

	if len(samples) == 0 {
		return nil, nil
	}

	// Calculate averages
	avg := &models.MemoryMetrics{}
	count := uint64(len(samples))
	countF := float64(count)

	var usedPercentSum float64

	for _, s := range samples {
		avg.Total = s.Total // Total shouldn't change, just take last (or average, effectively same)
		avg.Used += s.Used
		avg.Free += s.Free
		avg.Available += s.Available
		
		usedPercentSum += s.UsedPercent
		
		avg.Cached += s.Cached
		avg.Buffers += s.Buffers
		
		avg.SwapTotal = s.SwapTotal
		avg.SwapUsed += s.SwapUsed
		avg.SwapFree += s.SwapFree
	}

	avg.Used = avg.Used / count
	avg.Free = avg.Free / count
	avg.Available = avg.Available / count
	avg.UsedPercent = usedPercentSum / countF
	avg.Cached = avg.Cached / count
	avg.Buffers = avg.Buffers / count
	avg.SwapUsed = avg.SwapUsed / count
	avg.SwapFree = avg.SwapFree / count

	return avg, nil
}
