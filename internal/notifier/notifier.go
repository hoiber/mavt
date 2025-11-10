package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/thomas/mavt/pkg/models"
)

// Notifier handles sending notifications for app updates
type Notifier struct {
	appriseURL string
	enabled    bool
	client     *http.Client
}

// NewNotifier creates a new notifier instance
func NewNotifier(appriseURL string) *Notifier {
	enabled := appriseURL != ""

	return &Notifier{
		appriseURL: appriseURL,
		enabled:    enabled,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsEnabled returns whether notifications are enabled
func (n *Notifier) IsEnabled() bool {
	return n.enabled
}

// NotifyUpdate sends a notification for an app update
func (n *Notifier) NotifyUpdate(update *models.VersionUpdate) error {
	if !n.enabled {
		return nil
	}

	title := fmt.Sprintf("📱 %s Updated", update.TrackName)
	body := fmt.Sprintf("Version %s → %s", update.OldVersion, update.NewVersion)

	if update.ReleaseNotes != "" {
		// Truncate long release notes for notification
		notes := update.ReleaseNotes
		if len(notes) > 500 {
			notes = notes[:500] + "..."
		}
		body += "\n\n" + notes
	}

	return n.sendNotification(title, body, "info")
}

// NotifyUpdates sends a batch notification for multiple updates
func (n *Notifier) NotifyUpdates(updates []models.VersionUpdate) error {
	if !n.enabled || len(updates) == 0 {
		return nil
	}

	if len(updates) == 1 {
		return n.NotifyUpdate(&updates[0])
	}

	title := fmt.Sprintf("📱 %d App Updates Detected", len(updates))

	var body strings.Builder
	for i, update := range updates {
		if i > 0 {
			body.WriteString("\n")
		}
		body.WriteString(fmt.Sprintf("• %s: %s → %s",
			update.TrackName, update.OldVersion, update.NewVersion))

		// Limit to first 10 updates in notification
		if i >= 9 && len(updates) > 10 {
			body.WriteString(fmt.Sprintf("\n... and %d more", len(updates)-10))
			break
		}
	}

	return n.sendNotification(title, body.String(), "success")
}

// sendNotification sends a notification via Apprise API
func (n *Notifier) sendNotification(title, body, notifyType string) error {
	payload := map[string]interface{}{
		"title": title,
		"body":  body,
		"type":  notifyType,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification payload: %w", err)
	}

	req, err := http.NewRequest("POST", n.appriseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create notification request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("notification failed with status: %d", resp.StatusCode)
	}

	log.Printf("Notification sent: %s", title)
	return nil
}
