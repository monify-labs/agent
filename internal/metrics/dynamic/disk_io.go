package dynamic

import (
	"context"
	"sync"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v4/disk"
)

// ioStats represents I/O statistics for a device
type ioStats struct {
	readBytes  uint64
	writeBytes uint64
	readCount  uint64
	writeCount uint64
}

// diskIOSample represents a single disk I/O sample
type diskIOSample struct {
	devices   map[string]ioStats
	timestamp time.Time
}

// DiskIOCollector samples disk I/O in background
type DiskIOCollector struct {
	mu      sync.Mutex
	samples []diskIOSample
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewDiskIOCollector creates a new disk I/O collector
func NewDiskIOCollector() *DiskIOCollector {
	return &DiskIOCollector{
		samples: make([]diskIOSample, 0, maxSamples),
	}
}

// Start begins background sampling
func (d *DiskIOCollector) Start() {
	d.ctx, d.cancel = context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				d.sample()
			}
		}
	}()
}

// Stop halts background sampling
func (d *DiskIOCollector) Stop() {
	if d.cancel != nil {
		d.cancel()
	}
}

// sample takes a single disk I/O measurement
func (d *DiskIOCollector) sample() {
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return
	}

	devices := make(map[string]ioStats)
	for device, counters := range ioCounters {
		devices[device] = ioStats{
			readBytes:  counters.ReadBytes,
			writeBytes: counters.WriteBytes,
			readCount:  counters.ReadCount,
			writeCount: counters.WriteCount,
		}
	}

	sample := diskIOSample{
		devices:   devices,
		timestamp: time.Now(),
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.samples = append(d.samples, sample)

	if len(d.samples) > maxSamples {
		d.samples = d.samples[len(d.samples)-maxSamples:]
	}
}

// Collect drains samples and calculates I/O rates
func (d *DiskIOCollector) Collect(ctx context.Context) (*models.DiskIOMetrics, error) {
	// Drain samples
	d.mu.Lock()
	samples := make([]diskIOSample, len(d.samples))
	copy(samples, d.samples)
	d.samples = d.samples[:0]
	d.mu.Unlock()

	// Need at least 2 samples to calculate rates
	if len(samples) < 2 {
		return &models.DiskIOMetrics{
			ReadMBps:  0,
			WriteMBps: 0,
			ReadIOPS:  0,
			WriteIOPS: 0,
		}, nil
	}

	// Calculate rates between consecutive samples and average them
	var totalReadMBps, totalWriteMBps, totalReadIOPS, totalWriteIOPS float64
	rateCount := 0

	for i := 1; i < len(samples); i++ {
		prev := samples[i-1]
		curr := samples[i]

		duration := curr.timestamp.Sub(prev.timestamp).Seconds()
		if duration <= 0 {
			continue
		}

		var readBytesDelta, writeBytesDelta, readCountDelta, writeCountDelta uint64

		// Aggregate deltas across all devices
		for device, currStats := range curr.devices {
			if prevStats, ok := prev.devices[device]; ok {
				readBytesDelta += currStats.readBytes - prevStats.readBytes
				writeBytesDelta += currStats.writeBytes - prevStats.writeBytes
				readCountDelta += currStats.readCount - prevStats.readCount
				writeCountDelta += currStats.writeCount - prevStats.writeCount
			}
		}

		// Calculate rates
		readMBps := float64(readBytesDelta) / duration / 1024 / 1024
		writeMBps := float64(writeBytesDelta) / duration / 1024 / 1024
		readIOPS := float64(readCountDelta) / duration
		writeIOPS := float64(writeCountDelta) / duration

		totalReadMBps += readMBps
		totalWriteMBps += writeMBps
		totalReadIOPS += readIOPS
		totalWriteIOPS += writeIOPS
		rateCount++
	}

	// Average the rates
	if rateCount > 0 {
		return &models.DiskIOMetrics{
			ReadMBps:  totalReadMBps / float64(rateCount),
			WriteMBps: totalWriteMBps / float64(rateCount),
			ReadIOPS:  totalReadIOPS / float64(rateCount),
			WriteIOPS: totalWriteIOPS / float64(rateCount),
		}, nil
	}

	return &models.DiskIOMetrics{}, nil
}
