package collector

import (
	"context"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v3/disk"
)

// DiskCollector collects disk metrics
type DiskCollector struct {
	*BaseCollector
}

// NewDiskCollector creates a new disk collector
func NewDiskCollector(enabled bool) *DiskCollector {
	return &DiskCollector{
		BaseCollector: NewBaseCollector("disk", enabled),
	}
}

// Collect collects disk metrics
func (d *DiskCollector) Collect(ctx context.Context) (interface{}, error) {
	if !d.Enabled() {
		return nil, nil
	}

	var diskMetrics []models.DiskMetrics

	// Get all partitions
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, err
	}

	// Collect usage stats for each partition
	for _, partition := range partitions {
		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			// Skip partitions we can't access
			continue
		}

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

		// Get I/O stats for the device
		ioCounters, err := disk.IOCountersWithContext(ctx, partition.Device)
		if err == nil {
			if counter, ok := ioCounters[partition.Device]; ok {
				metric.ReadBytes = counter.ReadBytes
				metric.WriteBytes = counter.WriteBytes
				metric.ReadCount = counter.ReadCount
				metric.WriteCount = counter.WriteCount
			}
		}

		diskMetrics = append(diskMetrics, metric)
	}

	return diskMetrics, nil
}
