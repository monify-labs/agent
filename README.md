# Monify Agent v1.0.0

[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/monify-labs/agent/releases)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://golang.org/dl/)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos-lightgrey.svg)]()

**Lightweight system monitoring agent with continuous 1-second sampling for accurate metrics.**

---

## 🚀 Quick Start

```bash
# Download and install
curl -sSL https://monify.cloud/install.sh | sudo bash

# Configure your token
sudo monify login

# Or provide token directly
sudo monify login YOUR_TOKEN

# Start service
sudo systemctl start monify
sudo systemctl enable monify
```

Get your token at: [dash.monify.cloud](https://dash.monify.cloud) → Servers → Add Server

---

## ✨ Features

- **Continuous Sampling**: 1-second background sampling catches all spikes
- **Aggregated Metrics**: 70% smaller payloads (1-2 KB vs 5-8 KB)
- **Accurate Averages**: Averages from 15-60 samples per collection
- **Low Overhead**: < 1% CPU, < 50 MB memory
- **Flexible**: 15s-300s configurable collection intervals
- **Secure**: Token auth, TLS encrypted, no personal data sent

---

## 📊 Metrics Collected

### What We Measure

| Category       | Metrics                             | Sampling       |
| -------------- | ----------------------------------- | -------------- |
| **CPU**        | Usage %, Load avg (1m, 5m, 15m)     | 1s continuous  |
| **Memory**     | Used, cached, buffers, swap         | 1s continuous  |
| **Disk Space** | Total, used, free (all partitions)  | Per collection |
| **Disk I/O**   | Read/write MB/s, IOPS (all devices) | 1s continuous  |
| **Network**    | Public/private bandwidth, health    | 1s continuous  |
| **System**     | Uptime, process count               | Per collection |

### Payload Example

```json
{
  "hostname": "web-01",
  "timestamp": "2024-12-12T10:00:00Z",
  "cpu": {
    "usage_percent": 45.8,
    "load_avg_1m": 2.5,
    "load_avg_5m": 2.1,
    "load_avg_15m": 1.8
  },
  "memory": {
    "total": 8589934592,
    "used": 6442450944,
    "used_percent": 75.0
  },
  "swap": {
    "total": 4294967296,
    "used": 1073741824,
    "used_percent": 25.0
  },
  "disk_space": {
    "total": 1099511627776,
    "used": 549755813888,
    "used_percent": 50.0
  },
  "disk_io": {
    "read_mbps": 10.5,
    "write_mbps": 20.1,
    "read_iops": 1000,
    "write_iops": 2000
  },
  "network_public": {
    "send_mbps": 1.0,
    "recv_mbps": 2.0,
    "total_sent_gb": 1.15,
    "total_recv_gb": 9.2
  },
  "network_private": {
    "send_mbps": 0.4,
    "recv_mbps": 0.6,
    "total_sent_gb": 0.53,
    "total_recv_gb": 0.83
  },
  "network_health": {
    "errors_in": 0,
    "errors_out": 0,
    "drops_in": 5,
    "drops_out": 2
  }
}
```

**Size**: ~1.7 KB (70% smaller than previous version)

---

## ⚙️ Configuration

### Authentication

Token is stored in `/etc/monify/token`:

```bash
# Login with interactive prompt
sudo monify login

# Or provide token directly
sudo monify login YOUR_TOKEN
```

### Remote Configuration

All configuration is managed remotely from the server dashboard:

| Setting             | Default          | Configurable via Server |
| ------------------- | ---------------- | ----------------------- |
| Collection Interval | 30s              | ✅ Yes                  |
| Log Level           | info             | ✅ Yes                  |
| Server URL          | api.monify.cloud | ✅ Yes (migrations)     |
| Metrics Toggle      | All enabled      | ✅ Yes                  |

**Changes take effect immediately** without restarting the agent.

### Development Mode

For local development only:

```bash
# Use env var instead of token file
export MONIFY_TOKEN=your_dev_token
make dev
```

### Collection Intervals

| Interval | Use Case                 | Samples | Trade-off                       |
| -------- | ------------------------ | ------- | ------------------------------- |
| **15s**  | Real-time monitoring     | ~15     | Fast detection, higher traffic  |
| **30s**  | Production (recommended) | ~30     | Balanced                        |
| **60s**  | Cost-sensitive           | ~60     | Best accuracy, slower detection |

**Note**: All intervals use same 1s background sampling. Only affects how many samples averaged.

---

## 🔧 Management

### Systemd Commands

```bash
sudo systemctl start monify      # Start
sudo systemctl stop monify       # Stop
sudo systemctl restart monify    # Restart
sudo systemctl status monify     # Status
sudo systemctl enable monify     # Auto-start on boot

# Logs
sudo journalctl -u monify -f     # Follow logs
sudo journalctl -u monify -n 100 # Last 100 lines
```

### Agent Commands

```bash
monify status        # Check status
monify version       # Show version
monify login <token> # Configure token
```

### Uninstall

```bash
# Download and run uninstall script
curl -sSL https://monify.cloud/uninstall.sh | sudo bash

# Or if you have the script locally
sudo ./scripts/uninstall.sh
```

The uninstall script will:

- Stop and disable the systemd service
- Remove the binary from `/usr/local/bin`
- Remove the systemd service file
- Optionally remove config (`/etc/monify`) and logs (`/var/log/monify`)

---

## 📈 Performance

| Metric        | Value                                  |
| ------------- | -------------------------------------- |
| **CPU Usage** | < 1% (0.1% sampling + 0.5% collection) |
| **Memory**    | < 50 MB                                |
| **Network**   | ~5 KB/min (1.7 KB @ 30s interval)      |
| **Disk**      | Logs only (~1 MB/day)                  |

---

## 🛠️ Building from Source

```bash
git clone https://github.com/monify-labs/agent.git
cd agent

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run in development mode
make dev MONIFY_TOKEN=your_token_here

# Build for all Linux platforms
make build-linux
```

---

## 🔒 Security

- ✅ **HTTPS only** - All communication encrypted
- ✅ **Token auth** - Secure authentication
- ✅ **No personal data** - System metrics only
- ✅ **Open source** - Fully auditable
- ✅ **Minimal permissions** - Root needed for system metrics only

---

## ❓ FAQ

**Q: Does it work in Docker?**
A: Not recommended. Docker can't accurately measure host metrics. Use systemd service.

**Q: Can I change interval dynamically?**
A: Yes! Server can send `update_config` command.

**Q: How much data does it send?**
A: ~1.7 KB @ 30s = ~5 KB/min = ~7 MB/day per server.

**Q: What if server is down?**
A: Agent continues sampling but can't send. Data not buffered (prevents memory growth).

**Q: How accurate are metrics?**
A: Very accurate. 1s sampling catches all spikes, averages from 15-60 samples.

---

## 📝 Changelog

### v1.0.0 (2024-12-12)

**🎉 Initial Release**

- ✅ Continuous 1s sampling (CPU, memory, disk I/O, network)
- ✅ Aggregated metrics (70% payload reduction)
- ✅ Configurable intervals (15s-300s)
- ✅ Systemd service with auto-restart
- ✅ Token-based auth
- ✅ Server commands (dynamic config, pause, refresh)
- ✅ Linux and macOS support

---

## 📞 Support

- **Docs**: https://docs.monify.cloud
- **Issues**: https://github.com/monify-labs/agent/issues
- **Email**: support@monify.cloud

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file

---

<div align="center">

**Made with ❤️ by Monify Labs**

[Website](https://monify.cloud) • [Dashboard](https://dash.monify.cloud) • [Docs](https://docs.monify.cloud)

</div>
