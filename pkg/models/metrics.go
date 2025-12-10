package models

import "time"

// MetricPayload represents the complete payload sent to the server
// Authentication is done via token in Authorization header
type MetricPayload struct {
	Hostname   string          `json:"hostname"`
	Timestamp  time.Time       `json:"timestamp"`
	StaticInfo *StaticMetrics  `json:"static_info,omitempty"` // Only sent when changed or first time
	Metrics    *DynamicMetrics `json:"metrics"`               // Always sent
}

// StaticMetrics contains rarely-changing system information
type StaticMetrics struct {
	// System Info
	Platform        string `json:"platform"`        // ubuntu, centos, etc.
	PlatformFamily  string `json:"platform_family"` // debian, rhel, etc.
	PlatformVersion string `json:"platform_version"`
	OS              string `json:"os"`             // linux, darwin, windows
	Arch            string `json:"arch"`           // amd64, arm64
	KernelVersion   string `json:"kernel_version"` // 5.15.0-91-generic
	KernelArch      string `json:"kernel_arch"`    // x86_64
	Virtualization  string `json:"virtualization"` // kvm, docker, vmware, etc.
	HostID          string `json:"host_id"`

	// Network Info
	InternalIPs []string `json:"internal_ips"`        // All internal IPs
	PublicIP    string   `json:"public_ip,omitempty"` // Public facing IP
	Hostname    string   `json:"hostname"`            // Server hostname
	FQDN        string   `json:"fqdn,omitempty"`      // Fully qualified domain name

	// Hardware Info
	CPUModel    string `json:"cpu_model"`    // Intel(R) Xeon(R) CPU...
	CPUCores    int    `json:"cpu_cores"`    // Physical cores
	CPUThreads  int    `json:"cpu_threads"`  // Logical processors
	TotalMemory uint64 `json:"total_memory"` // Total RAM in bytes

	// Additional Info
	Timezone     string `json:"timezone,omitempty"`      // Server timezone
	Region       string `json:"region,omitempty"`        // Cloud region (if detectable)
	InstanceType string `json:"instance_type,omitempty"` // EC2 type, etc.
}

// DynamicMetrics contains frequently-changing metrics
type DynamicMetrics struct {
	CPU     *CPUMetrics      `json:"cpu,omitempty"`
	Memory  *MemoryMetrics   `json:"memory,omitempty"`
	Disk    []DiskMetrics    `json:"disk,omitempty"`
	Network []NetworkMetrics `json:"network,omitempty"`
	System  *SystemDynamic   `json:"system,omitempty"` // Dynamic system metrics only
}

// SystemDynamic contains frequently-changing system metrics
type SystemDynamic struct {
	Uptime       uint64 `json:"uptime"`        // seconds
	BootTime     uint64 `json:"boot_time"`     // Unix timestamp
	ProcessCount uint64 `json:"process_count"` // Number of running processes
}

// MetricsData contains all collected metrics (deprecated - kept for backward compatibility)
type MetricsData struct {
	CPU     *CPUMetrics      `json:"cpu,omitempty"`
	Memory  *MemoryMetrics   `json:"memory,omitempty"`
	Disk    []DiskMetrics    `json:"disk,omitempty"`
	Network []NetworkMetrics `json:"network,omitempty"`
	System  *SystemMetrics   `json:"system,omitempty"`
}

// CPUMetrics contains CPU usage information
type CPUMetrics struct {
	UsagePercent float64   `json:"usage_percent"`
	PerCore      []float64 `json:"per_core"`
	LoadAvg      []float64 `json:"load_avg"` // 1, 5, 15 minute averages
}

// MemoryMetrics contains memory usage information
type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"used_percent"`
	Cached      uint64  `json:"cached"`
	Buffers     uint64  `json:"buffers"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	SwapFree    uint64  `json:"swap_free"`
}

// DiskMetrics contains disk usage information for a single mount point
type DiskMetrics struct {
	Device      string  `json:"device"`
	MountPoint  string  `json:"mount"`
	FSType      string  `json:"fstype"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
	InodesTotal uint64  `json:"inodes_total"`
	InodesUsed  uint64  `json:"inodes_used"`
	InodesFree  uint64  `json:"inodes_free"`
	ReadBytes   uint64  `json:"read_bytes"`
	WriteBytes  uint64  `json:"write_bytes"`
	ReadCount   uint64  `json:"read_count"`
	WriteCount  uint64  `json:"write_count"`
}

// NetworkMetrics contains network interface statistics
type NetworkMetrics struct {
	Interface   string `json:"interface"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	ErrorsIn    uint64 `json:"errors_in"`
	ErrorsOut   uint64 `json:"errors_out"`
	DropsIn     uint64 `json:"drops_in"`
	DropsOut    uint64 `json:"drops_out"`
}

// SystemMetrics contains general system information
type SystemMetrics struct {
	// System Info
	Uptime          uint64 `json:"uptime"` // seconds
	BootTime        uint64 `json:"boot_time"`
	ProcessCount    uint64 `json:"process_count"`
	Platform        string `json:"platform"`        // ubuntu, centos, etc.
	PlatformFamily  string `json:"platform_family"` // debian, rhel, etc.
	PlatformVersion string `json:"platform_version"`
	OS              string `json:"os"`             // linux, darwin, windows
	Arch            string `json:"arch"`           // amd64, arm64
	KernelVersion   string `json:"kernel_version"` // 5.15.0-91-generic
	KernelArch      string `json:"kernel_arch"`    // x86_64
	Virtualization  string `json:"virtualization"` // kvm, docker, vmware, etc.
	HostID          string `json:"host_id"`

	// Network Info
	InternalIPs []string `json:"internal_ips"`        // All internal IPs
	PublicIP    string   `json:"public_ip,omitempty"` // Public facing IP
	Hostname    string   `json:"hostname"`            // Server hostname
	FQDN        string   `json:"fqdn,omitempty"`      // Fully qualified domain name

	// Hardware Info
	CPUModel    string `json:"cpu_model"`    // Intel(R) Xeon(R) CPU...
	CPUCores    int    `json:"cpu_cores"`    // Physical cores
	CPUThreads  int    `json:"cpu_threads"`  // Logical processors
	TotalMemory uint64 `json:"total_memory"` // Total RAM in bytes

	// Additional Info
	Timezone     string `json:"timezone,omitempty"`      // Server timezone
	Region       string `json:"region,omitempty"`        // Cloud region (if detectable)
	InstanceType string `json:"instance_type,omitempty"` // EC2 type, etc.
}

// PortScanRequest represents a request to scan ports
type PortScanRequest struct {
	Target    string     `json:"target"`               // IP address or hostname
	Ports     []int      `json:"ports"`                // List of ports to scan
	PortRange *PortRange `json:"port_range,omitempty"` // Alternative: range of ports
}

// PortRange represents a range of ports
type PortRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// PortScanResult represents the result of a port scan
type PortScanResult struct {
	Target    string        `json:"target"`
	Timestamp time.Time     `json:"timestamp"`
	OpenPorts []OpenPort    `json:"open_ports"`
	ScanTime  time.Duration `json:"scan_time_ms"`
}

// OpenPort represents an open port
type OpenPort struct {
	Port    int    `json:"port"`
	Service string `json:"service,omitempty"` // e.g., "http", "ssh"
}

// RefreshRequest represents a request to refresh metrics immediately
type RefreshRequest struct {
	Force bool `json:"force"` // Force collection even if recently collected
}

// AgentStatus represents the current status of the agent
type AgentStatus struct {
	Hostname       string    `json:"hostname"`
	Version        string    `json:"version"`
	Uptime         uint64    `json:"uptime"`
	LastCollection time.Time `json:"last_collection"`
	LastSend       time.Time `json:"last_send"`
	MetricsCount   uint64    `json:"metrics_count"`
	ErrorCount     uint64    `json:"error_count"`
	Status         string    `json:"status"` // "running", "stopped", "error"
}

// ServerCommand represents a command from server to agent
type ServerCommand struct {
	Command string         `json:"command"` // "update_config", "refresh", "scan_ports", "restart"
	Params  map[string]any `json:"params,omitempty"`
}

// ServerResponse represents the response from server after sending metrics
type ServerResponse struct {
	Status   string          `json:"status"` // "success", "error"
	Message  string          `json:"message,omitempty"`
	Commands []ServerCommand `json:"commands,omitempty"` // Commands for agent to execute
}
