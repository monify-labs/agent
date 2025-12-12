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
	CPU            *CPUMetrics              `json:"cpu,omitempty"`
	Memory         *MemoryMetrics           `json:"memory,omitempty"`
	Swap           *SwapMetrics             `json:"swap,omitempty"`
	DiskSpace      *DiskSpaceMetrics        `json:"disk_space,omitempty"`
	DiskIO         *DiskIOMetrics           `json:"disk_io,omitempty"`
	NetworkPublic  *NetworkAggregateMetrics `json:"network_public,omitempty"`
	NetworkPrivate *NetworkAggregateMetrics `json:"network_private,omitempty"`
	NetworkHealth  *NetworkHealthMetrics    `json:"network_health,omitempty"`
	System         *SystemDynamic           `json:"system,omitempty"` // Dynamic system metrics only

	// Deprecated - kept for backward compatibility
	Disk    []DiskMetrics    `json:"disk,omitempty"`
	Network []NetworkMetrics `json:"network,omitempty"`
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
	LoadAvg1m    float64   `json:"load_avg_1m"`
	LoadAvg5m    float64   `json:"load_avg_5m"`
	LoadAvg15m   float64   `json:"load_avg_15m"`

	// Deprecated - kept for backward compatibility
	PerCore []float64 `json:"per_core,omitempty"`
	LoadAvg []float64 `json:"load_avg,omitempty"` // 1, 5, 15 minute averages
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

	// Deprecated - kept for backward compatibility
	SwapTotal uint64 `json:"swap_total,omitempty"`
	SwapUsed  uint64 `json:"swap_used,omitempty"`
	SwapFree  uint64 `json:"swap_free,omitempty"`
}

// SwapMetrics contains swap memory usage information
type SwapMetrics struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskMetrics contains disk usage information for a single mount point
type DiskMetrics struct {
	Device      string  `json:"device"`
	MountPoint  string  `json:"mount"`
	FSType      string  `json:"fstype"`

	// Disk space usage
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`

	// Inode usage
	InodesTotal uint64  `json:"inodes_total"`
	InodesUsed  uint64  `json:"inodes_used"`
	InodesFree  uint64  `json:"inodes_free"`

	// Cumulative I/O counters (since boot)
	ReadBytes   uint64  `json:"read_bytes"`
	WriteBytes  uint64  `json:"write_bytes"`
	ReadCount   uint64  `json:"read_count"`
	WriteCount  uint64  `json:"write_count"`

	// I/O rates (bytes per second)
	ReadRate     float64 `json:"read_rate"`      // Read bandwidth in bytes/s
	WriteRate    float64 `json:"write_rate"`     // Write bandwidth in bytes/s
	ReadRateMBps float64 `json:"read_rate_mbps"` // Read in MB/s
	WriteRateMBps float64 `json:"write_rate_mbps"` // Write in MB/s

	// IOPS (I/O operations per second)
	ReadIOPS  float64 `json:"read_iops"`  // Read operations per second
	WriteIOPS float64 `json:"write_iops"` // Write operations per second
}

// DiskSpaceMetrics contains aggregated disk space usage across all partitions
type DiskSpaceMetrics struct {
	Total       uint64  `json:"total"`        // Total disk space in bytes
	Used        uint64  `json:"used"`         // Used disk space in bytes
	Free        uint64  `json:"free"`         // Free disk space in bytes
	UsedPercent float64 `json:"used_percent"` // Usage percentage
}

// DiskIOMetrics contains aggregated disk I/O metrics across all devices
type DiskIOMetrics struct {
	ReadMBps  float64 `json:"read_mbps"`  // Aggregate read bandwidth in MB/s
	WriteMBps float64 `json:"write_mbps"` // Aggregate write bandwidth in MB/s
	ReadIOPS  float64 `json:"read_iops"`  // Aggregate read IOPS
	WriteIOPS float64 `json:"write_iops"` // Aggregate write IOPS
}

// NetworkMetrics contains network interface statistics (deprecated - kept for backward compatibility)
type NetworkMetrics struct {
	Interface   string `json:"interface"`
	Type        string `json:"type"` // "public" or "private"

	// Cumulative counters (since boot)
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`

	// Bandwidth rates (bytes per second)
	SendRate    float64 `json:"send_rate"`    // Outbound bandwidth in bytes/s
	RecvRate    float64 `json:"recv_rate"`    // Inbound bandwidth in bytes/s
	SendRateMbps float64 `json:"send_rate_mbps"` // Outbound in Mbps
	RecvRateMbps float64 `json:"recv_rate_mbps"` // Inbound in Mbps

	// Error and drop statistics
	ErrorsIn    uint64 `json:"errors_in"`
	ErrorsOut   uint64 `json:"errors_out"`
	DropsIn     uint64 `json:"drops_in"`
	DropsOut    uint64 `json:"drops_out"`
}

// NetworkAggregateMetrics contains aggregated network bandwidth by type (public/private)
type NetworkAggregateMetrics struct {
	SendMbps    float64 `json:"send_mbps"`      // Aggregate outbound bandwidth in Mbps
	RecvMbps    float64 `json:"recv_mbps"`      // Aggregate inbound bandwidth in Mbps
	TotalSentGB float64 `json:"total_sent_gb"`  // Cumulative sent in GB
	TotalRecvGB float64 `json:"total_recv_gb"`  // Cumulative received in GB
}

// NetworkHealthMetrics contains aggregated network health statistics
type NetworkHealthMetrics struct {
	ErrorsIn  uint64 `json:"errors_in"`  // Total inbound errors
	ErrorsOut uint64 `json:"errors_out"` // Total outbound errors
	DropsIn   uint64 `json:"drops_in"`   // Total inbound drops
	DropsOut  uint64 `json:"drops_out"`  // Total outbound drops
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
