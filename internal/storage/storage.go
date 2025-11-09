package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/thomas/mavt/pkg/models"
)

// Storage handles persistence of app information and version updates
type Storage struct {
	dataDir string
	mu      sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &Storage{
		dataDir: dataDir,
	}, nil
}

// SaveApp saves app information to disk
func (s *Storage) SaveApp(app *models.AppInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	appFile := filepath.Join(s.dataDir, "apps", fmt.Sprintf("%s.json", app.BundleID))
	if err := os.MkdirAll(filepath.Dir(appFile), 0755); err != nil {
		return fmt.Errorf("failed to create apps directory: %w", err)
	}

	data, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal app data: %w", err)
	}

	if err := os.WriteFile(appFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write app file: %w", err)
	}

	return nil
}

// LoadApp loads app information from disk
func (s *Storage) LoadApp(bundleID string) (*models.AppInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	appFile := filepath.Join(s.dataDir, "apps", fmt.Sprintf("%s.json", bundleID))
	data, err := os.ReadFile(appFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read app file: %w", err)
	}

	var app models.AppInfo
	if err := json.Unmarshal(data, &app); err != nil {
		return nil, fmt.Errorf("failed to unmarshal app data: %w", err)
	}

	return &app, nil
}

// SaveVersionUpdate saves a version update event
func (s *Storage) SaveVersionUpdate(update *models.VersionUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	updatesFile := filepath.Join(s.dataDir, "updates", fmt.Sprintf("%s.json", update.BundleID))
	if err := os.MkdirAll(filepath.Dir(updatesFile), 0755); err != nil {
		return fmt.Errorf("failed to create updates directory: %w", err)
	}

	// Load existing updates
	var updates []models.VersionUpdate
	if data, err := os.ReadFile(updatesFile); err == nil {
		json.Unmarshal(data, &updates)
	}

	// Append new update
	updates = append(updates, *update)

	data, err := json.MarshalIndent(updates, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updates: %w", err)
	}

	if err := os.WriteFile(updatesFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write updates file: %w", err)
	}

	return nil
}

// GetAllApps returns all tracked apps
func (s *Storage) GetAllApps() ([]*models.AppInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	appsDir := filepath.Join(s.dataDir, "apps")
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.AppInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read apps directory: %w", err)
	}

	var apps []*models.AppInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(appsDir, entry.Name()))
		if err != nil {
			continue
		}

		var app models.AppInfo
		if err := json.Unmarshal(data, &app); err != nil {
			continue
		}

		apps = append(apps, &app)
	}

	return apps, nil
}

// GetVersionUpdates returns all version updates for a specific app
func (s *Storage) GetVersionUpdates(bundleID string) ([]models.VersionUpdate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	updatesFile := filepath.Join(s.dataDir, "updates", fmt.Sprintf("%s.json", bundleID))
	data, err := os.ReadFile(updatesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.VersionUpdate{}, nil
		}
		return nil, fmt.Errorf("failed to read updates file: %w", err)
	}

	var updates []models.VersionUpdate
	if err := json.Unmarshal(data, &updates); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updates: %w", err)
	}

	return updates, nil
}

// GetRecentUpdates returns all version updates within the specified duration
func (s *Storage) GetRecentUpdates(since time.Duration) ([]models.VersionUpdate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	updatesDir := filepath.Join(s.dataDir, "updates")
	entries, err := os.ReadDir(updatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.VersionUpdate{}, nil
		}
		return nil, fmt.Errorf("failed to read updates directory: %w", err)
	}

	cutoff := time.Now().Add(-since)
	var recentUpdates []models.VersionUpdate

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(updatesDir, entry.Name()))
		if err != nil {
			continue
		}

		var updates []models.VersionUpdate
		if err := json.Unmarshal(data, &updates); err != nil {
			continue
		}

		for _, update := range updates {
			if update.UpdatedAt.After(cutoff) {
				recentUpdates = append(recentUpdates, update)
			}
		}
	}

	return recentUpdates, nil
}
