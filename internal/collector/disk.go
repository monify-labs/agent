package collector

import (
	"context"
	"sync"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v3/disk"
)

// diskIOSample stores a single disk I/O measurement sample
type diskIOSample struct {
	DeviceStats map[string]struct {
		ReadBytes  uint64
		WriteBytes uint64
		ReadCount  uint64
		WriteCount uint64
	}
	Timestamp time.Time
}

// DiskCollector collects disk metrics with continuous I/O sampling
type DiskCollector struct {
	*BaseCollector
	mu         sync.Mutex
	samples    []*diskIOSample // Buffered I/O samples (1 per second)
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

// NewDiskCollector creates a new disk collector
func NewDiskCollector(enabled bool) *DiskCollector {
	c := &DiskCollector{
		BaseCollector: NewBaseCollector("disk", enabled),
		samples:       make([]*diskIOSample, 0, 60),
	}

	if enabled {
		ctx, cancel := context.WithCancel(context.Background())
		c.cancelFunc = cancel
		c.wg.Add(1)
		go c.startSampling(ctx)
	}

	return c
}

// startSampling collects disk I/O metrics every second
func (d *DiskCollector) startSampling(ctx context.Context) {
	defer d.wg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Initial sample
	d.collectIOSample(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.collectIOSample(ctx)
		}
	}
}

// collectIOSample captures a single disk I/O snapshot
func (d *DiskCollector) collectIOSample(ctx context.Context) {
	// Get all partitions to know which devices to track
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return
	}

	sample := &diskIOSample{
		DeviceStats: make(map[string]struct {
			ReadBytes  uint64
			WriteBytes uint64
			ReadCount  uint64
			WriteCount uint64
		}),
		Timestamp: time.Now(),
	}

	for _, partition := range partitions {
		// Get I/O stats for this device
		ioCounters, err := disk.IOCountersWithContext(ctx, partition.Device)
		if err != nil {
			continue
		}

		if counter, ok := ioCounters[partition.Device]; ok {
			sample.DeviceStats[partition.Device] = struct {
				ReadBytes  uint64
				WriteBytes uint64
				ReadCount  uint64
				WriteCount uint64
			}{
				ReadBytes:  counter.ReadBytes,
				WriteBytes: counter.WriteBytes,
				ReadCount:  counter.ReadCount,
				WriteCount: counter.WriteCount,
			}
		}
	}

	d.mu.Lock()
	if len(d.samples) > 600 {
		// Drop oldest samples
		copy(d.samples, d.samples[len(d.samples)-600:])
		d.samples = d.samples[:600]
	}
	d.samples = append(d.samples, sample)
	d.mu.Unlock()
}

// Stop stops the background sampling
func (d *DiskCollector) Stop() error {
	if d.cancelFunc != nil {
		d.cancelFunc()
		d.wg.Wait()
	}
	return nil
}

// Collect collects disk metrics (averaging I/O samples)
func (d *DiskCollector) Collect(ctx context.Context) (interface{}, error) {
	if !d.Enabled() {
		return nil, nil
	}

	// Get samples and clear buffer
	d.mu.Lock()
	samples := d.samples
	d.samples = make([]*diskIOSample, 0, 60)
	d.mu.Unlock()

	// If no samples, take one immediately
	if len(samples) == 0 {
		d.collectIOSample(ctx)
		d.mu.Lock()
		if len(d.samples) > 0 {
			samples = append(samples, d.samples...)
			d.samples = make([]*diskIOSample, 0, 60)
		}
		d.mu.Unlock()
	}

	// Get all partitions
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, err
	}

	// Calculate I/O rates from samples
	deviceRates := d.calculateRatesFromSamples(samples)

	// Aggregated disk space metrics
	var totalSpace, usedSpace, freeSpace uint64

	// Aggregated disk I/O metrics
	var totalReadMBps, totalWriteMBps, totalReadIOPS, totalWriteIOPS float64

	// Keep per-partition metrics for backward compatibility
	var diskMetrics []models.DiskMetrics

	// Collect usage stats for each partition
	for _, partition := range partitions {
		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			// Skip partitions we can't access
			continue
		}

		// Skip partitions with 0 total size (pseudo-filesystems, etc.)
		if usage.Total == 0 {
			continue
		}

		// Aggregate disk space
		totalSpace += usage.Total
		usedSpace += usage.Used
		freeSpace += usage.Free

		metric := models.DiskMetrics{
			Device:      partition.Device,
			MountPoint:  partition.Mountpoint,
			FSType:      partition.Fstype,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
			InodesTotal: usage.InodesTotal,
			InodesUsed:  usage.InodesUsed,
			InodesFree:  usage.InodesFree,
		}

		// Get I/O stats for the device (cumulative counters)
		ioCounters, err := disk.IOCountersWithContext(ctx, partition.Device)
		if err == nil {
			if counter, ok := ioCounters[partition.Device]; ok {
				metric.ReadBytes = counter.ReadBytes
				metric.WriteBytes = counter.WriteBytes
				metric.ReadCount = counter.ReadCount
				metric.WriteCount = counter.WriteCount

				// Get calculated rates for this device
				if rates, hasRates := deviceRates[partition.Device]; hasRates {
					metric.ReadRate = rates.ReadRate
					metric.WriteRate = rates.WriteRate
					metric.ReadRateMBps = rates.ReadRate / 1_000_000   // MB/s
					metric.WriteRateMBps = rates.WriteRate / 1_000_000 // MB/s
					metric.ReadIOPS = rates.ReadIOPS
					metric.WriteIOPS = rates.WriteIOPS

					// Aggregate I/O metrics
					totalReadMBps += metric.ReadRateMBps
					totalWriteMBps += metric.WriteRateMBps
					totalReadIOPS += metric.ReadIOPS
					totalWriteIOPS += metric.WriteIOPS
				}
			}
		}

		diskMetrics = append(diskMetrics, metric)
	}

	// Create aggregated disk space metrics
	diskSpace := &models.DiskSpaceMetrics{
		Total: totalSpace,
		Used:  usedSpace,
		Free:  freeSpace,
	}
	if totalSpace > 0 {
		diskSpace.UsedPercent = (float64(usedSpace) / float64(totalSpace)) * 100
	}

	// Create aggregated disk I/O metrics
	diskIO := &models.DiskIOMetrics{
		ReadMBps:  totalReadMBps,
		WriteMBps: totalWriteMBps,
		ReadIOPS:  totalReadIOPS,
		WriteIOPS: totalWriteIOPS,
	}

	// Return both aggregated and per-partition metrics
	return map[string]interface{}{
		"disk_space": diskSpace,
		"disk_io":    diskIO,
		"disk":       diskMetrics, // Deprecated - for backward compatibility
	}, nil
}

// calculateRatesFromSamples calculates average I/O rates from samples
func (d *DiskCollector) calculateRatesFromSamples(samples []*diskIOSample) map[string]struct {
	ReadRate  float64
	WriteRate float64
	ReadIOPS  float64
	WriteIOPS float64
} {
	rates := make(map[string]struct {
		ReadRate  float64
		WriteRate float64
		ReadIOPS  float64
		WriteIOPS float64
	})

	if len(samples) < 2 {
		return rates
	}

	// Get all unique device names
	deviceNames := make(map[string]bool)
	for _, sample := range samples {
		for name := range sample.DeviceStats {
			deviceNames[name] = true
		}
	}

	// Calculate average rate for each device
	for deviceName := range deviceNames {
		var totalReadRate, totalWriteRate float64
		var totalReadIOPS, totalWriteIOPS float64
		var count int

		// Calculate rate between consecutive samples
		for i := 1; i < len(samples); i++ {
			prev := samples[i-1]
			curr := samples[i]

			prevStats, hasPrev := prev.DeviceStats[deviceName]
			currStats, hasCurr := curr.DeviceStats[deviceName]

			if !hasPrev || !hasCurr {
				continue
			}

			duration := curr.Timestamp.Sub(prev.Timestamp).Seconds()
			if duration > 0 {
				readRate := float64(currStats.ReadBytes-prevStats.ReadBytes) / duration
				writeRate := float64(currStats.WriteBytes-prevStats.WriteBytes) / duration
				readIOPS := float64(currStats.ReadCount-prevStats.ReadCount) / duration
				writeIOPS := float64(currStats.WriteCount-prevStats.WriteCount) / duration

				totalReadRate += readRate
				totalWriteRate += writeRate
				totalReadIOPS += readIOPS
				totalWriteIOPS += writeIOPS
				count++
			}
		}

		if count > 0 {
			rates[deviceName] = struct {
				ReadRate  float64
				WriteRate float64
				ReadIOPS  float64
				WriteIOPS float64
			}{
				ReadRate:  totalReadRate / float64(count),
				WriteRate: totalWriteRate / float64(count),
				ReadIOPS:  totalReadIOPS / float64(count),
				WriteIOPS: totalWriteIOPS / float64(count),
			}
		}
	}

	return rates
}
