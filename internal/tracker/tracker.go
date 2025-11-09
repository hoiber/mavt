package tracker

import (
	"fmt"
	"log"
	"time"

	"github.com/thomas/mavt/internal/appstore"
	"github.com/thomas/mavt/internal/storage"
	"github.com/thomas/mavt/pkg/models"
)

// Tracker monitors app versions and detects updates
type Tracker struct {
	client  *appstore.Client
	storage *storage.Storage
}

// NewTracker creates a new app version tracker
func NewTracker(storage *storage.Storage) *Tracker {
	return &Tracker{
		client:  appstore.NewClient(),
		storage: storage,
	}
}

// TrackApp adds an app to tracking by bundle ID
func (t *Tracker) TrackApp(bundleID string) error {
	app, err := t.client.LookupByBundleID(bundleID)
	if err != nil {
		return fmt.Errorf("failed to lookup app: %w", err)
	}

	// Check if we already have this app
	existing, err := t.storage.LoadApp(bundleID)
	if err != nil {
		return fmt.Errorf("failed to load existing app: %w", err)
	}

	if existing == nil {
		// First time tracking this app
		log.Printf("Now tracking %s (%s) - version %s", app.TrackName, app.BundleID, app.Version)
	} else {
		// Update first discovered time
		app.FirstDiscovered = existing.FirstDiscovered
	}

	if err := t.storage.SaveApp(app); err != nil {
		return fmt.Errorf("failed to save app: %w", err)
	}

	return nil
}

// CheckForUpdates checks all tracked apps for version updates
func (t *Tracker) CheckForUpdates() ([]models.VersionUpdate, error) {
	apps, err := t.storage.GetAllApps()
	if err != nil {
		return nil, fmt.Errorf("failed to get tracked apps: %w", err)
	}

	var updates []models.VersionUpdate

	for _, app := range apps {
		update, err := t.checkSingleApp(app)
		if err != nil {
			log.Printf("Error checking %s: %v", app.BundleID, err)
			continue
		}

		if update != nil {
			updates = append(updates, *update)
		}
	}

	return updates, nil
}

// checkSingleApp checks a single app for updates
func (t *Tracker) checkSingleApp(existingApp *models.AppInfo) (*models.VersionUpdate, error) {
	// Fetch current version from App Store
	currentApp, err := t.client.LookupByBundleID(existingApp.BundleID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current version: %w", err)
	}

	// Preserve first discovered time
	currentApp.FirstDiscovered = existingApp.FirstDiscovered

	// Check if version changed
	if currentApp.Version != existingApp.Version {
		update := &models.VersionUpdate{
			BundleID:     currentApp.BundleID,
			TrackID:      currentApp.TrackID,
			TrackName:    currentApp.TrackName,
			OldVersion:   existingApp.Version,
			NewVersion:   currentApp.Version,
			UpdatedAt:    time.Now(),
			ReleaseNotes: currentApp.ReleaseNotes,
		}

		log.Printf("Version update detected for %s: %s -> %s",
			currentApp.TrackName, existingApp.Version, currentApp.Version)

		// Save the update
		if err := t.storage.SaveVersionUpdate(update); err != nil {
			return nil, fmt.Errorf("failed to save version update: %w", err)
		}

		// Update stored app info
		if err := t.storage.SaveApp(currentApp); err != nil {
			return nil, fmt.Errorf("failed to update app info: %w", err)
		}

		return update, nil
	}

	// No version change, just update last checked time
	currentApp.LastChecked = time.Now()
	if err := t.storage.SaveApp(currentApp); err != nil {
		return nil, fmt.Errorf("failed to update app info: %w", err)
	}

	return nil, nil
}

// GetTrackedApps returns all apps being tracked
func (t *Tracker) GetTrackedApps() ([]*models.AppInfo, error) {
	return t.storage.GetAllApps()
}

// GetVersionHistory returns version update history for an app
func (t *Tracker) GetVersionHistory(bundleID string) ([]models.VersionUpdate, error) {
	return t.storage.GetVersionUpdates(bundleID)
}
