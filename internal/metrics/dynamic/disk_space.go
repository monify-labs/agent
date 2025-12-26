package dynamic

import (
	"context"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v4/disk"
)

// CollectDiskSpace aggregates disk space usage across all partitions (no sampling needed)
func CollectDiskSpace(ctx context.Context) (*models.DiskSpaceMetrics, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, err
	}

	var totalSpace, usedSpace, freeSpace uint64

	for _, partition := range partitions {
		// Skip special filesystems
		if shouldSkipFilesystem(partition.Fstype) {
			continue
		}

		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			continue
		}

		totalSpace += usage.Total
		usedSpace += usage.Used
		freeSpace += usage.Free
	}

	// Calculate usage percentage
	usedPercent := 0.0
	if totalSpace > 0 {
		usedPercent = float64(usedSpace) / float64(totalSpace) * 100
	}

	return &models.DiskSpaceMetrics{
		Total:       totalSpace,
		Used:        usedSpace,
		Free:        freeSpace,
		UsedPercent: usedPercent,
	}, nil
}

// shouldSkipFilesystem determines if a filesystem type should be skipped
func shouldSkipFilesystem(fstype string) bool {
	skipTypes := map[string]bool{
		"tmpfs":    true,
		"devtmpfs": true,
		"devfs":    true,
		"proc":     true,
		"sysfs":    true,
		"cgroup":   true,
		"cgroup2":  true,
		"nsfs":     true,
		"overlay":  true,
		"squashfs": true,
		"iso9660":  true,
	}

	return skipTypes[fstype]
}
