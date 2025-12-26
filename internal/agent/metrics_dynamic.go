package agent

import (
	"context"
	"sync"

	"github.com/monify-labs/agent/internal/metrics/dynamic"
	"github.com/monify-labs/agent/pkg/models"
)

// DynamicCollector orchestrates collection of all dynamic metrics
type DynamicCollector struct {
	cpu     *dynamic.CPUCollector
	memory  *dynamic.MemoryCollector
	diskIO  *dynamic.DiskIOCollector
	network *dynamic.NetworkCollector
}

// NewDynamicCollector creates a new dynamic metrics collector
func NewDynamicCollector() *DynamicCollector {
	return &DynamicCollector{
		cpu:     dynamic.NewCPUCollector(),
		memory:  dynamic.NewMemoryCollector(),
		diskIO:  dynamic.NewDiskIOCollector(),
		network: dynamic.NewNetworkCollector(),
	}
}

// Start begins background sampling for all dynamic collectors
func (d *DynamicCollector) Start() {
	d.cpu.Start()
	d.memory.Start()
	d.diskIO.Start()
	d.network.Start()
}

// Stop halts background sampling for all dynamic collectors
func (d *DynamicCollector) Stop() {
	d.cpu.Stop()
	d.memory.Stop()
	d.diskIO.Stop()
	d.network.Stop()
}

// Collect gathers all dynamic metrics in parallel
func (d *DynamicCollector) Collect(ctx context.Context) (*models.DynamicMetrics, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	result := &models.DynamicMetrics{}

	// CPU (with sampling)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if cpu, err := d.cpu.Collect(ctx); err == nil {
			mu.Lock()
			result.CPU = cpu
			mu.Unlock()
		}
	}()

	// Memory (with sampling)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if mem, err := d.memory.Collect(ctx); err == nil {
			mu.Lock()
			result.Memory = mem
			mu.Unlock()
		}
	}()

	// Swap (instant query)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if swap, err := dynamic.CollectSwap(ctx); err == nil {
			mu.Lock()
			result.Swap = swap
			mu.Unlock()
		}
	}()

	// Disk Space (instant aggregation)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if diskSpace, err := dynamic.CollectDiskSpace(ctx); err == nil {
			mu.Lock()
			result.DiskSpace = diskSpace
			mu.Unlock()
		}
	}()

	// Disk I/O (with sampling)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if diskIO, err := d.diskIO.Collect(ctx); err == nil {
			mu.Lock()
			result.DiskIO = diskIO
			mu.Unlock()
		}
	}()

	// Network (with sampling)
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Public network
		if pub, err := d.network.CollectPublic(ctx); err == nil {
			mu.Lock()
			result.NetworkPublic = pub
			mu.Unlock()
		}

		// Private network
		if priv, err := d.network.CollectPrivate(ctx); err == nil {
			mu.Lock()
			result.NetworkPrivate = priv
			mu.Unlock()
		}

		// Network health
		if health, err := d.network.CollectHealth(ctx); err == nil {
			mu.Lock()
			result.NetworkHealth = health
			mu.Unlock()
		}
	}()

	// System dynamic (instant query)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if sysDynamic, err := dynamic.CollectSystemDynamic(ctx); err == nil {
			mu.Lock()
			result.System = sysDynamic
			mu.Unlock()
		}
	}()

	wg.Wait()
	return result, nil
}
