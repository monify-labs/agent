# Monify Agent

A lightweight, high-performance monitoring agent for Linux servers. Collects system metrics and sends them to Monify Cloud for real-time monitoring and alerting.

## Features

- **CPU Monitoring**: Usage percentage, load averages (1m, 5m, 15m)
- **Memory Monitoring**: Total, used, free, available, cached, buffers
- **Disk Monitoring**: Space usage, I/O rates (read/write MB/s, IOPS)
- **Network Monitoring**: Public/private bandwidth, errors, drops
- **System Info**: OS, kernel, virtualization, cloud provider detection
- **Low Resource Usage**: ~20MB RAM, <1% CPU
- **Secure**: TLS encryption, token-based authentication
- **Easy Deployment**: Single binary, systemd integration

## Quick Install

One command to install and start monitoring:

```bash
curl -sSL https://monify.cloud/install.sh | sudo bash -s -- YOUR_TOKEN
```

Replace `YOUR_TOKEN` with your server token from [dash.monify.cloud](https://dash.monify.cloud).

That's it! The agent will be installed, configured, and started automatically.

## Requirements

- **OS**: Linux (Ubuntu, Debian, CentOS, RHEL, Amazon Linux, etc.)
- **Architecture**: amd64 (x86_64) or arm64 (aarch64)
- **Init System**: systemd

## Usage

```bash
monify [options]
```

### Options

| Option      | Description                                     |
| ----------- | ----------------------------------------------- |
| `--token`   | Authentication token (or set MONIFY_TOKEN)      |
| `--url`     | Server URL (or set MONIFY_SERVER_URL)           |
| `--debug`   | Enable debug logging (or set MONIFY_DEBUG=true) |
| `--version` | Show version information                        |
| `--help`    | Show help message                               |

### Examples

```bash
# Run with token from command line
monify --token=YOUR_TOKEN

# Run with debug mode
monify --token=YOUR_TOKEN --debug

# Run with custom server URL
monify --token=YOUR_TOKEN --url=https://custom-api.example.com

# Show version
monify --version

# Show help
monify --help
```

## Configuration

The agent reads configuration in this order (later overrides earlier):

1. **Config file**: `/etc/monify/env`
2. **Environment variables**
3. **Command line flags**

### Config file example (`/etc/monify/env`)

```bash
# Required: Your server token from https://dash.monify.cloud
MONIFY_TOKEN=your_server_token_here

# Optional: Custom server URL
MONIFY_SERVER_URL=https://api.monify.cloud/v1/agent/metrics

# Optional: Enable debug logging
MONIFY_DEBUG=false
```

### Environment variables

```bash
# Run with environment variables
MONIFY_TOKEN=xxx monify
MONIFY_TOKEN=xxx MONIFY_DEBUG=true monify
```

## Systemd Service

The agent runs as a systemd service:

```bash
# Start/stop/restart
sudo systemctl start monify
sudo systemctl stop monify
sudo systemctl restart monify

# View logs
sudo journalctl -u monify -f

# Check status
sudo systemctl status monify
```

## Development

### Prerequisites

- Go 1.22+
- Make

### Building

```bash
# Build for current architecture
make build

# Build for all platforms
make build-all

# Build for specific architecture
make build GOARCH=arm64
```

### Running Locally

```bash
# Development mode with debug logging
make dev

# Or manually
MONIFY_TOKEN=your_token MONIFY_DEBUG=true go run ./cmd/monify
```

### Project Structure

```
.
├── cmd/
│   └── monify/          # Entry point
│       └── main.go
├── internal/
│   ├── agent/           # Agent core
│   ├── config/          # Configuration
│   ├── metrics/         # Metric collectors
│   │   ├── dynamic/     # Frequently changing metrics
│   │   └── static/      # Rarely changing metrics
│   └── sender/          # HTTP sender
├── pkg/
│   └── models/          # Data models
├── scripts/
│   ├── install.sh       # Installation script
│   └── uninstall.sh     # Uninstallation script
├── Makefile
└── README.md
```

## Update

### Re-run install script

```bash
curl -sSL https://monify.cloud/install.sh | sudo bash
```

If already installed, the script will automatically use your existing token.

### Install with new token

```bash
curl -sSL https://monify.cloud/install.sh | sudo bash -s -- NEW_TOKEN --force
```

Use `--force` when replacing an existing token to avoid accidental overwrites.

## Troubleshooting

### Check service status

```bash
sudo systemctl status monify
```

### View logs

```bash
# Last 50 lines
journalctl -u monify --no-pager -n 50

# Follow logs in real-time
journalctl -u monify -f
```

### Common issues

| Issue                         | Solution                                                                                      |
| ----------------------------- | --------------------------------------------------------------------------------------------- |
| Service stopped (auth failed) | Token invalid. Re-install with new token using `--force`                                      |
| Token not configured          | Install with token: `curl -sSL https://monify.cloud/install.sh \| sudo bash -s -- YOUR_TOKEN` |
| Service won't start           | Check logs: `journalctl -u monify --no-pager -n 20`                                           |
| Agent using too much CPU      | Restart: `sudo systemctl restart monify`                                                      |

## Uninstall

```bash
curl -sSL https://monify.cloud/uninstall.sh | sudo bash
```

Or manually:

```bash
sudo systemctl stop monify
sudo systemctl disable monify
sudo rm -f /usr/local/bin/monify
sudo rm -f /etc/systemd/system/monify.service
sudo rm -rf /etc/monify
sudo rm -rf /var/log/monify
sudo systemctl daemon-reload
```

## Metrics Collected

### Static Metrics (sent on startup, then hourly)

| Metric            | Description                             |
| ----------------- | --------------------------------------- |
| Platform          | OS distribution (ubuntu, centos, etc.)  |
| Platform Version  | Distribution version                    |
| Kernel Version    | Linux kernel version                    |
| Architecture      | CPU architecture (amd64, arm64)         |
| Virtualization    | Virtualization type (kvm, docker, etc.) |
| CPU Model         | CPU model name                          |
| CPU Cores/Threads | Physical cores and logical processors   |
| Total Memory      | Total RAM                               |
| Internal IPs      | Private IP addresses                    |
| Public IP         | Public-facing IP                        |
| Cloud Region      | AWS/GCP/Azure region (if applicable)    |
| Instance Type     | Cloud instance type (if applicable)     |
| Disk Inventory    | Mounted filesystems                     |

### Dynamic Metrics (sent every 15s)

| Metric          | Description                             |
| --------------- | --------------------------------------- |
| CPU Usage       | Overall CPU usage percentage            |
| Load Average    | 1m, 5m, 15m load averages               |
| Memory          | Used, free, available, cached, buffers  |
| Swap            | Swap usage                              |
| Disk Space      | Total, used, free across all partitions |
| Disk I/O        | Read/write MB/s and IOPS                |
| Network Public  | Public interface bandwidth              |
| Network Private | Private interface bandwidth             |
| Network Health  | Errors and drops                        |
| System          | Uptime, boot time, process count        |

## Security

- All data is transmitted over HTTPS
- Token-based authentication
- Minimal privileges (requires root only for some metrics)
- No sensitive data collection (no file contents, no user data)
- Systemd hardening (NoNewPrivileges, ProtectSystem, etc.)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- Issues: https://github.com/monify-labs/agent/issues
- Email: support@monify.cloud
