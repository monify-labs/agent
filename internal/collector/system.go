package collector

import (
	"context"
	"io"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// SystemCollector collects system metrics
type SystemCollector struct {
	*BaseCollector
	publicIPCache  string
	lastIPCheck    time.Time
	staticCache    *models.StaticMetrics
	staticCacheSet bool
}

// NewSystemCollector creates a new system collector
func NewSystemCollector(enabled bool) *SystemCollector {
	return &SystemCollector{
		BaseCollector: NewBaseCollector("system", enabled),
	}
}

// CollectStatic collects static system information (cached)
func (s *SystemCollector) CollectStatic(ctx context.Context) (*models.StaticMetrics, error) {
	if !s.Enabled() {
		return nil, nil
	}

	// Return cached static info if already collected
	if s.staticCacheSet {
		return s.staticCache, nil
	}

	static := &models.StaticMetrics{}

	// Get host info
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	static.Platform = hostInfo.Platform
	static.PlatformFamily = hostInfo.PlatformFamily
	static.PlatformVersion = hostInfo.PlatformVersion
	static.OS = hostInfo.OS
	static.KernelVersion = hostInfo.KernelVersion
	static.KernelArch = hostInfo.KernelArch
	static.Virtualization = hostInfo.VirtualizationSystem
	static.HostID = hostInfo.HostID
	static.Arch = runtime.GOARCH
	static.Hostname = hostInfo.Hostname

	// Get CPU info
	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err == nil && len(cpuInfo) > 0 {
		static.CPUModel = cpuInfo[0].ModelName
		static.CPUCores = int(cpuInfo[0].Cores)

		// Count total threads
		totalThreads := 0
		for _, info := range cpuInfo {
			totalThreads += int(info.Cores)
		}
		static.CPUThreads = totalThreads
	}

	// Get total memory
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		static.TotalMemory = memInfo.Total
	}

	// Get internal IPs
	internalIPs, err := getInternalIPs()
	if err == nil {
		static.InternalIPs = internalIPs
	}

	// Get public IP (cached for 5 minutes)
	publicIP := s.getPublicIP(ctx)
	if publicIP != "" {
		static.PublicIP = publicIP
	}

	// Get FQDN
	if fqdn, err := net.LookupAddr(publicIP); err == nil && len(fqdn) > 0 {
		static.FQDN = fqdn[0]
	}

	// Get timezone
	_, offset := time.Now().Zone()
	static.Timezone = time.Now().Location().String()
	if static.Timezone == "Local" {
		// Get timezone from offset
		hours := offset / 3600
		if hours >= 0 {
			static.Timezone = time.FixedZone("UTC+", offset).String()
		} else {
			static.Timezone = time.FixedZone("UTC-", -offset).String()
		}
	}

	// Try to detect cloud provider and region
	s.detectCloudInfo(ctx, static)

	// Cache the static info
	s.staticCache = static
	s.staticCacheSet = true

	return static, nil
}

// CollectDynamic collects dynamic system metrics
func (s *SystemCollector) CollectDynamic(ctx context.Context) (*models.SystemDynamic, error) {
	if !s.Enabled() {
		return nil, nil
	}

	dynamic := &models.SystemDynamic{}

	// Get host info
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	dynamic.Uptime = hostInfo.Uptime
	dynamic.BootTime = hostInfo.BootTime

	// Get process count
	processes, err := process.ProcessesWithContext(ctx)
	if err == nil {
		dynamic.ProcessCount = uint64(len(processes))
	}

	return dynamic, nil
}

// Collect collects system metrics (deprecated - use CollectStatic and CollectDynamic)
func (s *SystemCollector) Collect(ctx context.Context) (interface{}, error) {
	if !s.Enabled() {
		return nil, nil
	}

	metrics := &models.SystemMetrics{}

	// Get host info
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	metrics.Uptime = hostInfo.Uptime
	metrics.BootTime = hostInfo.BootTime
	metrics.Platform = hostInfo.Platform
	metrics.PlatformFamily = hostInfo.PlatformFamily
	metrics.PlatformVersion = hostInfo.PlatformVersion
	metrics.OS = hostInfo.OS
	metrics.KernelVersion = hostInfo.KernelVersion
	metrics.KernelArch = hostInfo.KernelArch
	metrics.Virtualization = hostInfo.VirtualizationSystem
	metrics.HostID = hostInfo.HostID
	metrics.Arch = runtime.GOARCH

	// Get process count
	processes, err := process.ProcessesWithContext(ctx)
	if err == nil {
		metrics.ProcessCount = uint64(len(processes))
	}

	// Get CPU info
	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err == nil && len(cpuInfo) > 0 {
		metrics.CPUModel = cpuInfo[0].ModelName
		metrics.CPUCores = int(cpuInfo[0].Cores)

		// Count total threads
		totalThreads := 0
		for _, info := range cpuInfo {
			totalThreads += int(info.Cores)
		}
		metrics.CPUThreads = totalThreads
	}

	// Get total memory
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		metrics.TotalMemory = memInfo.Total
	}

	// Get internal IPs
	internalIPs, err := getInternalIPs()
	if err == nil {
		metrics.InternalIPs = internalIPs
	}

	// Get public IP (cached for 5 minutes)
	publicIP := s.getPublicIP(ctx)
	if publicIP != "" {
		metrics.PublicIP = publicIP
	}

	// Get hostname and FQDN
	metrics.Hostname = hostInfo.Hostname
	if fqdn, err := net.LookupAddr(publicIP); err == nil && len(fqdn) > 0 {
		metrics.FQDN = fqdn[0]
	}

	// Get timezone
	_, offset := time.Now().Zone()
	metrics.Timezone = time.Now().Location().String()
	if metrics.Timezone == "Local" {
		// Get timezone from offset
		hours := offset / 3600
		if hours >= 0 {
			metrics.Timezone = time.FixedZone("UTC+", offset).String()
		} else {
			metrics.Timezone = time.FixedZone("UTC-", -offset).String()
		}
	}

	// Try to detect cloud provider and region
	s.detectCloudInfoForSystemMetrics(ctx, metrics)

	return metrics, nil
}

// getInternalIPs gets all internal IP addresses
func getInternalIPs() ([]string, error) {
	var ips []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		// Skip down interfaces and loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip loopback and link-local addresses
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}

			// Only include IPv4 and global IPv6 addresses
			if ip.To4() != nil || ip.IsGlobalUnicast() {
				ips = append(ips, ip.String())
			}
		}
	}

	return ips, nil
}

// getPublicIP gets the public IP address (with caching)
func (s *SystemCollector) getPublicIP(ctx context.Context) string {
	// Use cache if available and less than 5 minutes old
	if s.publicIPCache != "" && time.Since(s.lastIPCheck) < 5*time.Minute {
		return s.publicIPCache
	}

	// Try multiple services in case one is down
	services := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ifconfig.me/ip",
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, service := range services {
		req, err := http.NewRequestWithContext(ctx, "GET", service, nil)
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
		if ip != "" && net.ParseIP(ip) != nil {
			s.publicIPCache = ip
			s.lastIPCheck = time.Now()
			return ip
		}
	}

	return s.publicIPCache
}

// detectCloudInfo tries to detect cloud provider and region from metadata services
func (s *SystemCollector) detectCloudInfo(ctx context.Context, static *models.StaticMetrics) {
	// AWS EC2 metadata
	if region, instanceType := s.detectAWS(ctx); region != "" {
		static.Region = region
		static.InstanceType = instanceType
		return
	}

	// Google Cloud metadata
	if region, instanceType := s.detectGCP(ctx); region != "" {
		static.Region = region
		static.InstanceType = instanceType
		return
	}

	// Azure metadata
	if region, instanceType := s.detectAzure(ctx); region != "" {
		static.Region = region
		static.InstanceType = instanceType
		return
	}

	// DigitalOcean metadata
	if region := s.detectDigitalOcean(ctx); region != "" {
		static.Region = region
		return
	}
}

// detectCloudInfoForSystemMetrics is for backward compatibility
func (s *SystemCollector) detectCloudInfoForSystemMetrics(ctx context.Context, metrics *models.SystemMetrics) {
	// AWS EC2 metadata
	if region, instanceType := s.detectAWS(ctx); region != "" {
		metrics.Region = region
		metrics.InstanceType = instanceType
		return
	}

	// Google Cloud metadata
	if region, instanceType := s.detectGCP(ctx); region != "" {
		metrics.Region = region
		metrics.InstanceType = instanceType
		return
	}

	// Azure metadata
	if region, instanceType := s.detectAzure(ctx); region != "" {
		metrics.Region = region
		metrics.InstanceType = instanceType
		return
	}

	// DigitalOcean metadata
	if region := s.detectDigitalOcean(ctx); region != "" {
		metrics.Region = region
		return
	}
}

// detectAWS detects AWS EC2 instance metadata
func (s *SystemCollector) detectAWS(ctx context.Context) (region, instanceType string) {
	client := &http.Client{Timeout: 2 * time.Second}

	// Get region
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://169.254.169.254/latest/meta-data/placement/region", nil)
	if resp, err := client.Do(req); err == nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		region = strings.TrimSpace(string(body))
	}

	// Get instance type
	req, _ = http.NewRequestWithContext(ctx, "GET", "http://169.254.169.254/latest/meta-data/instance-type", nil)
	if resp, err := client.Do(req); err == nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		instanceType = strings.TrimSpace(string(body))
	}

	return
}

// detectGCP detects Google Cloud instance metadata
func (s *SystemCollector) detectGCP(ctx context.Context) (region, instanceType string) {
	client := &http.Client{Timeout: 2 * time.Second}

	// Get zone (GCP doesn't have region directly)
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://metadata.google.internal/computeMetadata/v1/instance/zone", nil)
	req.Header.Set("Metadata-Flavor", "Google")
	if resp, err := client.Do(req); err == nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		zone := strings.TrimSpace(string(body))
		// Extract region from zone (e.g., us-central1-a -> us-central1)
		if parts := strings.Split(zone, "/"); len(parts) > 0 {
			zoneName := parts[len(parts)-1]
			if idx := strings.LastIndex(zoneName, "-"); idx > 0 {
				region = zoneName[:idx]
			}
		}
	}

	// Get machine type
	req, _ = http.NewRequestWithContext(ctx, "GET", "http://metadata.google.internal/computeMetadata/v1/instance/machine-type", nil)
	req.Header.Set("Metadata-Flavor", "Google")
	if resp, err := client.Do(req); err == nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		machineType := strings.TrimSpace(string(body))
		if parts := strings.Split(machineType, "/"); len(parts) > 0 {
			instanceType = parts[len(parts)-1]
		}
	}

	return
}

// detectAzure detects Azure instance metadata
func (s *SystemCollector) detectAzure(ctx context.Context) (region, instanceType string) {
	client := &http.Client{Timeout: 2 * time.Second}

	req, _ := http.NewRequestWithContext(ctx, "GET", "http://169.254.169.254/metadata/instance?api-version=2021-02-01", nil)
	req.Header.Set("Metadata", "true")
	if resp, err := client.Do(req); err == nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		data := string(body)
		// Simple parsing (would need proper JSON parsing in production)
		if strings.Contains(data, "location") {
			// Extract region from JSON response
			parts := strings.Split(data, "\"location\":\"")
			if len(parts) > 1 {
				region = strings.Split(parts[1], "\"")[0]
			}
		}
		if strings.Contains(data, "vmSize") {
			parts := strings.Split(data, "\"vmSize\":\"")
			if len(parts) > 1 {
				instanceType = strings.Split(parts[1], "\"")[0]
			}
		}
	}

	return
}

// detectDigitalOcean detects DigitalOcean metadata
func (s *SystemCollector) detectDigitalOcean(ctx context.Context) string {
	client := &http.Client{Timeout: 2 * time.Second}

	req, _ := http.NewRequestWithContext(ctx, "GET", "http://169.254.169.254/metadata/v1/region", nil)
	if resp, err := client.Do(req); err == nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return strings.TrimSpace(string(body))
	}

	return ""
}
