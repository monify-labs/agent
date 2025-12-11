# Release Notes - Monify Agent v0.2.2

**Release Date:** December 11, 2025  
**Type:** Patch Release

---

## 🎯 Overview

Version 0.2.2 is a patch release that addresses critical accuracy issues in hardware metric collection. It ensures that CPU core/thread counts are reported correctly across different architectures and that irrelevant disk partitions (with 0 size) are excluded from monitoring.

---

## 🛠 Fixes & Improvements

### ✅ Accurate CPU Topology
- Fixed logic for counting physical cores vs. logical threads. The agent now correctly reports the number of physical cores and logical processors (threads) using `gopsutil/v3/cpu` 'Counts' method, ensuring accurate hardware representation in the dashboard.

### ✅ Cleaner Disk Metrics
- Implemented filtering to exclude disk partitions with 0 total size (e.g., pseudo-filesystems like `/proc`, `/sys` mounts or unmounted volumes) from the metrics payload. This reduces noise and ensures only relevant storage devices are monitored.

### ✅ Improved Status Authentication
- Fixed an issue where `monify status` would incorrectly report "Not authenticated" even when a valid token was configured in the environment file. The status command now correctly loads the environment configuration.

### ✅ Safer Installation Script
- The installation script now checks for existing running instances of the agent before proceeding. If an active instance is detecting, it will abort the installation and guide the user to uninstall first, preventing conflicts.

---

## 📦 Upgrade Instructions

If you are already running v0.2.0, you can upgrade simply by downloading the new binary.

### Automatic Upgrade
```bash
# Re-run the install script
curl -fsSL https://monify.cloud/install.sh | sudo bash -s -- YOUR_TOKEN
```

### Manual Upgrade
1. Stop the service: `sudo systemctl stop monify`
2. Download new binary:
   ```bash
   wget https://github.com/monify-labs/agent/releases/download/v0.2.1/monify-linux-amd64
   sudo mv monify-linux-amd64 /usr/local/bin/monify
   sudo chmod +x /usr/local/bin/monify
   ```
3. Start the service: `sudo systemctl start monify`

---

## 📝 Changelog

See [CHANGELOG.md](CHANGELOG.md) for full details.
