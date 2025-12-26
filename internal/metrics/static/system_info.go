package static

import (
	"context"

	"github.com/shirou/gopsutil/v4/host"
)

// SystemInfo contains OS and kernel information
type SystemInfo struct {
	Platform        string
	PlatformFamily  string
	PlatformVersion string
	OS              string
	Arch            string
	KernelVersion   string
	KernelArch      string
	Virtualization  string
	HostID          string
}

// CollectSystemInfo gathers OS, kernel, and virtualization information
func CollectSystemInfo(ctx context.Context) (*SystemInfo, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		Platform:        info.Platform,
		PlatformFamily:  info.PlatformFamily,
		PlatformVersion: info.PlatformVersion,
		OS:              info.OS,
		Arch:            info.KernelArch,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
		Virtualization:  info.VirtualizationSystem,
		HostID:          info.HostID,
	}, nil
}
