# Monify Agent

> Lightweight monitoring agent for Linux servers

[![Platform](https://img.shields.io/badge/platform-linux-orange)](https://github.com/monify-labs/agent)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue)](https://golang.org)

---

## Quick Start

```bash
curl -fsSL https://monify.cloud/install.sh | sudo bash -s -- YOUR_TOKEN
```

**Get your token:** [dash.monify.cloud](https://dash.monify.cloud) → Servers → Add Server

---

## Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `monify start` | Start the agent (daemon mode) |
| `monify status` | Show agent status and configuration |
| `monify version` | Show version information |
| `monify help` | Show help message |

**Aliases:** `--version`, `-v`, `--help`, `-h`

### Service Management

| Command | Description |
|---------|-------------|
| `sudo systemctl start monify` | Start service |
| `sudo systemctl stop monify` | Stop service |
| `sudo systemctl restart monify` | Restart service |
| `sudo systemctl status monify` | Check service status |
| `sudo systemctl enable monify` | Enable auto-start on boot |

### Logs

| Command | Description |
|---------|-------------|
| `sudo journalctl -u monify -f` | Follow logs in real-time |
| `sudo journalctl -u monify -n 100` | Show last 100 log lines |
| `sudo journalctl -u monify --since "1 hour ago"` | Logs from last hour |

---

## Configuration

### Environment Variables

Edit `/etc/monify/.env`:

```bash
# Required
TOKEN_DEVICE=your_token_here

# Optional
MONIFY_COLLECTION_INTERVAL=30s    # Metrics collection interval
MONIFY_LOG_LEVEL=info             # Log level: debug, info, warn, error
MONIFY_SERVER_URL=https://api.monify.cloud/v1/agent/metrics
```

**Apply changes:**
```bash
sudo systemctl restart monify
```

### Advanced Configuration

Edit `/etc/monify/config.yaml`:

```yaml
collection:
  interval: 30s

metrics:
  cpu: true
  memory: true
  disk: true
  network: true
  system: true

logging:
  level: "info"
  format: "text"
```

---

## Metrics

The agent collects and sends the following metrics every 30 seconds:

| Category | Metrics |
|----------|---------|
| **CPU** | Usage percentage, per-core usage, load averages (1m, 5m, 15m) |
| **Memory** | Total, used, free, cached, swap usage |
| **Disk** | Space usage, I/O statistics for all mount points |
| **Network** | Traffic, packets, errors, drops for all interfaces |
| **System** | Uptime, boot time, process count |

**View metrics:** [dash.monify.cloud](https://dash.monify.cloud)

---

## Troubleshooting

### Check Status

```bash
monify status                      # Agent status
sudo systemctl status monify       # Service status
sudo journalctl -u monify -n 50    # Recent logs
```

### Common Issues

**Service not running:**
```bash
sudo systemctl restart monify
sudo journalctl -u monify -f
```

**No data in dashboard:**
```bash
cat /etc/monify/.env              # Verify token
curl https://api.monify.cloud/health  # Test connectivity
```

**Agent locked/stuck:**
```bash
sudo rm /var/run/monify.lock
sudo systemctl restart monify
```

---

## Uninstall

```bash
sudo systemctl stop monify
sudo systemctl disable monify
sudo rm /etc/systemd/system/monify.service
sudo rm /usr/local/bin/monify
sudo rm -rf /etc/monify
sudo systemctl daemon-reload
```

---

## Manual Installation

<details>
<summary>Expand for manual installation steps</summary>

### 1. Download Binary

```bash
wget https://github.com/monify-labs/agent/releases/latest/download/monify-linux-amd64
sudo mv monify-linux-amd64 /usr/local/bin/monify
sudo chmod +x /usr/local/bin/monify
```

### 2. Configure

```bash
sudo mkdir -p /etc/monify
echo "TOKEN_DEVICE=YOUR_TOKEN" | sudo tee /etc/monify/.env
```

### 3. Create Service

```bash
sudo tee /etc/systemd/system/monify.service > /dev/null <<'EOF'
[Unit]
Description=Monify Monitoring Agent
After=network.target

[Service]
Type=simple
User=root
EnvironmentFile=/etc/monify/.env
ExecStart=/usr/local/bin/monify start
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
```

### 4. Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable monify
sudo systemctl start monify
```

### 5. Verify

```bash
monify status
sudo systemctl status monify
```

</details>

---

## Performance

| Metric | Value |
|--------|-------|
| CPU Usage | < 1% |
| Memory Usage | < 50 MB |
| Bandwidth | ~10-20 MB/month |
| Compression | ~70-80% (gzip) |

---

## Security

- **Root Access:** Required for system metrics collection
- **Data Sent:** System metrics only (no file contents or personal data)
- **Transport:** HTTPS encrypted communication
- **Authentication:** Token-based authentication
- **Source Code:** Open source and auditable

---

## Support

- **Documentation:** [docs.monify.cloud](https://docs.monify.cloud)
- **Issues:** [github.com/monify-labs/agent/issues](https://github.com/monify-labs/agent/issues)
- **Email:** support@monify.cloud

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

<div align="center">

**Platform:** Linux • **Version:** 0.2.0 • **License:** MIT

Made with ❤️ for Linux servers

</div>
