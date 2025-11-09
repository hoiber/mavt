package models

import "time"

// AppInfo represents an app's information from the App Store
type AppInfo struct {
	BundleID        string    `json:"bundle_id"`
	TrackID         int64     `json:"track_id"`
	TrackName       string    `json:"track_name"`
	Version         string    `json:"version"`
	ReleaseDate     time.Time `json:"release_date"`
	ReleaseNotes    string    `json:"release_notes"`
	ArtistName      string    `json:"artist_name"`
	MinOSVersion    string    `json:"min_os_version"`
	FileSizeBytes   int64     `json:"file_size_bytes"`
	Price           float64   `json:"price"`
	Currency        string    `json:"currency"`
	LastChecked     time.Time `json:"last_checked"`
	FirstDiscovered time.Time `json:"first_discovered"`
}

// VersionUpdate represents a version change event
type VersionUpdate struct {
	BundleID     string    `json:"bundle_id"`
	TrackID      int64     `json:"track_id"`
	TrackName    string    `json:"track_name"`
	OldVersion   string    `json:"old_version"`
	NewVersion   string    `json:"new_version"`
	UpdatedAt    time.Time `json:"updated_at"`
	ReleaseNotes string    `json:"release_notes"`
}
