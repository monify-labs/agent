package updater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/monify-labs/agent/internal/config"
)

const (
	// Default download URL template
	// Agent will automatically replace {arch} with its own architecture (amd64/arm64)
	DefaultDownloadURL = "https://github.com/monify-labs/agent/releases/latest/download/monify-linux-{arch}"

	// Binary install path
	DefaultBinaryPath = "/usr/local/bin/monify"

	// Timeout for download
	DownloadTimeout = 5 * time.Minute
)

// UpdateParams contains parameters for update command
type UpdateParams struct {
	Version  string `json:"version"`
	URL      string `json:"url,omitempty"`      // Optional: custom download URL
	Checksum string `json:"checksum,omitempty"` // Optional: SHA256 checksum
	Force    bool   `json:"force,omitempty"`    // Force update even if same version
}

// Updater handles agent self-update
type Updater struct {
	binaryPath string
	debug      bool
}

// NewUpdater creates a new updater
func NewUpdater(debug bool) *Updater {
	return &Updater{
		binaryPath: DefaultBinaryPath,
		debug:      debug,
	}
}

// ShouldUpdate checks if update is needed
func (u *Updater) ShouldUpdate(targetVersion string, force bool) bool {
	if force {
		return true
	}

	// Compare versions (simple string comparison)
	currentVersion := config.Version
	return compareVersions(targetVersion, currentVersion) > 0
}

// Update performs the self-update
func (u *Updater) Update(ctx context.Context, params UpdateParams) error {
	log.Printf("INFO: Checking for update... current version: %s", config.Version)

	// Determine download URL (Hardcoded to GitHub Latest)
	downloadURL := strings.ReplaceAll(DefaultDownloadURL, "{arch}", runtime.GOARCH)

	// Create temp file for download
	tmpFile, err := os.CreateTemp("", "monify-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	tmpFile.Close()

	// Download new binary
	log.Printf("INFO: Downloading update from %s", downloadURL)
	if err := u.downloadFile(ctx, downloadURL, tmpPath); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Make executable to test it
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to chmod: %w", err)
	}

	// TRICK: Run the new binary to get its REAL version
	newVersion, err := u.getBinaryVersion(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to get new binary version: %w", err)
	}

	// RULE: Only update if version is different
	if !params.Force && newVersion == config.Version {
		log.Printf("INFO: Downloaded binary is same version (%s). Skipping update.", newVersion)
		return nil
	}

	log.Printf("INFO: New version detected: %s. Proceeding with update...", newVersion)

	// Verify checksum if provided
	if params.Checksum != "" {
		if err := u.verifyChecksum(tmpPath, params.Checksum); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	// Atomic replace
	if err := u.atomicReplace(tmpPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	log.Printf("INFO: Update to %s successful, restarting...", newVersion)

	// Restart agent
	return u.restart()
}

// getBinaryVersion runs the binary with --version and extracts the version string
func (u *Updater) getBinaryVersion(path string) (string, error) {
	cmd := exec.Command(path, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Output format: "Monify Agent v1.3" -> extract "1.3"
	outStr := string(output)
	parts := strings.Split(outStr, "v")
	if len(parts) < 2 {
		return "unknown", nil
	}

	version := strings.TrimSpace(parts[1])
	// Split by newline or spaces if any
	version = strings.Fields(version)[0]

	return version, nil
}

// downloadFile downloads a file from URL
func (u *Updater) downloadFile(ctx context.Context, url, dest string) error {
	ctx, cancel := context.WithTimeout(ctx, DownloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("monify-agent/%s", config.Version))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// verifyChecksum verifies SHA256 checksum of a file
func (u *Updater) verifyChecksum(path, expected string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actual := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}

	return nil
}

// extractTarGz extracts binary from tar.gz archive
func (u *Updater) extractTarGz(archivePath string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Look for the binary in the archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for monify binary
		if header.Typeflag == tar.TypeReg && strings.Contains(header.Name, "monify") {
			tmpFile, err := os.CreateTemp("", "monify-extracted-*")
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(tmpFile, tr); err != nil {
				tmpFile.Close()
				return "", err
			}
			tmpFile.Close()

			return tmpFile.Name(), nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

// verifyBinary runs the binary with --version to ensure it works
func (u *Updater) verifyBinary(path string) error {
	cmd := exec.Command(path, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("binary test failed: %w, output: %s", err, string(output))
	}

	if u.debug {
		log.Printf("DEBUG: New binary version output: %s", string(output))
	}

	return nil
}

// atomicReplace atomically replaces the current binary
func (u *Updater) atomicReplace(newBinaryPath string) error {
	// Get current binary path
	currentBinary, err := os.Executable()
	if err != nil {
		currentBinary = u.binaryPath
	}
	currentBinary, _ = filepath.EvalSymlinks(currentBinary)

	// Backup current binary
	backupPath := currentBinary + ".bak"
	if err := os.Rename(currentBinary, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Copy new binary to target location
	src, err := os.Open(newBinaryPath)
	if err != nil {
		// Restore backup on failure
		os.Rename(backupPath, currentBinary)
		return fmt.Errorf("failed to open new binary: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(currentBinary, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		// Restore backup on failure
		os.Rename(backupPath, currentBinary)
		return fmt.Errorf("failed to create new binary: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		// Restore backup on failure
		dst.Close()
		os.Remove(currentBinary)
		os.Rename(backupPath, currentBinary)
		return fmt.Errorf("failed to copy new binary: %w", err)
	}

	// Remove backup after successful update
	os.Remove(backupPath)

	return nil
}

// restart restarts the agent
func (u *Updater) restart() error {
	// Try systemctl restart first (if running as systemd service)
	if _, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command("systemctl", "restart", "monify")
		if err := cmd.Start(); err == nil {
			// Give systemd a moment to restart us
			time.Sleep(500 * time.Millisecond)
			os.Exit(0)
		}
	}

	// Fallback: exec ourselves
	binary, err := os.Executable()
	if err != nil {
		binary = u.binaryPath
	}

	args := os.Args
	env := os.Environ()

	log.Printf("INFO: Re-executing %s", binary)

	// syscall.Exec replaces current process with new one
	return syscall.Exec(binary, args, env)
}

// compareVersions compares two semantic versions
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Pad shorter version
	for len(parts1) < len(parts2) {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < len(parts1) {
		parts2 = append(parts2, "0")
	}

	for i := range parts1 {
		n1 := parseVersionPart(parts1[i])
		n2 := parseVersionPart(parts2[i])

		if n1 > n2 {
			return 1
		} else if n1 < n2 {
			return -1
		}
	}

	return 0
}

// parseVersionPart parses a version part as integer
func parseVersionPart(s string) int {
	// Remove any suffix like "-beta", "-rc1", etc.
	if idx := strings.IndexAny(s, "-+"); idx >= 0 {
		s = s[:idx]
	}

	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
