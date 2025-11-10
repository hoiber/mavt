package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thomas/mavt/internal/appstore"
	"github.com/thomas/mavt/internal/tracker"
	"github.com/thomas/mavt/internal/version"
	"github.com/thomas/mavt/pkg/models"
)

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
}

// NewServer creates a new HTTP server
func NewServer(tracker *tracker.Tracker) *Server {
	s := &Server{
		tracker:       tracker,
		appstoreClient: appstore.NewClient(),
		mux:           http.NewServeMux(),
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

	html := `<!DOCTYPE html>
<html>
<head>
    <title>MAVT - App Version Tracker</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px 20px;
            text-align: center;
            border-radius: 10px;
            margin-bottom: 30px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        h1 { font-size: 2.5em; margin-bottom: 10px; }
        .subtitle { font-size: 1.2em; opacity: 0.9; }
        .section {
            background: white;
            padding: 30px;
            margin-bottom: 20px;
            border-radius: 10px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h2 {
            color: #667eea;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 2px solid #f0f0f0;
        }
        .app-card {
            background: #f9f9f9;
            padding: 20px;
            margin: 15px 0;
            border-radius: 8px;
            border-left: 4px solid #667eea;
        }
        .app-name {
            font-size: 1.3em;
            font-weight: bold;
            color: #333;
            margin-bottom: 10px;
        }
        .app-details {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 10px;
            margin-top: 10px;
        }
        .detail {
            display: flex;
            flex-direction: column;
        }
        .detail-label {
            font-size: 0.85em;
            color: #666;
            margin-bottom: 3px;
        }
        .detail-value {
            font-weight: 500;
            color: #333;
        }
        .version {
            background: #667eea;
            color: white;
            padding: 4px 12px;
            border-radius: 20px;
            display: inline-block;
            font-weight: bold;
        }
        .update-card {
            background: #e8f5e9;
            padding: 15px;
            margin: 10px 0;
            border-radius: 8px;
            border-left: 4px solid #4caf50;
        }
        .update-header {
            font-weight: bold;
            color: #2e7d32;
            margin-bottom: 5px;
        }
        .update-time {
            color: #666;
            font-size: 0.9em;
        }
        .loading {
            text-align: center;
            padding: 40px;
            color: #666;
        }
        .error {
            background: #ffebee;
            color: #c62828;
            padding: 15px;
            border-radius: 8px;
            border-left: 4px solid #c62828;
        }
        .empty-state {
            text-align: center;
            padding: 40px;
            color: #999;
        }
        .search-box {
            margin-bottom: 20px;
        }
        .search-input {
            width: 100%;
            padding: 12px;
            font-size: 16px;
            border: 2px solid #ddd;
            border-radius: 8px;
            box-sizing: border-box;
        }
        .search-input:focus {
            outline: none;
            border-color: #667eea;
        }
        .search-results {
            margin-top: 15px;
        }
        .search-result-card {
            background: #f9f9f9;
            padding: 15px;
            margin: 10px 0;
            border-radius: 8px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            border-left: 4px solid #667eea;
        }
        .search-result-info {
            flex: 1;
        }
        .search-result-name {
            font-weight: bold;
            font-size: 1.1em;
            margin-bottom: 5px;
        }
        .search-result-details {
            color: #666;
            font-size: 0.9em;
        }
        .btn {
            padding: 8px 20px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 14px;
            transition: background 0.3s;
        }
        .btn:hover {
            background: #764ba2;
        }
        .btn:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        .btn-success {
            background: #4caf50;
        }
        .success-message {
            background: #e8f5e9;
            color: #2e7d32;
            padding: 15px;
            border-radius: 8px;
            border-left: 4px solid #4caf50;
            margin: 10px 0;
        }
        .release-notes {
            background: #f5f5f5;
            padding: 12px;
            margin-top: 12px;
            border-radius: 6px;
            font-size: 0.9em;
            color: #555;
            line-height: 1.5;
            white-space: pre-wrap;
            border-left: 3px solid #667eea;
        }
        .release-notes-label {
            font-weight: bold;
            color: #667eea;
            margin-bottom: 8px;
            display: block;
        }
        .update-notes {
            background: #f0f7f0;
            padding: 10px;
            margin-top: 8px;
            border-radius: 4px;
            font-size: 0.85em;
            color: #555;
            line-height: 1.4;
            white-space: pre-wrap;
            max-height: 150px;
            overflow-y: auto;
        }
        .toggle-notes {
            color: #667eea;
            cursor: pointer;
            font-size: 0.9em;
            margin-top: 5px;
            display: inline-block;
            text-decoration: underline;
        }
        .toggle-notes:hover {
            color: #764ba2;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>📱 MAVT</h1>
            <p class="subtitle">Mobile App Version Tracker</p>
        </header>

        <div class="section">
            <h2>Search & Add Apps</h2>
            <div class="search-box">
                <input type="text" id="searchInput" class="search-input" placeholder="Search App Store (e.g., 'Instagram', 'WhatsApp')..." />
            </div>
            <div id="searchResults" class="search-results"></div>
        </div>

        <div class="section">
            <h2>Recent Updates (Last 7 Days)</h2>
            <div id="updates" class="loading">Loading updates...</div>
        </div>

        <div class="section">
            <h2>Tracked Apps</h2>
            <div id="apps" class="loading">Loading apps...</div>
        </div>
    </div>

    <footer style="text-align: center; padding: 20px; color: #999; font-size: 0.9em;">
        MAVT v` + version.Version + ` &bull; <a href="https://github.com/thomas/mavt" style="color: #667eea; text-decoration: none;">GitHub</a>
    </footer>

    <script>
        async function loadApps() {
            try {
                const response = await fetch('/api/apps');
                const apps = await response.json();
                const container = document.getElementById('apps');

                if (!apps || apps.length === 0) {
                    container.innerHTML = '<div class="empty-state">No apps are currently being tracked</div>';
                    return;
                }

                container.innerHTML = apps.map((app, index) => {
                    let releaseNotesHtml = '';
                    if (app.release_notes && app.release_notes.trim()) {
                        const notesId = 'app-notes-' + index;
                        releaseNotesHtml = '<span class="toggle-notes" onclick="toggleNotes(\'' + notesId + '\', this)">Latest Release Notes (v' + app.version + ') ▼</span>' +
                            '<div class="release-notes" id="' + notesId + '" style="display:none;">' +
                            app.release_notes +
                        '</div>';
                    }

                    return '<div class="app-card">' +
                        '<div class="app-name">' + app.track_name + '</div>' +
                        '<div class="app-details">' +
                            '<div class="detail">' +
                                '<span class="detail-label">Version</span>' +
                                '<span class="version">' + app.version + '</span>' +
                            '</div>' +
                            '<div class="detail">' +
                                '<span class="detail-label">Bundle ID</span>' +
                                '<span class="detail-value">' + app.bundle_id + '</span>' +
                            '</div>' +
                            '<div class="detail">' +
                                '<span class="detail-label">Developer</span>' +
                                '<span class="detail-value">' + app.artist_name + '</span>' +
                            '</div>' +
                            '<div class="detail">' +
                                '<span class="detail-label">Last Checked</span>' +
                                '<span class="detail-value">' + new Date(app.last_checked).toLocaleString() + '</span>' +
                            '</div>' +
                        '</div>' +
                        releaseNotesHtml +
                    '</div>';
                }).join('');
            } catch (error) {
                document.getElementById('apps').innerHTML =
                    '<div class="error">Failed to load apps: ' + error.message + '</div>';
            }
        }

        async function loadUpdates() {
            try {
                const response = await fetch('/api/updates?since=168h'); // 7 days
                const updates = await response.json();
                const container = document.getElementById('updates');

                if (!updates || updates.length === 0) {
                    container.innerHTML = '<div class="empty-state">No updates in the last 7 days</div>';
                    return;
                }

                container.innerHTML = updates.map((update, index) => {
                    let notesHtml = '';
                    if (update.release_notes && update.release_notes.trim()) {
                        const notesId = 'notes-' + index;
                        notesHtml = '<div class="update-notes" id="' + notesId + '" style="display:none;">' +
                            update.release_notes +
                        '</div>' +
                        '<span class="toggle-notes" onclick="toggleNotes(\'' + notesId + '\', this)">Show release notes ▼</span>';
                    }

                    return '<div class="update-card">' +
                        '<div class="update-header">' +
                            update.track_name + ': ' + update.old_version + ' → ' + update.new_version +
                        '</div>' +
                        '<div class="update-time">' + new Date(update.updated_at).toLocaleString() + '</div>' +
                        notesHtml +
                    '</div>';
                }).join('');
            } catch (error) {
                document.getElementById('updates').innerHTML =
                    '<div class="error">Failed to load updates: ' + error.message + '</div>';
            }
        }

        // Search functionality
        let searchTimeout;
        const searchInput = document.getElementById('searchInput');
        const searchResults = document.getElementById('searchResults');

        searchInput.addEventListener('input', (e) => {
            clearTimeout(searchTimeout);
            const query = e.target.value.trim();

            if (query.length < 2) {
                searchResults.innerHTML = '';
                return;
            }

            searchTimeout = setTimeout(() => searchApps(query), 500);
        });

        async function searchApps(query) {
            try {
                searchResults.innerHTML = '<div class="loading">Searching...</div>';
                const response = await fetch('/api/search?q=' + encodeURIComponent(query) + '&limit=10');
                const apps = await response.json();

                if (!apps || apps.length === 0) {
                    searchResults.innerHTML = '<div class="empty-state">No apps found</div>';
                    return;
                }

                searchResults.innerHTML = apps.map(app =>
                    '<div class="search-result-card">' +
                        '<div class="search-result-info">' +
                            '<div class="search-result-name">' + app.track_name + '</div>' +
                            '<div class="search-result-details">' +
                                app.artist_name + ' • v' + app.version + ' • ' + app.bundle_id +
                            '</div>' +
                        '</div>' +
                        '<button class="btn" onclick="trackApp(\'' + app.bundle_id + '\', this)">Track</button>' +
                    '</div>'
                ).join('');
            } catch (error) {
                searchResults.innerHTML = '<div class="error">Search failed: ' + error.message + '</div>';
            }
        }

        async function trackApp(bundleId, button) {
            const originalText = button.textContent;
            button.disabled = true;
            button.textContent = 'Adding...';

            try {
                const response = await fetch('/api/track', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ bundle_id: bundleId })
                });

                if (!response.ok) {
                    const error = await response.text();
                    throw new Error(error);
                }

                button.textContent = '✓ Added';
                button.className = 'btn btn-success';

                // Refresh tracked apps list
                setTimeout(() => {
                    loadApps();
                }, 1000);
            } catch (error) {
                alert('Failed to add app: ' + error.message);
                button.disabled = false;
                button.textContent = originalText;
            }
        }

        function toggleNotes(notesId, toggleElement) {
            const notesDiv = document.getElementById(notesId);
            const currentText = toggleElement.textContent;

            if (notesDiv.style.display === 'none') {
                notesDiv.style.display = 'block';
                // Replace ▼ with ▲ and update show/hide text
                if (currentText.includes('Latest Release Notes')) {
                    toggleElement.textContent = currentText.replace('▼', '▲');
                } else {
                    toggleElement.textContent = 'Hide release notes ▲';
                }
            } else {
                notesDiv.style.display = 'none';
                // Replace ▲ with ▼ and update show/hide text
                if (currentText.includes('Latest Release Notes')) {
                    toggleElement.textContent = currentText.replace('▲', '▼');
                } else {
                    toggleElement.textContent = 'Show release notes ▼';
                }
            }
        }

        // Load data on page load
        loadApps();
        loadUpdates();

        // Refresh every 30 seconds
        setInterval(() => {
            loadApps();
            loadUpdates();
        }, 30000);
    </script>
</body>
</html>`

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

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	json.NewEncoder(w).Encode(apps)
}

// handleTrack adds an app to tracking
func (s *Server) handleTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
