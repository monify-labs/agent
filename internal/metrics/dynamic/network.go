package dynamic

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	gopsutilNet "github.com/shirou/gopsutil/v4/net"
)

// networkStats represents network statistics for an interface
type networkStats struct {
	bytesSent uint64
	bytesRecv uint64
	errorsIn  uint64
	errorsOut uint64
	dropsIn   uint64
	dropsOut  uint64
}

// networkSample represents a single network sample
type networkSample struct {
	interfaces map[string]networkStats
	timestamp  time.Time
}

// NetworkCollector samples network I/O in background
type NetworkCollector struct {
	mu             sync.Mutex
	samples        []networkSample
	interfaceTypes map[string]string // cache: interface -> "public" or "private"
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewNetworkCollector creates a new network collector
func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{
		samples:        make([]networkSample, 0, maxSamples),
		interfaceTypes: make(map[string]string),
	}
}

// Start begins background sampling
func (n *NetworkCollector) Start() {
	n.ctx, n.cancel = context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-n.ctx.Done():
				return
			case <-ticker.C:
				n.sample()
			}
		}
	}()
}

// Stop halts background sampling
func (n *NetworkCollector) Stop() {
	if n.cancel != nil {
		n.cancel()
	}
}

// sample takes a single network I/O measurement
func (n *NetworkCollector) sample() {
	ioCounters, err := gopsutilNet.IOCounters(true) // per interface
	if err != nil {
		return
	}

	interfaces := make(map[string]networkStats)
	for _, counter := range ioCounters {
		interfaces[counter.Name] = networkStats{
			bytesSent: counter.BytesSent,
			bytesRecv: counter.BytesRecv,
			errorsIn:  counter.Errin,
			errorsOut: counter.Errout,
			dropsIn:   counter.Dropin,
			dropsOut:  counter.Dropout,
		}

		// Classify interface type on first encounter
		n.mu.Lock()
		if _, exists := n.interfaceTypes[counter.Name]; !exists {
			n.interfaceTypes[counter.Name] = n.classifyInterface(counter.Name)
		}
		n.mu.Unlock()
	}

	sample := networkSample{
		interfaces: interfaces,
		timestamp:  time.Now(),
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	n.samples = append(n.samples, sample)

	if len(n.samples) > maxSamples {
		n.samples = n.samples[len(n.samples)-maxSamples:]
	}
}

// classifyInterface determines if an interface is public or private
func (n *NetworkCollector) classifyInterface(ifaceName string) string {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "private" // default to private on error
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "private"
	}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			// Try parsing as plain IP
			ip = net.ParseIP(addr.String())
		}

		if ip != nil && !ip.IsLoopback() && !ip.IsUnspecified() {
			if isPrivateIP(ip) {
				return "private"
			}
			// Found a public IP
			return "public"
		}
	}

	return "private" // default
}

// CollectPublic collects public network bandwidth metrics
func (n *NetworkCollector) CollectPublic(ctx context.Context) (*models.NetworkAggregateMetrics, error) {
	return n.collectByType("public")
}

// CollectPrivate collects private network bandwidth metrics
func (n *NetworkCollector) CollectPrivate(ctx context.Context) (*models.NetworkAggregateMetrics, error) {
	return n.collectByType("private")
}

// collectByType calculates bandwidth metrics for interfaces of a specific type
func (n *NetworkCollector) collectByType(ifaceType string) (*models.NetworkAggregateMetrics, error) {
	// Drain samples
	n.mu.Lock()
	samples := make([]networkSample, len(n.samples))
	copy(samples, n.samples)
	interfaceTypes := make(map[string]string)
	for k, v := range n.interfaceTypes {
		interfaceTypes[k] = v
	}
	n.samples = n.samples[:0]
	n.mu.Unlock()

	// Need at least 2 samples to calculate rates
	if len(samples) < 2 {
		return &models.NetworkAggregateMetrics{
			SendMbps:    0,
			RecvMbps:    0,
			TotalSentGB: 0,
			TotalRecvGB: 0,
		}, nil
	}

	// Calculate cumulative totals from last sample
	lastSample := samples[len(samples)-1]
	var totalSentBytes, totalRecvBytes uint64

	for ifaceName, stats := range lastSample.interfaces {
		if interfaceTypes[ifaceName] == ifaceType {
			totalSentBytes += stats.bytesSent
			totalRecvBytes += stats.bytesRecv
		}
	}

	// Calculate bandwidth rates between consecutive samples and average them
	var totalSendMbps, totalRecvMbps float64
	rateCount := 0

	for i := 1; i < len(samples); i++ {
		prev := samples[i-1]
		curr := samples[i]

		duration := curr.timestamp.Sub(prev.timestamp).Seconds()
		if duration <= 0 {
			continue
		}

		var sentDelta, recvDelta uint64

		// Aggregate deltas for matching interface type
		for ifaceName, currStats := range curr.interfaces {
			if interfaceTypes[ifaceName] != ifaceType {
				continue
			}

			if prevStats, ok := prev.interfaces[ifaceName]; ok {
				sentDelta += currStats.bytesSent - prevStats.bytesSent
				recvDelta += currStats.bytesRecv - prevStats.bytesRecv
			}
		}

		// Calculate rates in Mbps
		sendMbps := float64(sentDelta) * 8 / duration / 1_000_000
		recvMbps := float64(recvDelta) * 8 / duration / 1_000_000

		totalSendMbps += sendMbps
		totalRecvMbps += recvMbps
		rateCount++
	}

	// Average the rates
	avgSendMbps := 0.0
	avgRecvMbps := 0.0
	if rateCount > 0 {
		avgSendMbps = totalSendMbps / float64(rateCount)
		avgRecvMbps = totalRecvMbps / float64(rateCount)
	}

	return &models.NetworkAggregateMetrics{
		SendMbps:    avgSendMbps,
		RecvMbps:    avgRecvMbps,
		TotalSentGB: float64(totalSentBytes) / 1_000_000_000,
		TotalRecvGB: float64(totalRecvBytes) / 1_000_000_000,
	}, nil
}

// CollectHealth aggregates network health statistics
func (n *NetworkCollector) CollectHealth(ctx context.Context) (*models.NetworkHealthMetrics, error) {
	// Get latest sample
	n.mu.Lock()
	if len(n.samples) == 0 {
		n.mu.Unlock()
		return &models.NetworkHealthMetrics{}, nil
	}
	lastSample := n.samples[len(n.samples)-1]
	n.mu.Unlock()

	var errorsIn, errorsOut, dropsIn, dropsOut uint64

	for _, stats := range lastSample.interfaces {
		errorsIn += stats.errorsIn
		errorsOut += stats.errorsOut
		dropsIn += stats.dropsIn
		dropsOut += stats.dropsOut
	}

	return &models.NetworkHealthMetrics{
		ErrorsIn:  errorsIn,
		ErrorsOut: errorsOut,
		DropsIn:   dropsIn,
		DropsOut:  dropsOut,
	}, nil
}

// isPrivateIP checks if an IP is in private address space
func isPrivateIP(ip net.IP) bool {
	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fd00::/8", // IPv6 ULA
	}

	for _, block := range privateBlocks {
		_, subnet, _ := net.ParseCIDR(block)
		if subnet != nil && subnet.Contains(ip) {
			return true
		}
	}

	return false
}
