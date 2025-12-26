package dynamic

import (
	"context"
	"sync"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v4/mem"
)

// memorySample represents a single memory usage sample
type memorySample struct {
	total       uint64
	used        uint64
	free        uint64
	available   uint64
	usedPercent float64
	cached      uint64
	buffers     uint64
	timestamp   time.Time
}

// MemoryCollector samples memory usage in background
type MemoryCollector struct {
	mu      sync.Mutex
	samples []memorySample
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewMemoryCollector creates a new memory collector
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{
		samples: make([]memorySample, 0, maxSamples),
	}
}

// Start begins background sampling
func (m *MemoryCollector) Start() {
	m.ctx, m.cancel = context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
				m.sample()
			}
		}
	}()
}

// Stop halts background sampling
func (m *MemoryCollector) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
}

// sample takes a single memory usage measurement
func (m *MemoryCollector) sample() {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return
	}

	sample := memorySample{
		total:       vmem.Total,
		used:        vmem.Used,
		free:        vmem.Free,
		available:   vmem.Available,
		usedPercent: vmem.UsedPercent,
		cached:      vmem.Cached,
		buffers:     vmem.Buffers,
		timestamp:   time.Now(),
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.samples = append(m.samples, sample)

	if len(m.samples) > maxSamples {
		m.samples = m.samples[len(m.samples)-maxSamples:]
	}
}

// Collect drains samples and returns averaged metrics
func (m *MemoryCollector) Collect(ctx context.Context) (*models.MemoryMetrics, error) {
	// Drain samples
	m.mu.Lock()
	samples := make([]memorySample, len(m.samples))
	copy(samples, m.samples)
	m.samples = m.samples[:0]
	m.mu.Unlock()

	// If no samples, do immediate query
	if len(samples) == 0 {
		vmem, err := mem.VirtualMemoryWithContext(ctx)
		if err != nil {
			return nil, err
		}

		return &models.MemoryMetrics{
			Total:       vmem.Total,
			Used:        vmem.Used,
			Free:        vmem.Free,
			Available:   vmem.Available,
			UsedPercent: vmem.UsedPercent,
			Cached:      vmem.Cached,
			Buffers:     vmem.Buffers,
		}, nil
	}

	// Calculate averages from samples
	var avgMetrics models.MemoryMetrics
	var sumTotal, sumUsed, sumFree, sumAvailable, sumCached, sumBuffers uint64
	var sumUsedPercent float64

	for _, s := range samples {
		sumTotal += s.total
		sumUsed += s.used
		sumFree += s.free
		sumAvailable += s.available
		sumUsedPercent += s.usedPercent
		sumCached += s.cached
		sumBuffers += s.buffers
	}

	count := uint64(len(samples))
	avgMetrics.Total = sumTotal / count
	avgMetrics.Used = sumUsed / count
	avgMetrics.Free = sumFree / count
	avgMetrics.Available = sumAvailable / count
	avgMetrics.UsedPercent = sumUsedPercent / float64(len(samples))
	avgMetrics.Cached = sumCached / count
	avgMetrics.Buffers = sumBuffers / count

	return &avgMetrics, nil
}
