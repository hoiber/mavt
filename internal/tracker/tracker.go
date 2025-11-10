package tracker

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/thomas/mavt/internal/appstore"
	"github.com/thomas/mavt/internal/notifier"
	"github.com/thomas/mavt/internal/storage"
	"github.com/thomas/mavt/pkg/models"
)

// sanitizeForLog removes newlines and control characters to prevent log injection attacks
func sanitizeForLog(s string) string {
	// Replace newlines and carriage returns with spaces
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	// Remove other control characters (ASCII 0-31 except space)
	var result strings.Builder
	for _, r := range s {
		if r >= 32 || r == '\t' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Tracker monitors app versions and detects updates
type Tracker struct {
	client   *appstore.Client
	storage  *storage.Storage
	notifier *notifier.Notifier
}

// NewTracker creates a new app version tracker
func NewTracker(storage *storage.Storage, notifier *notifier.Notifier) *Tracker {
	return &Tracker{
		client:   appstore.NewClient(),
		storage:  storage,
		notifier: notifier,
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
		log.Printf("Now tracking %s (%s) - version %s",
			sanitizeForLog(app.TrackName), sanitizeForLog(app.BundleID), sanitizeForLog(app.Version))
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
			log.Printf("Error checking %s: %v", sanitizeForLog(app.BundleID), err)
			continue
		}

		if update != nil {
			updates = append(updates, *update)
		}
	}

	// Send notifications if updates were found
	if len(updates) > 0 && t.notifier.IsEnabled() {
		if err := t.notifier.NotifyUpdates(updates); err != nil {
			log.Printf("Failed to send notifications: %v", err)
			// Don't fail the whole operation if notification fails
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
			sanitizeForLog(currentApp.TrackName),
			sanitizeForLog(existingApp.Version),
			sanitizeForLog(currentApp.Version))

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
