package collector

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	gopsutilNet "github.com/shirou/gopsutil/v3/net"
)

// networkSample stores a single network measurement sample
type networkSample struct {
	InterfaceStats map[string]struct {
		BytesSent uint64
		BytesRecv uint64
	}
	Timestamp time.Time
}

// NetworkCollector collects network metrics with continuous sampling
type NetworkCollector struct {
	*BaseCollector
	mu         sync.Mutex
	samples    []*networkSample // Buffered samples (1 per second)
	publicIPs  map[string]bool  // Cache of public IPs
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

// NewNetworkCollector creates a new network collector
func NewNetworkCollector(enabled bool) *NetworkCollector {
	c := &NetworkCollector{
		BaseCollector: NewBaseCollector("network", enabled),
		samples:       make([]*networkSample, 0, 60),
		publicIPs:     make(map[string]bool),
	}

	if enabled {
		ctx, cancel := context.WithCancel(context.Background())
		c.cancelFunc = cancel
		c.wg.Add(1)
		go c.startSampling(ctx)
	}

	return c
}

// startSampling collects network metrics every second
func (n *NetworkCollector) startSampling(ctx context.Context) {
	defer n.wg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Initial sample
	n.collectSample(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n.collectSample(ctx)
		}
	}
}

// collectSample captures a single network snapshot
func (n *NetworkCollector) collectSample(ctx context.Context) {
	ioCounters, err := gopsutilNet.IOCountersWithContext(ctx, true)
	if err != nil {
		return
	}

	sample := &networkSample{
		InterfaceStats: make(map[string]struct {
			BytesSent uint64
			BytesRecv uint64
		}),
		Timestamp: time.Now(),
	}

	for _, counter := range ioCounters {
		// Skip loopback
		if counter.Name == "lo" || counter.Name == "lo0" {
			continue
		}

		sample.InterfaceStats[counter.Name] = struct {
			BytesSent uint64
			BytesRecv uint64
		}{
			BytesSent: counter.BytesSent,
			BytesRecv: counter.BytesRecv,
		}
	}

	n.mu.Lock()
	if len(n.samples) > 600 {
		// Drop oldest samples
		copy(n.samples, n.samples[len(n.samples)-600:])
		n.samples = n.samples[:600]
	}
	n.samples = append(n.samples, sample)
	n.mu.Unlock()
}

// Stop stops the background sampling
func (n *NetworkCollector) Stop() error {
	if n.cancelFunc != nil {
		n.cancelFunc()
		n.wg.Wait()
	}
	return nil
}

// Collect collects network metrics (averaging samples)
func (n *NetworkCollector) Collect(ctx context.Context) (interface{}, error) {
	if !n.Enabled() {
		return nil, nil
	}

	// Get samples and clear buffer
	n.mu.Lock()
	samples := n.samples
	n.samples = make([]*networkSample, 0, 60)
	n.mu.Unlock()

	// If no samples, take one immediately
	if len(samples) == 0 {
		n.collectSample(ctx)
		n.mu.Lock()
		if len(n.samples) > 0 {
			samples = append(samples, n.samples...)
			n.samples = make([]*networkSample, 0, 60)
		}
		n.mu.Unlock()
	}

	if len(samples) == 0 {
		return nil, nil
	}

	// Get current I/O counters for cumulative values and static info
	ioCounters, err := gopsutilNet.IOCountersWithContext(ctx, true)
	if err != nil {
		return nil, err
	}

	// Get all network interfaces for IP classification (once per collection)
	interfaces, err := net.Interfaces()
	if err != nil {
		interfaces = nil
	}

	// Calculate bandwidth rates from samples
	interfaceRates := n.calculateRatesFromSamples(samples)

	// Aggregated network metrics by type
	var publicSendMbps, publicRecvMbps float64
	var privateSendMbps, privateRecvMbps float64
	var publicBytesSent, publicBytesRecv uint64
	var privateBytesSent, privateBytesRecv uint64

	// Aggregated health metrics
	var totalErrorsIn, totalErrorsOut, totalDropsIn, totalDropsOut uint64

	// Keep per-interface metrics for backward compatibility
	var networkMetrics []models.NetworkMetrics

	for _, counter := range ioCounters {
		// Skip loopback interface
		if counter.Name == "lo" || counter.Name == "lo0" {
			continue
		}

		// Get calculated rates for this interface
		rates, hasRates := interfaceRates[counter.Name]

		// Classify interface as public or private
		interfaceType := n.classifyInterface(counter.Name, interfaces)

		metric := models.NetworkMetrics{
			Interface:   counter.Name,
			Type:        interfaceType,
			BytesSent:   counter.BytesSent,
			BytesRecv:   counter.BytesRecv,
			PacketsSent: counter.PacketsSent,
			PacketsRecv: counter.PacketsRecv,
			ErrorsIn:    counter.Errin,
			ErrorsOut:   counter.Errout,
			DropsIn:     counter.Dropin,
			DropsOut:    counter.Dropout,
		}

		if hasRates {
			metric.SendRate = rates.SendRate
			metric.RecvRate = rates.RecvRate
			metric.SendRateMbps = (rates.SendRate * 8) / 1_000_000
			metric.RecvRateMbps = (rates.RecvRate * 8) / 1_000_000

			// Aggregate by type
			if interfaceType == "public" {
				publicSendMbps += metric.SendRateMbps
				publicRecvMbps += metric.RecvRateMbps
				publicBytesSent += counter.BytesSent
				publicBytesRecv += counter.BytesRecv
			} else {
				privateSendMbps += metric.SendRateMbps
				privateRecvMbps += metric.RecvRateMbps
				privateBytesSent += counter.BytesSent
				privateBytesRecv += counter.BytesRecv
			}
		}

		// Aggregate health metrics (all interfaces)
		totalErrorsIn += counter.Errin
		totalErrorsOut += counter.Errout
		totalDropsIn += counter.Dropin
		totalDropsOut += counter.Dropout

		networkMetrics = append(networkMetrics, metric)
	}

	// Create aggregated public network metrics
	networkPublic := &models.NetworkAggregateMetrics{
		SendMbps:    publicSendMbps,
		RecvMbps:    publicRecvMbps,
		TotalSentGB: float64(publicBytesSent) / 1_000_000_000,
		TotalRecvGB: float64(publicBytesRecv) / 1_000_000_000,
	}

	// Create aggregated private network metrics
	networkPrivate := &models.NetworkAggregateMetrics{
		SendMbps:    privateSendMbps,
		RecvMbps:    privateRecvMbps,
		TotalSentGB: float64(privateBytesSent) / 1_000_000_000,
		TotalRecvGB: float64(privateBytesRecv) / 1_000_000_000,
	}

	// Create network health metrics
	networkHealth := &models.NetworkHealthMetrics{
		ErrorsIn:  totalErrorsIn,
		ErrorsOut: totalErrorsOut,
		DropsIn:   totalDropsIn,
		DropsOut:  totalDropsOut,
	}

	// Return both aggregated and per-interface metrics
	return map[string]interface{}{
		"network_public":  networkPublic,
		"network_private": networkPrivate,
		"network_health":  networkHealth,
		"network":         networkMetrics, // Deprecated - for backward compatibility
	}, nil
}

// calculateRatesFromSamples calculates average bandwidth rates from samples
func (n *NetworkCollector) calculateRatesFromSamples(samples []*networkSample) map[string]struct {
	SendRate float64
	RecvRate float64
} {
	rates := make(map[string]struct {
		SendRate float64
		RecvRate float64
	})

	if len(samples) < 2 {
		return rates
	}

	// Get all unique interface names
	interfaceNames := make(map[string]bool)
	for _, sample := range samples {
		for name := range sample.InterfaceStats {
			interfaceNames[name] = true
		}
	}

	// Calculate average rate for each interface
	for ifaceName := range interfaceNames {
		var totalSendRate, totalRecvRate float64
		var count int

		// Calculate rate between consecutive samples
		for i := 1; i < len(samples); i++ {
			prev := samples[i-1]
			curr := samples[i]

			prevStats, hasPrev := prev.InterfaceStats[ifaceName]
			currStats, hasCurr := curr.InterfaceStats[ifaceName]

			if !hasPrev || !hasCurr {
				continue
			}

			duration := curr.Timestamp.Sub(prev.Timestamp).Seconds()
			if duration > 0 {
				sendRate := float64(currStats.BytesSent-prevStats.BytesSent) / duration
				recvRate := float64(currStats.BytesRecv-prevStats.BytesRecv) / duration

				totalSendRate += sendRate
				totalRecvRate += recvRate
				count++
			}
		}

		if count > 0 {
			rates[ifaceName] = struct {
				SendRate float64
				RecvRate float64
			}{
				SendRate: totalSendRate / float64(count),
				RecvRate: totalRecvRate / float64(count),
			}
		}
	}

	return rates
}

// classifyInterface determines if an interface is public or private
func (n *NetworkCollector) classifyInterface(ifaceName string, interfaces []net.Interface) string {
	// Check cache first
	if isPublic, exists := n.publicIPs[ifaceName]; exists {
		if isPublic {
			return "public"
		}
		return "private"
	}

	// Default to private
	interfaceType := "private"

	// Find the interface
	if interfaces != nil {
		for _, iface := range interfaces {
			if iface.Name != ifaceName {
				continue
			}

			// Get addresses for this interface
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			// Check each address
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				if ip == nil {
					continue
				}

				// Check if it's a public IP
				if isPublicIP(ip) {
					interfaceType = "public"
					n.publicIPs[ifaceName] = true
					return interfaceType
				}
			}
			break
		}
	}

	// Common interface name patterns for public interfaces
	if strings.HasPrefix(ifaceName, "eth") ||
	   strings.HasPrefix(ifaceName, "en") ||
	   strings.HasPrefix(ifaceName, "ens") ||
	   strings.HasPrefix(ifaceName, "eno") {
		// These could be public, but default to private if no public IP found
		n.publicIPs[ifaceName] = false
	}

	return interfaceType
}

// isPublicIP checks if an IP address is public (not private/loopback/etc)
func isPublicIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return false
	}

	// Check for private IP ranges (IPv4)
	if ip4 := ip.To4(); ip4 != nil {
		// 10.0.0.0/8
		if ip4[0] == 10 {
			return false
		}
		// 172.16.0.0/12
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return false
		}
		// 192.168.0.0/16
		if ip4[0] == 192 && ip4[1] == 168 {
			return false
		}
		// 169.254.0.0/16 (link-local)
		if ip4[0] == 169 && ip4[1] == 254 {
			return false
		}
		return true
	}

	// Check for private IP ranges (IPv6)
	if len(ip) == net.IPv6len {
		// fc00::/7 (unique local address)
		if ip[0] >= 0xfc && ip[0] <= 0xfd {
			return false
		}
		// fe80::/10 (link-local)
		if ip[0] == 0xfe && ip[1] >= 0x80 && ip[1] <= 0xbf {
			return false
		}
		return true
	}

	return false
}
