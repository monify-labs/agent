# Release Notes - Monify Agent v0.2.3

**Release Date:** December 11, 2025  
**Type:** Patch Release

---

## 🎯 Overview

Version 0.2.3 significantly improves the accuracy of CPU and Memory metrics. Previously, metrics were collected as instantaneous snapshots every 30 seconds, which could miss usage spikes or report misleading low utilization. This release implements continuous background sampling (every 1 second) with 30-second averaging, ensuring that the reported metrics accurately reflect the true system load and align with other monitoring dashboards like DigitalOcean.

---

## 🛠 Fixes & Improvements

### ✅ Accurate CPU & Memory Measurement
- **Continuous Sampling**: The agent now samples CPU and Memory usage every 1 second in the background.
- **Averaging**: When reporting to the server (default every 30s), the agent calculates the average of all samples collected in that period.
- **Benefit**: This eliminates "blind spots" between collection intervals and provides a smooth, accurate representation of resource usage, matching what you see on provider dashboards.

### ✅ Graceful Shutdown
- Enhanced the collector architecture to properly clean up background sampling routines when the agent stops or restarts, preventing resource leaks.

---

## 📦 Upgrade Instructions

### Automatic Upgrade
```bash
# Re-run the install script
curl -fsSL https://monify.cloud/install.sh | sudo bash -s -- YOUR_TOKEN
```

### Manual Upgrade
1. Stop the service: `sudo systemctl stop monify`
2. Download new binary:
   ```bash
   wget https://github.com/monify-labs/agent/releases/download/v0.2.3/monify-linux-amd64
   sudo mv monify-linux-amd64 /usr/local/bin/monify
   sudo chmod +x /usr/local/bin/monify
   ```
3. Start the service: `sudo systemctl start monify`

---

## 📝 Changelog

See [CHANGELOG.md](CHANGELOG.md) for full details.
