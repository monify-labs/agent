package static

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	gopsutilNet "github.com/shirou/gopsutil/v4/net"
)

// NetworkInfo contains network configuration information
type NetworkInfo struct {
	InternalIPs []string
	PublicIP    string
	Hostname    string
	FQDN        string
	Timezone    string
}

// NetworkInfoCollector handles network information collection with public IP caching
type NetworkInfoCollector struct {
	mu            sync.RWMutex
	publicIPCache string
	cacheTime     time.Time
	cacheDuration time.Duration
}

// NewNetworkInfoCollector creates a new NetworkInfoCollector with 5-minute cache
func NewNetworkInfoCollector() *NetworkInfoCollector {
	return &NetworkInfoCollector{
		cacheDuration: 5 * time.Minute,
	}
}

// Collect gathers network configuration information
func (n *NetworkInfoCollector) Collect(ctx context.Context) (*NetworkInfo, error) {
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get timezone
	timezone := getTimezone()

	// Get internal IPs
	internalIPs := n.getInternalIPs(ctx)

	// Get public IP (with caching)
	publicIP := n.getPublicIP(ctx)

	// Get FQDN (best effort)
	fqdn := n.getFQDN(hostname)

	return &NetworkInfo{
		InternalIPs: internalIPs,
		PublicIP:    publicIP,
		Hostname:    hostname,
		FQDN:        fqdn,
		Timezone:    timezone,
	}, nil
}

// getInternalIPs retrieves all internal IP addresses
func (n *NetworkInfoCollector) getInternalIPs(ctx context.Context) []string {
	var ips []string

	interfaces, err := gopsutilNet.InterfacesWithContext(ctx)
	if err != nil {
		return ips
	}

	for _, iface := range interfaces {
		for _, addr := range iface.Addrs {
			// Parse IP from CIDR notation
			ip, _, err := net.ParseCIDR(addr.Addr)
			if err != nil {
				// Try parsing as plain IP
				ip = net.ParseIP(addr.Addr)
			}

			if ip != nil && !ip.IsLoopback() && !ip.IsUnspecified() {
				// Only include private IPs
				if isPrivateIP(ip) {
					ips = append(ips, ip.String())
				}
			}
		}
	}

	return ips
}

// getPublicIP retrieves the public IP address with caching
func (n *NetworkInfoCollector) getPublicIP(ctx context.Context) string {
	// Check cache first
	n.mu.RLock()
	if time.Since(n.cacheTime) < n.cacheDuration && n.publicIPCache != "" {
		cachedIP := n.publicIPCache
		n.mu.RUnlock()
		return cachedIP
	}
	n.mu.RUnlock()

	// Fetch new public IP
	publicIP := n.fetchPublicIP(ctx)

	// Update cache
	n.mu.Lock()
	n.publicIPCache = publicIP
	n.cacheTime = time.Now()
	n.mu.Unlock()

	return publicIP
}

// fetchPublicIP queries external service for public IP
func (n *NetworkInfoCollector) fetchPublicIP(ctx context.Context) string {
	endpoints := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ifconfig.me",
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	for _, endpoint := range endpoints {
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			continue
		}

		ip := strings.TrimSpace(string(body))
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	return ""
}

// getFQDN attempts to get fully qualified domain name
func (n *NetworkInfoCollector) getFQDN(hostname string) string {
	// Try reverse DNS lookup
	addrs, err := net.LookupAddr(hostname)
	if err == nil && len(addrs) > 0 {
		return strings.TrimSuffix(addrs[0], ".")
	}

	// Fallback to hostname
	return hostname
}

// getTimezone returns the system timezone
func getTimezone() string {
	zone, _ := time.Now().Zone()
	return zone
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
