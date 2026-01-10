package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// GameSnapshot represents a complete game state snapshot
type GameSnapshot struct {
	Timestamp      time.Time              `json:"timestamp"`
	Version        string                 `json:"version"`
	GameStatus     string                 `json:"game_status"`
	CurrentPot     int                    `json:"current_pot"`
	HighestBet     int                    `json:"highest_bet"`
	DealerID       int                    `json:"dealer_id"`
	CurrentTurn    int                    `json:"current_turn"`
	Players        []PlayerSnapshot       `json:"players"`
	CommunityCards []byte                 `json:"community_cards,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// PlayerSnapshot represents a player's state in a snapshot
type PlayerSnapshot struct {
	PlayerID         string `json:"player_id"`
	RotationID       int    `json:"rotation_id"`
	Stack            int    `json:"stack"`
	CurrentBet       int    `json:"current_bet"`
	TotalBetThisHand int    `json:"total_bet_this_hand"`
	IsActive         bool   `json:"is_active"`
	IsFolded         bool   `json:"is_folded"`
	IsAllIn          bool   `json:"is_all_in"`
	IsReady          bool   `json:"is_ready"`
}

// SaveSnapshot saves a game snapshot to a file
func SaveSnapshot(snapshot *GameSnapshot, filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	logrus.Infof("Game snapshot saved to %s", filename)
	return nil
}

// LoadSnapshot loads a game snapshot from a file
func LoadSnapshot(filename string) (*GameSnapshot, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}

	var snapshot GameSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}

	logrus.Infof("Game snapshot loaded from %s", filename)
	return &snapshot, nil
}

// SaveSnapshotWithTimestamp saves a snapshot with a timestamp in the filename
func SaveSnapshotWithTimestamp(snapshot *GameSnapshot, baseDir string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(baseDir, fmt.Sprintf("snapshot_%s.json", timestamp))
	
	err := SaveSnapshot(snapshot, filename)
	return filename, err
}

// ListSnapshots returns a list of all snapshot files in a directory
func ListSnapshots(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	snapshots := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			snapshots = append(snapshots, filepath.Join(dir, file.Name()))
		}
	}

	return snapshots, nil
}

// DeleteOldSnapshots deletes snapshots older than the specified duration
func DeleteOldSnapshots(dir string, maxAge time.Duration) error {
	files, err := ListSnapshots(dir)
	if err != nil {
		return err
	}

	now := time.Now()
	deletedCount := 0

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			logrus.Warnf("Failed to stat file %s: %v", file, err)
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			if err := os.Remove(file); err != nil {
				logrus.Warnf("Failed to delete old snapshot %s: %v", file, err)
			} else {
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		logrus.Infof("Deleted %d old snapshots from %s", deletedCount, dir)
	}

	return nil
}

// GetLatestSnapshot returns the most recent snapshot from a directory
func GetLatestSnapshot(dir string) (*GameSnapshot, error) {
	files, err := ListSnapshots(dir)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no snapshots found in %s", dir)
	}

	var latestFile string
	var latestTime time.Time

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestFile = file
		}
	}

	if latestFile == "" {
		return nil, fmt.Errorf("no valid snapshots found")
	}

	return LoadSnapshot(latestFile)
}
