package static

import (
	"context"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// HardwareInfo contains CPU and memory hardware specifications
type HardwareInfo struct {
	CPUModel    string
	CPUCores    int
	CPUThreads  int
	TotalMemory uint64
}

// CollectHardwareInfo gathers CPU and memory hardware specifications
func CollectHardwareInfo(ctx context.Context) (*HardwareInfo, error) {
	// Get CPU info
	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get CPU counts
	physicalCores, err := cpu.CountsWithContext(ctx, false)
	if err != nil {
		return nil, err
	}

	logicalCores, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return nil, err
	}

	// Get memory info
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get CPU model from first CPU (usually all are the same)
	cpuModel := ""
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	return &HardwareInfo{
		CPUModel:    cpuModel,
		CPUCores:    physicalCores,
		CPUThreads:  logicalCores,
		TotalMemory: memInfo.Total,
	}, nil
}
