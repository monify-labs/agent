//go:build !windows
// +build !windows

package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

const (
	// LockFileName is the name of the lock file
	LockFileName = "monify.lock"
)

// Lock represents a file-based lock
type Lock struct {
	path string
	file *os.File
}

// NewLock creates a new lock instance
func NewLock(lockDir string) *Lock {
	return &Lock{
		path: filepath.Join(lockDir, LockFileName),
	}
}

// Acquire attempts to acquire the lock
// Returns error if another instance is already running
func (l *Lock) Acquire() error {
	// Try to open the lock file
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	// Try to acquire exclusive lock
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		file.Close()

		// Read PID from lock file if it exists
		existingPID := l.readPID()
		if existingPID > 0 {
			return fmt.Errorf("another instance is already running (PID: %d)\n\n"+
				"To stop the existing agent:\n"+
				"  kill %d\n\n"+
				"Or to forcefully remove the lock:\n"+
				"  rm %s\n\n"+
				"If you're sure no agent is running, remove the lock file manually.",
				existingPID, existingPID, l.path)
		}

		return fmt.Errorf("failed to acquire lock: another instance may be running\n\n"+
			"Lock file: %s\n\n"+
			"If you're sure no agent is running, remove the lock file manually:\n"+
			"  rm %s", l.path, l.path)
	}

	// Write current PID to lock file
	pid := os.Getpid()
	if err := file.Truncate(0); err != nil {
		file.Close()
		return fmt.Errorf("failed to truncate lock file: %w", err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		file.Close()
		return fmt.Errorf("failed to seek lock file: %w", err)
	}

	if _, err := file.WriteString(fmt.Sprintf("%d\n", pid)); err != nil {
		file.Close()
		return fmt.Errorf("failed to write PID to lock file: %w", err)
	}

	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("failed to sync lock file: %w", err)
	}

	l.file = file
	return nil
}

// Release releases the lock
func (l *Lock) Release() error {
	if l.file == nil {
		return nil
	}

	// Release the lock
	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	// Close the file
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close lock file: %w", err)
	}

	// Remove the lock file
	if err := os.Remove(l.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	l.file = nil
	return nil
}

// readPID reads the PID from the lock file
func (l *Lock) readPID() int {
	data, err := os.ReadFile(l.path)
	if err != nil {
		return 0
	}

	pid, err := strconv.Atoi(string(data[:len(data)-1])) // Remove newline
	if err != nil {
		return 0
	}

	return pid
}

// GetLockPath returns the path to the lock file
func (l *Lock) GetLockPath() string {
	return l.path
}
