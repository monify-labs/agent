package collector

import (
	"context"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v3/net"
)

// NetworkCollector collects network metrics
type NetworkCollector struct {
	*BaseCollector
}

// NewNetworkCollector creates a new network collector
func NewNetworkCollector(enabled bool) *NetworkCollector {
	return &NetworkCollector{
		BaseCollector: NewBaseCollector("network", enabled),
	}
}

// Collect collects network metrics
func (n *NetworkCollector) Collect(ctx context.Context) (interface{}, error) {
	if !n.Enabled() {
		return nil, nil
	}

	var networkMetrics []models.NetworkMetrics

	// Get network I/O counters for all interfaces
	ioCounters, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		return nil, err
	}

	for _, counter := range ioCounters {
		// Skip loopback interface if desired
		if counter.Name == "lo" || counter.Name == "lo0" {
			continue
		}

		metric := models.NetworkMetrics{
			Interface:   counter.Name,
			BytesSent:   counter.BytesSent,
			BytesRecv:   counter.BytesRecv,
			PacketsSent: counter.PacketsSent,
			PacketsRecv: counter.PacketsRecv,
			ErrorsIn:    counter.Errin,
			ErrorsOut:   counter.Errout,
			DropsIn:     counter.Dropin,
			DropsOut:    counter.Dropout,
		}

		networkMetrics = append(networkMetrics, metric)
	}

	return networkMetrics, nil
}
