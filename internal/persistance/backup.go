package persistence

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// BackupManager handles backup operations
type BackupManager struct {
	backupDir      string
	maxBackups     int
	compressionOn  bool
}

// NewBackupManager creates a new backup manager
func NewBackupManager(backupDir string, maxBackups int, compression bool) *BackupManager {
	return &BackupManager{
		backupDir:     backupDir,
		maxBackups:    maxBackups,
		compressionOn: compression,
	}
}

// CreateBackup creates a backup of a snapshot file
func (bm *BackupManager) CreateBackup(snapshotFile string) error {
	// Ensure backup directory exists
	if err := os.MkdirAll(bm.backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename
	timestamp := time.Now().Format("20060102_150405")
	baseFilename := filepath.Base(snapshotFile)
	backupFilename := fmt.Sprintf("backup_%s_%s", timestamp, baseFilename)
	
	if bm.compressionOn {
		backupFilename += ".gz"
	}
	
	backupPath := filepath.Join(bm.backupDir, backupFilename)

	// Copy file
	if bm.compressionOn {
		return bm.compressFile(snapshotFile, backupPath)
	}
	return bm.copyFile(snapshotFile, backupPath)
}

// copyFile copies a file from src to dst
func (bm *BackupManager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	logrus.Infof("Backup created: %s", dst)
	return nil
}

// compressFile compresses a file using gzip
func (bm *BackupManager) compressFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	if _, err := io.Copy(gzipWriter, sourceFile); err != nil {
		return fmt.Errorf("failed to compress file: %w", err)
	}

	logrus.Infof("Compressed backup created: %s", dst)
	return nil
}

// decompressFile decompresses a gzip file
func (bm *BackupManager) decompressFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	gzipReader, err := gzip.NewReader(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, gzipReader); err != nil {
		return fmt.Errorf("failed to decompress file: %w", err)
	}

	logrus.Infof("File decompressed: %s", dst)
	return nil
}

// CleanOldBackups removes old backups keeping only the specified maximum
func (bm *BackupManager) CleanOldBackups() error {
	files, err := os.ReadDir(bm.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	if len(files) <= bm.maxBackups {
		return nil
	}

	// Sort files by modification time
	type fileInfo struct {
		name    string
		modTime time.Time
	}

	fileInfos := make([]fileInfo, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fullPath := filepath.Join(bm.backupDir, file.Name())
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		fileInfos = append(fileInfos, fileInfo{
			name:    fullPath,
			modTime: info.ModTime(),
		})
	}

	// Sort by modification time (oldest first)
	for i := 0; i < len(fileInfos)-1; i++ {
		for j := i + 1; j < len(fileInfos); j++ {
			if fileInfos[i].modTime.After(fileInfos[j].modTime) {
				fileInfos[i], fileInfos[j] = fileInfos[j], fileInfos[i]
			}
		}
	}

	// Delete oldest files
	deleteCount := len(fileInfos) - bm.maxBackups
	for i := 0; i < deleteCount; i++ {
		if err := os.Remove(fileInfos[i].name); err != nil {
			logrus.Warnf("Failed to delete old backup %s: %v", fileInfos[i].name, err)
		} else {
			logrus.Infof("Deleted old backup: %s", fileInfos[i].name)
		}
	}

	return nil
}

// RestoreBackup restores a backup file
func (bm *BackupManager) RestoreBackup(backupFile, destFile string) error {
	if filepath.Ext(backupFile) == ".gz" {
		return bm.decompressFile(backupFile, destFile)
	}
	return bm.copyFile(backupFile, destFile)
}

// ListBackups returns a list of all backup files
func (bm *BackupManager) ListBackups() ([]string, error) {
	files, err := os.ReadDir(bm.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	backups := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() {
			backups = append(backups, filepath.Join(bm.backupDir, file.Name()))
		}
	}

	return backups, nil
}
