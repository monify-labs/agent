# Release Notes - Monify Agent v0.2.0

**Release Date:** December 11, 2025  
**Type:** Initial Release

---

## 🎯 Overview

Version 0.2.0 is the initial release of Monify Agent - a lightweight, production-ready monitoring agent designed specifically for Linux servers. This release provides comprehensive system metrics collection, token-based authentication, and seamless integration with the Monify monitoring platform.

---

## 🚀 Key Features

### Core Functionality
- ✅ **System Metrics Collection** - CPU, Memory, Disk, Network, and System metrics
- ✅ **Real-time Monitoring** - 30-second collection interval (configurable)
- ✅ **Token-based Authentication** - Secure communication with Monify platform
- ✅ **Port Scanning** - Network port discovery and monitoring
- ✅ **Remote Commands** - Execute commands remotely (refresh, scan_ports, update_config)

### Performance & Efficiency
- ✅ **Lightweight** - < 1% CPU usage, < 50 MB memory
- ✅ **Efficient Transfer** - Gzip compression (~70-80% reduction)
- ✅ **Low Bandwidth** - ~10-20 MB/month data usage
- ✅ **Single Instance** - Lock mechanism prevents multiple instances

### Production Ready
- ✅ **Systemd Integration** - Native Linux service management
- ✅ **Auto-start on Boot** - Reliable service restart
- ✅ **Comprehensive Logging** - Configurable log levels with journald integration
- ✅ **Configuration Flexibility** - YAML files and environment variables

### Documentation & Support
- ✅ **Complete Documentation** - README, API docs, Contributing guide
- ✅ **Security Policy** - Vulnerability reporting and security best practices
- ✅ **Easy Installation** - One-line installation script
- ✅ **Multi-architecture** - Support for AMD64 and ARM64

---

## 📦 Installation

### Quick Install (Recommended)
```bash
curl -fsSL https://monify.cloud/install.sh | sudo bash -s -- YOUR_TOKEN
```

### Manual Installation

#### 1. Download Binary
```bash
# For AMD64 (x86_64)
wget https://github.com/monify-labs/agent/releases/download/v0.2.0/monify-linux-amd64
sudo mv monify-linux-amd64 /usr/local/bin/monify
sudo chmod +x /usr/local/bin/monify

# For ARM64
wget https://github.com/monify-labs/agent/releases/download/v0.2.0/monify-linux-arm64
sudo mv monify-linux-arm64 /usr/local/bin/monify
sudo chmod +x /usr/local/bin/monify
```

#### 2. Configure
```bash
sudo mkdir -p /etc/monify
echo "TOKEN_DEVICE=YOUR_TOKEN" | sudo tee /etc/monify/.env
```

#### 3. Create Systemd Service
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

#### 4. Start Service
```bash
sudo systemctl daemon-reload
sudo systemctl enable monify
sudo systemctl start monify
```

---

## 📊 System Requirements

### Supported Linux Distributions
- Ubuntu 18.04+ (LTS recommended)
- Debian 10+
- CentOS 7+
- RHEL 7+
- Fedora 30+
- Amazon Linux 2
- Other systemd-based distributions

### Architecture Support
- AMD64 (x86_64)
- ARM64 (aarch64)

### Minimum Requirements
- **CPU:** 1 core
- **RAM:** 100 MB available
- **Disk:** 50 MB free space
- **Network:** HTTPS connectivity to api.monify.cloud

---

## 📊 Performance Metrics

| Metric | Value |
|--------|-------|
| CPU Usage | < 1% |
| Memory Usage | < 50 MB |
| Bandwidth | ~10-20 MB/month |
| Compression | ~70-80% (gzip) |
| Binary Size | ~6-7 MB |

---

## 🔒 Security

- HTTPS-only communication
- Token-based authentication
- No sensitive data collection
- Open source and auditable code
- Regular security updates

**Report vulnerabilities:** See [SECURITY.md](SECURITY.md)

---

## 📝 Full Changelog

See [CHANGELOG.md](CHANGELOG.md) for complete details.

---

## 🐛 Known Issues

None at this time.

---

## 🆘 Support

- **Documentation:** [docs.monify.cloud](https://docs.monify.cloud)
- **Issues:** [github.com/monify-labs/agent/issues](https://github.com/monify-labs/agent/issues)
- **Email:** support@monify.cloud

---

## 🙏 Contributors

Thank you to everyone who contributed to this release!

---

## 📅 Next Steps

After installation/upgrade:
1. ✅ Verify agent is running: `monify status`
2. ✅ Check service status: `sudo systemctl status monify`
3. ✅ Monitor logs: `sudo journalctl -u monify -f`
4. ✅ Verify data in dashboard: [dash.monify.cloud](https://dash.monify.cloud)

---

**Made with ❤️ for Linux servers**
