package static

import (
	"context"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v4/disk"
)

// CollectDiskInventory gathers static disk/filesystem information
func CollectDiskInventory(ctx context.Context) ([]models.DiskInventoryMetrics, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, err
	}

	var disks []models.DiskInventoryMetrics

	for _, partition := range partitions {
		// Skip special filesystems
		if shouldSkipFilesystem(partition.Fstype) {
			continue
		}

		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			continue
		}

		disks = append(disks, models.DiskInventoryMetrics{
			Device:      partition.Device,
			MountPoint:  partition.Mountpoint,
			FSType:      partition.Fstype,
			Total:       usage.Total,
			InodesTotal: usage.InodesTotal,
		})
	}

	return disks, nil
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
