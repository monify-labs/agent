# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

We take the security of Monify Agent seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Where to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to:

**security@monify.cloud**

### What to Include

Please include the following information in your report:

- **Type of vulnerability** (e.g., authentication bypass, code injection, etc.)
- **Full paths of source file(s)** related to the vulnerability
- **Location of the affected source code** (tag/branch/commit or direct URL)
- **Step-by-step instructions** to reproduce the issue
- **Proof-of-concept or exploit code** (if possible)
- **Impact of the vulnerability** and how it might be exploited
- **Your contact information** for follow-up questions

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity and complexity

We will:
1. Confirm receipt of your vulnerability report
2. Investigate and validate the issue
3. Develop and test a fix
4. Release a security patch
5. Publicly disclose the vulnerability (with credit to you, if desired)

## Security Best Practices

### For Users

1. **Keep Updated**: Always use the latest version of Monify Agent
2. **Secure Tokens**: Store authentication tokens securely
   - File permissions: `chmod 600 /etc/monify/.env`
   - Never commit tokens to version control
   - Rotate tokens regularly

3. **Network Security**:
   - Agent communicates only via HTTPS
   - Verify TLS certificates are validated
   - Use firewall rules to restrict outbound connections if needed

4. **Access Control**:
   - Run agent as root only when necessary
   - Review systemd service configuration
   - Monitor agent logs regularly: `journalctl -u monify -f`

5. **Configuration**:
   - Review and understand all configuration options
   - Disable unused features (e.g., port scanning if not needed)
   - Use minimal required permissions

### For Developers

1. **Code Review**: All code changes require review
2. **Dependencies**: Keep dependencies updated and audit regularly
3. **Input Validation**: Validate all external inputs
4. **Error Handling**: Never expose sensitive information in errors
5. **Logging**: Avoid logging sensitive data (tokens, credentials)
6. **Testing**: Include security-focused tests

## Security Features

### Current Security Measures

1. **Authentication**:
   - Token-based authentication
   - Tokens transmitted only via HTTPS
   - No password storage

2. **Data Protection**:
   - HTTPS-only communication
   - Gzip compression for efficiency
   - No collection of sensitive data (passwords, keys, file contents)

3. **System Security**:
   - Single-instance lock prevents multiple agents
   - Minimal system permissions required
   - No shell command execution from remote commands

4. **Code Security**:
   - Open source for transparency
   - Regular dependency updates
   - Static code analysis in CI/CD

### Data Collection

The agent collects **only** the following system metrics:

- CPU usage and load averages
- Memory and swap usage
- Disk space and I/O statistics
- Network interface statistics
- System uptime and process count
- OS and hardware information

**We do NOT collect**:
- File contents
- Environment variables (except those explicitly configured)
- User credentials or passwords
- Personal data
- Application data
- Command history

## Known Security Considerations

### Root Access Required

The agent requires root access to collect system metrics. This is necessary for:
- Reading system statistics from `/proc` and `/sys`
- Accessing network interface information
- Monitoring all processes

**Mitigation**:
- Agent runs as systemd service with minimal privileges
- No remote code execution capabilities
- Open source code for audit
- Single-instance lock prevents unauthorized instances

### Network Communication

The agent sends data to Monify API servers.

**Mitigation**:
- HTTPS-only communication
- Certificate validation enabled by default
- Configurable server endpoint
- Gzip compression reduces data size

## Vulnerability Disclosure Policy

When we receive a security bug report, we will:

1. **Confirm** the problem and determine affected versions
2. **Audit** code to find similar problems
3. **Prepare** fixes for all supported versions
4. **Release** new versions as soon as possible
5. **Announce** the vulnerability in:
   - GitHub Security Advisories
   - CHANGELOG.md
   - Release notes
   - Email to users (if critical)

## Security Updates

Security updates are released as:
- **Critical**: Immediate patch release
- **High**: Patch within 7 days
- **Medium**: Patch within 30 days
- **Low**: Included in next regular release

## Compliance

The Monify Agent is designed to be compliant with:
- General security best practices
- Open source security standards
- Minimal data collection principles

## Contact

For security concerns:
- **Email**: security@monify.cloud
- **PGP Key**: Available on request

For general questions:
- **Email**: support@monify.cloud
- **Issues**: https://github.com/monify-labs/agent/issues

---

**Last Updated**: December 11, 2024
