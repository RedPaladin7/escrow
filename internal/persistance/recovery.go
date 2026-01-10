package persistence

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// RecoveryManager handles crash recovery
type RecoveryManager struct {
	snapshotDir string
	crashFile   string
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(snapshotDir, crashFile string) *RecoveryManager {
	return &RecoveryManager{
		snapshotDir: snapshotDir,
		crashFile:   crashFile,
	}
}

// MarkCrash creates a marker file indicating a crash
func (rm *RecoveryManager) MarkCrash() error {
	return os.WriteFile(rm.crashFile, []byte(time.Now().Format(time.RFC3339)), 0644)
}

// ClearCrashMarker removes the crash marker file
func (rm *RecoveryManager) ClearCrashMarker() error {
	if err := os.Remove(rm.crashFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove crash marker: %w", err)
	}
	return nil
}

// HasCrashed checks if there was a crash
func (rm *RecoveryManager) HasCrashed() bool {
	_, err := os.Stat(rm.crashFile)
	return err == nil
}

// RecoverFromCrash attempts to recover from a crash
func (rm *RecoveryManager) RecoverFromCrash() (*GameSnapshot, error) {
	if !rm.HasCrashed() {
		return nil, fmt.Errorf("no crash detected")
	}

	logrus.Warn("Crash detected, attempting recovery...")

	// Load the latest snapshot
	snapshot, err := GetLatestSnapshot(rm.snapshotDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load snapshot for recovery: %w", err)
	}

	// Clear crash marker
	if err := rm.ClearCrashMarker(); err != nil {
		logrus.Warnf("Failed to clear crash marker: %v", err)
	}

	logrus.Info("Recovery successful")
	return snapshot, nil
}

// PerformGracefulShutdown performs a graceful shutdown
func (rm *RecoveryManager) PerformGracefulShutdown(snapshot *GameSnapshot) error {
	logrus.Info("Performing graceful shutdown...")

	// Save final snapshot
	filename, err := SaveSnapshotWithTimestamp(snapshot, rm.snapshotDir)
	if err != nil {
		return fmt.Errorf("failed to save shutdown snapshot: %w", err)
	}

	logrus.Infof("Shutdown snapshot saved to %s", filename)

	// Clear crash marker
	return rm.ClearCrashMarker()
}
