package models

import "time"

// MetricPayload represents the complete payload sent to the server
// Authentication is done via token in Authorization header
type MetricPayload struct {
	Hostname       string          `json:"hostname"`
	Timestamp      time.Time       `json:"timestamp"`
	StaticMetrics  *StaticMetrics  `json:"static_info,omitempty"` // Only sent when changed or first time
	DynamicMetrics *DynamicMetrics `json:"metrics"`               // Always sent
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

	// Inventory
	Disks []DiskInventoryMetrics `json:"disks,omitempty"` // Disk/filesystem inventory
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
	System         *SystemMetrics           `json:"system,omitempty"`
}

// SystemMetrics contains frequently-changing system metrics
type SystemMetrics struct {
	Uptime       uint64 `json:"uptime"`        // seconds
	BootTime     uint64 `json:"boot_time"`     // Unix timestamp
	ProcessCount uint64 `json:"process_count"` // Number of running processes
}

// CPUMetrics contains CPU usage information
type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	LoadAvg1m    float64 `json:"load_avg_1m"`
	LoadAvg5m    float64 `json:"load_avg_5m"`
	LoadAvg15m   float64 `json:"load_avg_15m"`
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
}

// SwapMetrics contains swap memory usage information
type SwapMetrics struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskInventoryMetrics contains static disk/filesystem information
type DiskInventoryMetrics struct {
	Device      string `json:"device"`       // Device path (e.g., /dev/sda1)
	MountPoint  string `json:"mount"`        // Mount point (e.g., /)
	FSType      string `json:"fstype"`       // Filesystem type (e.g., ext4, xfs)
	Total       uint64 `json:"total"`        // Total capacity in bytes
	InodesTotal uint64 `json:"inodes_total"` // Total inodes
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

// NetworkAggregateMetrics contains aggregated network bandwidth by type (public/private)
type NetworkAggregateMetrics struct {
	SendMbps    float64 `json:"send_mbps"`     // Aggregate outbound bandwidth in Mbps
	RecvMbps    float64 `json:"recv_mbps"`     // Aggregate inbound bandwidth in Mbps
	TotalSentGB float64 `json:"total_sent_gb"` // Cumulative sent in GB
	TotalRecvGB float64 `json:"total_recv_gb"` // Cumulative received in GB
}

// NetworkHealthMetrics contains aggregated network health statistics
type NetworkHealthMetrics struct {
	ErrorsIn  uint64 `json:"errors_in"`  // Total inbound errors
	ErrorsOut uint64 `json:"errors_out"` // Total outbound errors
	DropsIn   uint64 `json:"drops_in"`   // Total inbound drops
	DropsOut  uint64 `json:"drops_out"`  // Total outbound drops
}

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
