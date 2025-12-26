package dynamic

import (
	"context"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/process"
)

// CollectSystemDynamic gathers frequently-changing system metrics (no sampling needed)
func CollectSystemDynamic(ctx context.Context) (*models.SystemMetrics, error) {
	// Get uptime and boot time
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get process count
	processes, err := process.ProcessesWithContext(ctx)
	processCount := uint64(0)
	if err == nil {
		processCount = uint64(len(processes))
	}

	return &models.SystemMetrics{
		Uptime:       info.Uptime,
		BootTime:     info.BootTime,
		ProcessCount: processCount,
	}, nil
}
