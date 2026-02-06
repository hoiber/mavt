package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/thomas/mavt/internal/appstore"
	"github.com/thomas/mavt/internal/backup"
	"github.com/thomas/mavt/internal/tracker"
	"github.com/thomas/mavt/internal/version"
	"github.com/thomas/mavt/pkg/models"
)

//go:embed static/index.html
var staticFS embed.FS

const (
	contentTypeHeader     = "Content-Type"
	contentTypeJSON       = "application/json"
	contentTypeHTML       = "text/html; charset=utf-8"
	methodNotAllowedMsg   = "Method not allowed"
	bundleIDField         = "bundle_id"
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

// Server handles HTTP requests
type Server struct {
	tracker       *tracker.Tracker
	appstoreClient *appstore.Client
	mux           *http.ServeMux
	checkInterval time.Duration
	dataDir       string
}

// NewServer creates a new HTTP server
func NewServer(tracker *tracker.Tracker, checkInterval time.Duration, dataDir string) *Server {
	s := &Server{
		tracker:       tracker,
		appstoreClient: appstore.NewClient(),
		mux:           http.NewServeMux(),
		checkInterval: checkInterval,
		dataDir:       dataDir,
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/apps", s.handleApps)
	s.mux.HandleFunc("/api/updates", s.handleUpdates)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/search", s.handleSearch)
	s.mux.HandleFunc("/api/track", s.handleTrack)
	s.mux.HandleFunc("/api/history", s.handleHistory)
	s.mux.HandleFunc("/api/last-update", s.handleLastUpdate)
	s.mux.HandleFunc("/api/export", s.handleExport)
	s.mux.HandleFunc("/api/import", s.handleImport)
}

// Start starts the HTTP server
func (s *Server) Start(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Starting HTTP server on http://%s", addr)
	return http.ListenAndServe(addr, s.mux)
}

// handleIndex serves the main page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Read the embedded HTML file
	htmlBytes, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "Failed to load page", http.StatusInternalServerError)
		return
	}

	html := string(htmlBytes)

	// Inject dynamic values
	checkIntervalMs := int64(s.checkInterval / time.Millisecond)
	html = strings.Replace(html, "{{VERSION}}", version.Version, -1)
	html = strings.Replace(html, "{{CHECK_INTERVAL_MS}}", fmt.Sprintf("%d", checkIntervalMs), -1)

	w.Header().Set(contentTypeHeader, contentTypeHTML)
	w.Write([]byte(html))
}

// handleApps returns all tracked apps
func (s *Server) handleApps(w http.ResponseWriter, r *http.Request) {
	apps, err := s.tracker.GetTrackedApps()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get apps: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(apps)
}

// handleUpdates returns recent version updates
func (s *Server) handleUpdates(w http.ResponseWriter, r *http.Request) {
	// Parse 'since' parameter (default to 24 hours)
	sinceStr := r.URL.Query().Get("since")
	if sinceStr == "" {
		sinceStr = "24h"
	}

	since, err := time.ParseDuration(sinceStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid 'since' parameter: %v", err), http.StatusBadRequest)
		return
	}

	// Get all apps to check their updates
	apps, err := s.tracker.GetTrackedApps()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get apps: %v", err), http.StatusInternalServerError)
		return
	}

	// Collect all updates within the timeframe
	cutoff := time.Now().Add(-since)
	var allUpdates []models.VersionUpdate

	for _, app := range apps {
		history, err := s.tracker.GetVersionHistory(app.BundleID)
		if err != nil {
			continue
		}

		for _, update := range history {
			if update.UpdatedAt.After(cutoff) {
				allUpdates = append(allUpdates, update)
			}
		}
	}

	// Sort updates by timestamp in descending order (most recent first)
	sort.Slice(allUpdates, func(i, j int) bool {
		return allUpdates[i].UpdatedAt.After(allUpdates[j].UpdatedAt)
	})

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(allUpdates)
}

// handleHealth returns health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	apps, err := s.tracker.GetTrackedApps()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "healthy",
		"version":      version.Version,
		"tracked_apps": len(apps),
		"timestamp":    time.Now(),
	})
}

// handleSearch searches for apps in the App Store
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, methodNotAllowedMsg, http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	apps, err := s.appstoreClient.SearchApps(query, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Get list of tracked apps to check which ones are already being tracked
	trackedApps, err := s.tracker.GetTrackedApps()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get tracked apps: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a map of tracked bundle IDs for quick lookup
	trackedMap := make(map[string]bool)
	for _, app := range trackedApps {
		trackedMap[app.BundleID] = true
	}

	// Add tracking status to search results
	type SearchResult struct {
		*models.AppInfo
		IsTracked bool `json:"is_tracked"`
	}

	results := make([]SearchResult, len(apps))
	for i, app := range apps {
		results[i] = SearchResult{
			AppInfo:   app,
			IsTracked: trackedMap[app.BundleID],
		}
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(results)
}

// handleTrack adds or removes an app from tracking
func (s *Server) handleTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, methodNotAllowedMsg, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BundleID string `json:"bundle_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.BundleID == "" {
		http.Error(w, "bundle_id is required", http.StatusBadRequest)
		return
	}

	// Handle DELETE request
	if r.Method == http.MethodDelete {
		if err := s.tracker.RemoveApp(req.BundleID); err != nil {
			http.Error(w, fmt.Sprintf("Failed to remove app: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("Removed app from tracking via API: %s", sanitizeForLog(req.BundleID))

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			bundleIDField: req.BundleID,
			"message":     "App successfully removed from tracking",
		})
		return
	}

	// Handle POST request (add app)
	if err := s.tracker.TrackApp(req.BundleID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to track app: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Added app to tracking via API: %s", sanitizeForLog(req.BundleID))

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		bundleIDField: req.BundleID,
		"message":     "App successfully added to tracking",
	})
}

// handleHistory returns version history for a specific app
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, methodNotAllowedMsg, http.StatusMethodNotAllowed)
		return
	}

	bundleID := r.URL.Query().Get("bundle_id")
	if bundleID == "" {
		http.Error(w, "Query parameter 'bundle_id' is required", http.StatusBadRequest)
		return
	}

	history, err := s.tracker.GetVersionHistory(bundleID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get version history: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(history)
}

// handleLastUpdate returns the timestamp of the most recent update
func (s *Server) handleLastUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, methodNotAllowedMsg, http.StatusMethodNotAllowed)
		return
	}

	// Get all apps to check their updates
	apps, err := s.tracker.GetTrackedApps()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get apps: %v", err), http.StatusInternalServerError)
		return
	}

	var latestUpdate time.Time

	// Find the most recent update across all apps
	for _, app := range apps {
		history, err := s.tracker.GetVersionHistory(app.BundleID)
		if err != nil {
			continue
		}

		for _, update := range history {
			if update.UpdatedAt.After(latestUpdate) {
				latestUpdate = update.UpdatedAt
			}
		}
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"last_update":   latestUpdate,
		"tracked_apps":  len(apps),
		"has_updates":   !latestUpdate.IsZero(),
	})
}

// handleExport creates and downloads a backup ZIP file
func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, methodNotAllowedMsg, http.StatusMethodNotAllowed)
		return
	}

	// Create temporary file for the backup
	tmpFile, err := os.CreateTemp("", "mavt-backup-*.zip")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create temporary file: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create backup
	if err := backup.Export(s.dataDir, tmpFile.Name()); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create backup: %v", err), http.StatusInternalServerError)
		return
	}

	// Read the backup file
	backupData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read backup file: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("mavt-backup-%s.zip", time.Now().Format("2006-01-02-150405"))

	// Send file as download
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(backupData)))
	w.Write(backupData)

	log.Printf("Export backup created: %s (%.2f MB)", filename, float64(len(backupData))/1024/1024)
}

// handleImport restores data from an uploaded backup ZIP file
func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, methodNotAllowedMsg, http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 100MB)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("backup")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get uploaded file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	if filepath.Ext(header.Filename) != ".zip" {
		http.Error(w, "Only ZIP files are supported", http.StatusBadRequest)
		return
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "mavt-import-*.zip")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create temporary file: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy uploaded file to temporary file
	if _, err := io.Copy(tmpFile, file); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save uploaded file: %v", err), http.StatusInternalServerError)
		return
	}
	tmpFile.Close()

	// Get file size for logging
	fileInfo, _ := os.Stat(tmpFile.Name())
	uploadSize := int64(0)
	if fileInfo != nil {
		uploadSize = fileInfo.Size()
	}

	log.Printf("Starting import from %s (%.2f MB)", sanitizeForLog(header.Filename), float64(uploadSize)/1024/1024)

	// Import backup with security validation
	metadata, err := backup.Import(tmpFile.Name(), s.dataDir)
	if err != nil {
		log.Printf("Import failed from %s: %v", sanitizeForLog(header.Filename), err)
		http.Error(w, fmt.Sprintf("Failed to import backup: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Import completed successfully: %s (backup version: %s, created: %s, apps: %d)",
		sanitizeForLog(header.Filename), metadata.Version, metadata.CreatedAt.Format(time.RFC3339), metadata.AppCount)

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"message":    "Backup successfully imported",
		"version":    metadata.Version,
		"created_at": metadata.CreatedAt,
		"app_count":  metadata.AppCount,
	})
}
