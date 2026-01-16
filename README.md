# Monify Agent

Monify Agent is a lightweight, high-performance monitoring agent designed for Linux servers. It collects comprehensive system metrics and securely transmits them to [Monify Cloud](https://monify.cloud) for real-time monitoring, alerting, and visualization.

## Key Features

**Comprehensive Metrics Collection**

- CPU usage and load averages (1m, 5m, 15m)
- Memory utilization (used, free, available, cached, buffers)
- Disk space and I/O performance (read/write MB/s, IOPS)
- Network bandwidth for public and private interfaces
- System information including uptime and process count

**Cloud-Native Design**

- Automatic detection of AWS, GCP, and Azure environments
- Instance type and region identification
- Virtualization and container awareness

**Production Ready**

- Minimal resource footprint (~20MB RAM, <1% CPU)
- Secure communication over HTTPS with token authentication
- Runs as a systemd service with automatic restart
- Single binary deployment with no dependencies

## System Requirements

| Component        | Requirement                                                        |
| ---------------- | ------------------------------------------------------------------ |
| Operating System | Linux (Ubuntu, Debian, CentOS, RHEL, Fedora, Amazon Linux, Alpine) |
| Architecture     | x86_64 (amd64) or ARM64 (aarch64)                                  |
| Init System      | systemd                                                            |
| Network          | Outbound HTTPS access to api.monify.cloud                          |

## Installation

Install the agent with a single command:

```bash
curl -sSL https://monify.cloud/install.sh | sudo bash -s -- YOUR_TOKEN
```

Replace `YOUR_TOKEN` with the server token from your [Monify Dashboard](https://dash.monify.cloud).

The installation script will:

1. Download the appropriate binary for your architecture
2. Install to `/usr/local/bin/monify`
3. Create configuration at `/etc/monify/env`
4. Set up and start the systemd service

## Updating

To update to the latest version:

```bash
curl -sSL https://monify.cloud/update.sh | sudo bash
```

Your existing configuration and token will be preserved.

## Uninstallation

To completely remove the agent:

```bash
curl -sSL https://monify.cloud/uninstall.sh | sudo bash
```

This removes the binary, service, configuration, and log files.

## Managing the Service

The agent runs as a systemd service named `monify`:

```bash
# Check service status
sudo systemctl status monify

# View real-time logs
sudo journalctl -u monify -f

# Restart the service
sudo systemctl restart monify

# Stop the service
sudo systemctl stop monify
```

## Metrics Reference

### Static Metrics

Collected at startup and refreshed every hour:

| Metric           | Description                     |
| ---------------- | ------------------------------- |
| Platform         | Operating system distribution   |
| Platform Version | Distribution version            |
| Kernel Version   | Linux kernel version            |
| Architecture     | CPU architecture                |
| Virtualization   | Hypervisor or container runtime |
| CPU Model        | Processor model name            |
| CPU Cores        | Physical and logical core count |
| Total Memory     | Total system RAM                |
| IP Addresses     | Public and private IPs          |
| Cloud Provider   | AWS, GCP, Azure, or other       |
| Instance Type    | Cloud instance type             |
| Region           | Cloud region/zone               |

### Dynamic Metrics

Collected and transmitted every 15 seconds:

| Metric          | Description                            |
| --------------- | -------------------------------------- |
| CPU Usage       | Overall CPU utilization percentage     |
| Load Average    | System load (1m, 5m, 15m)              |
| Memory Usage    | Used, free, available, cached, buffers |
| Swap Usage      | Swap utilization                       |
| Disk Space      | Usage per mounted filesystem           |
| Disk I/O        | Read/write throughput and IOPS         |
| Network Traffic | Bandwidth per interface                |
| Network Errors  | Packet errors and drops                |
| Process Count   | Total running processes                |
| Uptime          | System uptime                          |

## Security

Monify Agent is built with security as a priority:

- **Encrypted Transport**: All communication uses TLS 1.2+
- **Token Authentication**: Each server has a unique authentication token
- **Minimal Privileges**: Requires root only for accessing system metrics
- **No Sensitive Data**: Does not collect file contents, logs, or user data
- **Systemd Hardening**: Runs with `NoNewPrivileges`, `ProtectSystem=strict`, and other security options

## Troubleshooting

**Agent not sending data**

```bash
# Check if service is running
sudo systemctl status monify

# View recent logs for errors
sudo journalctl -u monify -n 50 --no-pager
```

**Authentication errors**

```bash
# Re-install with a new token
curl -sSL https://monify.cloud/install.sh | sudo bash -s -- NEW_TOKEN
```

**High resource usage**

```bash
# Restart the service
sudo systemctl restart monify
```

## Support

- **Dashboard**: [dash.monify.cloud](https://dash.monify.cloud)
- **Website**: [monify.cloud](https://monify.cloud)
- **Email**: support@monify.cloud

## License

MIT License — see [LICENSE](LICENSE) for details.
