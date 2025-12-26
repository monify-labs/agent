package agent

import (
	"context"
	"sync"
	"time"

	"github.com/monify-labs/agent/internal/metrics/static"
	"github.com/monify-labs/agent/pkg/models"
)

const staticRefreshInterval = 1 * time.Hour

// StaticCollector orchestrates collection of all static metrics
type StaticCollector struct {
	networkInfo *static.NetworkInfoCollector
	lastRefresh time.Time
	cache       *models.StaticMetrics
	mu          sync.RWMutex
}

// NewStaticCollector creates a new static metrics collector
func NewStaticCollector() *StaticCollector {
	return &StaticCollector{
		networkInfo: static.NewNetworkInfoCollector(),
	}
}

// Collect gathers all static metrics in parallel
func (s *StaticCollector) Collect(ctx context.Context) (*models.StaticMetrics, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	result := &models.StaticMetrics{}

	// System info
	wg.Add(1)
	go func() {
		defer wg.Done()
		if info, err := static.CollectSystemInfo(ctx); err == nil {
			mu.Lock()
			result.Platform = info.Platform
			result.PlatformFamily = info.PlatformFamily
			result.PlatformVersion = info.PlatformVersion
			result.OS = info.OS
			result.Arch = info.Arch
			result.KernelVersion = info.KernelVersion
			result.KernelArch = info.KernelArch
			result.Virtualization = info.Virtualization
			result.HostID = info.HostID
			mu.Unlock()
		}
	}()

	// Hardware info
	wg.Add(1)
	go func() {
		defer wg.Done()
		if info, err := static.CollectHardwareInfo(ctx); err == nil {
			mu.Lock()
			result.CPUModel = info.CPUModel
			result.CPUCores = info.CPUCores
			result.CPUThreads = info.CPUThreads
			result.TotalMemory = info.TotalMemory
			mu.Unlock()
		}
	}()

	// Network info (uses cached collector for public IP)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if info, err := s.networkInfo.Collect(ctx); err == nil {
			mu.Lock()
			result.InternalIPs = info.InternalIPs
			result.PublicIP = info.PublicIP
			result.Hostname = info.Hostname
			result.FQDN = info.FQDN
			result.Timezone = info.Timezone
			mu.Unlock()
		}
	}()

	// Cloud info
	wg.Add(1)
	go func() {
		defer wg.Done()
		if info, err := static.DetectCloudProvider(ctx); err == nil {
			mu.Lock()
			result.Region = info.Region
			result.InstanceType = info.InstanceType
			mu.Unlock()
		}
	}()

	// Disk inventory
	wg.Add(1)
	go func() {
		defer wg.Done()
		if disks, err := static.CollectDiskInventory(ctx); err == nil {
			mu.Lock()
			result.Disks = disks
			mu.Unlock()
		}
	}()

	wg.Wait()

	// Update cache
	s.mu.Lock()
	s.cache = result
	s.lastRefresh = time.Now()
	s.mu.Unlock()

	return result, nil
}

// ShouldRefresh checks if static metrics need refreshing
func (s *StaticCollector) ShouldRefresh() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Force refresh if never collected
	if s.cache == nil {
		return true
	}

	return time.Since(s.lastRefresh) >= staticRefreshInterval
}

// GetCached returns cached static metrics
func (s *StaticCollector) GetCached() *models.StaticMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache
}
