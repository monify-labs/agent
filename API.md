# API Documentation

## Endpoint

```
POST /v1/agent/metrics
```

## Authentication

```http
Authorization: Bearer {agentToken}
```

## Request

### Headers

```http
Content-Type: application/json
Content-Encoding: gzip
Authorization: Bearer {token}
```

### Payload

```json
{
  "hostname": "server-01",
  "timestamp": "2024-12-11T01:00:00Z",
  "static_info": {
    "platform": "ubuntu",
    "platform_version": "22.04",
    "os": "linux",
    "arch": "amd64",
    "cpu_model": "Intel Xeon",
    "cpu_cores": 4,
    "total_memory": 16777216000,
    "internal_ips": ["192.168.1.100"],
    "public_ip": "203.0.113.42"
  },
  "metrics": {
    "cpu": {
      "usage_percent": 45.2,
      "per_core": [42.1, 48.3, 44.5, 46.0],
      "load_avg": [1.5, 1.8, 2.1]
    },
    "memory": {
      "total": 16777216000,
      "used": 8388608000,
      "used_percent": 50.0,
      "cached": 4194304000,
      "swap_total": 2147483648,
      "swap_used": 0
    },
    "disk": [{
      "device": "/dev/sda1",
      "mount": "/",
      "fstype": "ext4",
      "total": 107374182400,
      "used": 53687091200,
      "used_percent": 50.0
    }],
    "network": [{
      "interface": "eth0",
      "bytes_sent": 1073741824,
      "bytes_recv": 2147483648,
      "packets_sent": 1000000,
      "packets_recv": 2000000
    }],
    "system": {
      "uptime": 864000,
      "boot_time": 1702080000,
      "process_count": 150
    }
  }
}
```

## Response

### Success

```json
{
  "status": "success",
  "message": "Metrics received",
  "commands": []
}
```

### With Commands

```json
{
  "status": "success",
  "message": "Metrics received",
  "commands": [
    {
      "command": "update_config",
      "params": { "collection_interval": "60s" }
    }
  ]
}
```

## Server Commands

### update_config
```json
{ "command": "update_config", "params": { "collection_interval": "60s" } }
```

### refresh
```json
{ "command": "refresh" }
```

### scan_ports
```json
{
  "command": "scan_ports",
  "params": { "target": "192.168.1.1", "ports": [22, 80, 443] }
}
```

### stop
```json
{ "command": "stop", "params": { "reason": "Server disabled" } }
```

### uninstall
```json
{ "command": "uninstall", "params": { "reason": "Server deleted" } }
```

## Data Types

### Static Info (sent once)

| Field | Type | Description |
|-------|------|-------------|
| platform | string | OS distribution |
| platform_version | string | OS version |
| os | string | Operating system |
| arch | string | Architecture |
| cpu_model | string | CPU model |
| cpu_cores | int | CPU cores |
| total_memory | uint64 | Total RAM (bytes) |
| internal_ips | []string | Internal IPs |
| public_ip | string | Public IP |

### Dynamic Metrics (sent every interval)

**CPU**
- `usage_percent` (float64): Overall usage
- `per_core` ([]float64): Per-core usage
- `load_avg` ([]float64): Load averages [1m, 5m, 15m]

**Memory**
- `total` (uint64): Total RAM
- `used` (uint64): Used RAM
- `used_percent` (float64): Usage %
- `cached` (uint64): Cached memory
- `swap_total` (uint64): Total swap
- `swap_used` (uint64): Used swap

**Disk** (array)
- `device` (string): Device name
- `mount` (string): Mount point
- `fstype` (string): Filesystem type
- `total` (uint64): Total space
- `used` (uint64): Used space
- `used_percent` (float64): Usage %

**Network** (array)
- `interface` (string): Interface name
- `bytes_sent` (uint64): Bytes sent
- `bytes_recv` (uint64): Bytes received
- `packets_sent` (uint64): Packets sent
- `packets_recv` (uint64): Packets received

**System**
- `uptime` (uint64): Uptime (seconds)
- `boot_time` (uint64): Boot time (unix timestamp)
- `process_count` (uint64): Running processes

## Error Responses

### 401 Unauthorized
```json
{ "statusCode": 401, "message": "Invalid token" }
```

### 400 Bad Request
```json
{ "statusCode": 400, "message": "Validation failed" }
```

### 429 Too Many Requests
```json
{ "statusCode": 429, "message": "Rate limit exceeded" }
```

## Notes

- All payloads are gzip-compressed
- Default collection interval: 30 seconds
- Rate limit: 2 requests/second per agent
- Compression reduces bandwidth by ~70-80%

---

**Backend Integration**: See `/backend/docs/AGENT_INTEGRATION.md`
