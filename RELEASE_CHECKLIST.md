# Release Checklist - v0.2.0

**Release Manager:** Dinoboy  
**Release Date:** 2025-12-11  
**Status:** Initial Release - Ready for Push


## Pre-Release Checklist

### Code Quality
- [x] All code formatted (`make fmt`)
- [x] Code passes `go vet` (`make vet`)
- [x] All tests pass (`make test`)
- [ ] Linter passes (`make lint`) - *golangci-lint not installed*
- [x] No compilation errors
- [x] Binaries built successfully for all platforms

### Documentation
- [x] CHANGELOG.md updated with all changes
- [x] README.md version updated
- [x] Breaking changes documented
- [x] CONTRIBUTING.md created
- [x] SECURITY.md created
- [x] API.md simplified and updated
- [x] Release notes created

### Version Control
- [x] All changes committed
- [x] Version tag created (v0.2.0)
- [x] Tag message includes release notes
- [ ] Changes pushed to remote
- [ ] Tag pushed to remote

### Build Artifacts
- [x] Linux AMD64 binary built
- [x] Linux ARM64 binary built
- [ ] Binaries tested on target platforms
- [ ] Release artifacts prepared

### Configuration
- [x] config.yaml.example updated
- [x] .env.example updated
- [x] Service file updated (monify.service)
- [x] Makefile updated

---

## Release Steps

### 1. Push to Remote
```bash
cd /Users/macos/monify/source/agent
git push origin main
git push origin v0.2.0
```

### 2. Create GitHub Release
- Go to: https://github.com/monify-labs/agent/releases/new
- Tag: v0.2.0
- Title: "Release v0.2.0 - Production Optimization"
- Description: Copy from RELEASE_NOTES_v0.2.0.md
- Attach binaries:
  - build/monify-linux-amd64
  - build/monify-linux-arm64

### 3. Update Installation Script (if needed)
- Verify install.sh points to correct version
- Test installation on clean system

### 4. Update Documentation Site (if exists)
- Update version references
- Add migration guide
- Update API documentation

### 5. Announce Release
- [ ] Update project website
- [ ] Send email to users (if applicable)
- [ ] Post on social media (if applicable)
- [ ] Update status page

---

## Post-Release Checklist

### Verification
- [ ] Installation script works with new version
- [ ] Upgrade path tested from v0.1.x
- [ ] Metrics collection working correctly
- [ ] Dashboard receiving data
- [ ] Service auto-starts on boot
- [ ] Logs are clean and informative

### Monitoring
- [ ] Monitor error reports
- [ ] Check GitHub issues
- [ ] Monitor support channels
- [ ] Track adoption metrics

### Communication
- [ ] Respond to user feedback
- [ ] Update FAQ if needed
- [ ] Document common issues

---

## Rollback Plan

If critical issues are discovered:

### 1. Quick Fix (Preferred)
```bash
# Fix the issue
git checkout -b hotfix/v0.2.1
# Make fixes
git commit -m "fix: critical issue description"
git tag -a v0.2.1 -m "Hotfix release"
git push origin hotfix/v0.2.1 v0.2.1
```

### 2. Rollback (If necessary)
- Mark v0.2.0 as deprecated in release notes
- Point users back to v0.1.2
- Document the issue and workaround

---

## Next Version Planning

### v0.2.1 (Patch - if needed)
- Bug fixes
- Documentation improvements
- Performance tweaks

### v0.3.0 (Minor - Future)
- New features
- Enhanced metrics
- Additional platform support (if needed)

---

## Notes

### Breaking Changes Summary
1. Windows support removed
2. Command structure simplified
3. Service file renamed
4. Docker configs removed

### Key Improvements
1. -1,159 net lines of code
2. Better documentation
3. Improved error handling
4. Enhanced performance

### Testing Platforms
- [ ] Ubuntu 20.04 LTS
- [ ] Ubuntu 22.04 LTS
- [ ] Debian 11
- [ ] Debian 12
- [ ] CentOS 8
- [ ] RHEL 8
- [ ] Amazon Linux 2

---

**Release Manager:** Dinoboy  
**Release Date:** 2024-12-11  
**Status:** Ready for Push
