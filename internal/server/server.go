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
	s.mux.HandleFunc("/api/history", s.handleHistory)
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
        :root {
            --bg-primary: #f5f5f5;
            --bg-secondary: #ffffff;
            --bg-card: #ffffff;
            --text-primary: #333;
            --text-secondary: #666;
            --text-muted: #999;
            --border-color: #ddd;
            --accent-primary: #667eea;
            --accent-secondary: #764ba2;
            --success-bg: #e8f5e9;
            --success-text: #2e7d32;
            --success-border: #4caf50;
            --error-bg: #ffebee;
            --error-text: #c62828;
            --error-border: #c62828;
            --update-bg: #e8f5e9;
            --update-border: #4caf50;
            --release-notes-bg: #f5f5f5;
            --search-result-bg: #f9f9f9;
        }

        [data-theme="dark"] {
            --bg-primary: #1a1a1a;
            --bg-secondary: #2d2d2d;
            --bg-card: #2d2d2d;
            --text-primary: #e0e0e0;
            --text-secondary: #b0b0b0;
            --text-muted: #808080;
            --border-color: #404040;
            --accent-primary: #667eea;
            --accent-secondary: #764ba2;
            --success-bg: #1b3a1f;
            --success-text: #81c784;
            --success-border: #4caf50;
            --error-bg: #3d1f1f;
            --error-text: #ef5350;
            --error-border: #c62828;
            --update-bg: #1b3a1f;
            --update-border: #4caf50;
            --release-notes-bg: #242424;
            --search-result-bg: #353535;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.4;
            color: var(--text-primary);
            background: var(--bg-primary);
            transition: background-color 0.3s, color 0.3s;
            font-size: 18px;
        }
        .container {
            max-width: 1820px;
            margin: 0 auto;
            padding: 16px;
        }
        header {
            background: linear-gradient(135deg, var(--accent-primary) 0%, var(--accent-secondary) 100%);
            color: white;
            padding: 20px 16px;
            text-align: center;
            border-radius: 8px;
            margin-bottom: 16px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            position: relative;
        }
        .theme-toggle {
            position: absolute;
            top: 12px;
            right: 12px;
            background: rgba(255, 255, 255, 0.2);
            border: 2px solid rgba(255, 255, 255, 0.3);
            color: white;
            width: 32px;
            height: 32px;
            border-radius: 50%;
            cursor: pointer;
            font-size: 16px;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.3s;
        }
        .theme-toggle:hover {
            background: rgba(255, 255, 255, 0.3);
            transform: scale(1.1);
        }
        h1 { font-size: 1.8em; margin-bottom: 4px; }
        .subtitle { font-size: 0.95em; opacity: 0.9; }
        .section {
            background: var(--bg-card);
            padding: 16px;
            margin-bottom: 12px;
            border-radius: 8px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            transition: background-color 0.3s;
        }
        h2 {
            color: var(--accent-primary);
            margin-bottom: 12px;
            padding-bottom: 8px;
            border-bottom: 2px solid var(--border-color);
            font-size: 1.3em;
        }
        .app-card {
            background: var(--search-result-bg);
            padding: 10px 12px;
            margin: 6px 0;
            border-radius: 6px;
            border-left: 3px solid var(--accent-primary);
            transition: background-color 0.3s;
            display: flex;
            align-items: center;
            gap: 16px;
            flex-wrap: wrap;
        }
        .app-name {
            font-size: 1em;
            font-weight: bold;
            color: var(--text-primary);
            min-width: 200px;
            flex: 0 0 auto;
        }
        .app-details {
            display: flex;
            align-items: center;
            gap: 16px;
            flex: 1;
            flex-wrap: wrap;
        }
        .detail {
            display: flex;
            align-items: center;
            gap: 6px;
            white-space: nowrap;
        }
        .detail-label {
            font-size: 0.75em;
            color: var(--text-secondary);
        }
        .detail-value {
            font-weight: 500;
            color: var(--text-primary);
            font-size: 0.85em;
        }
        .version {
            background: var(--accent-primary);
            color: white;
            padding: 3px 10px;
            border-radius: 12px;
            display: inline-block;
            font-weight: bold;
            font-size: 0.9em;
        }
        .version.critical {
            background: #dc3545;
            animation: pulse 2s ease-in-out infinite;
        }
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.8; }
        }
        .update-card {
            background: var(--update-bg);
            padding: 10px;
            margin: 6px 0;
            border-radius: 6px;
            border-left: 3px solid var(--update-border);
            transition: background-color 0.3s;
        }
        .update-card.critical {
            border-left: 3px solid #dc3545;
            background: rgba(220, 53, 69, 0.1);
        }
        .update-header {
            font-weight: bold;
            color: var(--success-text);
            margin-bottom: 4px;
            font-size: 0.95em;
        }
        .update-header.critical {
            color: #dc3545;
        }
        .update-header.critical::before {
            content: '🔴 ';
        }
        .update-time {
            color: var(--text-secondary);
            font-size: 0.8em;
        }
        .loading {
            text-align: center;
            padding: 24px;
            color: var(--text-secondary);
            font-size: 0.9em;
        }
        .error {
            background: var(--error-bg);
            color: var(--error-text);
            padding: 12px;
            border-radius: 6px;
            border-left: 3px solid var(--error-border);
            transition: background-color 0.3s;
            font-size: 0.9em;
        }
        .empty-state {
            text-align: center;
            padding: 24px;
            color: var(--text-muted);
            font-size: 0.9em;
        }
        .search-box {
            margin-bottom: 12px;
        }
        .search-input {
            width: 100%;
            padding: 10px;
            font-size: 14px;
            border: 2px solid var(--border-color);
            border-radius: 6px;
            box-sizing: border-box;
            background: var(--bg-secondary);
            color: var(--text-primary);
            transition: background-color 0.3s, border-color 0.3s;
        }
        .search-input:focus {
            outline: none;
            border-color: var(--accent-primary);
        }
        .search-results {
            margin-top: 10px;
        }
        .search-result-card {
            background: var(--search-result-bg);
            padding: 10px;
            margin: 6px 0;
            border-radius: 6px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            border-left: 3px solid var(--accent-primary);
            transition: background-color 0.3s;
        }
        .search-result-info {
            flex: 1;
        }
        .search-result-name {
            font-weight: bold;
            font-size: 1em;
            margin-bottom: 3px;
            color: var(--text-primary);
        }
        .search-result-details {
            color: var(--text-secondary);
            font-size: 0.8em;
        }
        .btn {
            padding: 6px 16px;
            background: var(--accent-primary);
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            transition: background 0.3s;
        }
        .btn:hover {
            background: var(--accent-secondary);
        }
        .btn:disabled {
            background: var(--text-muted);
            cursor: not-allowed;
        }
        .btn-success {
            background: var(--success-border);
        }
        .success-message {
            background: var(--success-bg);
            color: var(--success-text);
            padding: 12px;
            border-radius: 6px;
            border-left: 3px solid var(--success-border);
            margin: 8px 0;
            transition: background-color 0.3s;
            font-size: 0.9em;
        }
        .release-notes {
            background: var(--release-notes-bg);
            padding: 10px;
            margin-top: 8px;
            border-radius: 6px;
            font-size: 0.85em;
            color: var(--text-secondary);
            line-height: 1.4;
            white-space: pre-wrap;
            border-left: 2px solid var(--accent-primary);
            transition: background-color 0.3s;
            flex-basis: 100%;
        }
        .release-notes-label {
            font-weight: bold;
            color: var(--accent-primary);
            margin-bottom: 6px;
            display: block;
            font-size: 0.9em;
        }
        .update-notes {
            background: var(--release-notes-bg);
            padding: 8px;
            margin-top: 6px;
            border-radius: 4px;
            font-size: 0.8em;
            color: var(--text-secondary);
            line-height: 1.3;
            white-space: pre-wrap;
            max-height: 120px;
            overflow-y: auto;
            transition: background-color 0.3s;
        }
        .toggle-notes {
            color: var(--accent-primary);
            cursor: pointer;
            font-size: 0.85em;
            margin-left: auto;
            display: inline-block;
            text-decoration: underline;
            white-space: nowrap;
            flex-shrink: 0;
        }
        .toggle-notes:hover {
            color: var(--accent-secondary);
        }
        footer a {
            color: var(--accent-primary);
            text-decoration: none;
        }
        footer a:hover {
            color: var(--accent-secondary);
        }
        .app-card {
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .app-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        }
        .modal {
            display: none;
            position: fixed;
            z-index: 1000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0, 0, 0, 0.5);
            animation: fadeIn 0.3s;
        }
        @keyframes fadeIn {
            from { opacity: 0; }
            to { opacity: 1; }
        }
        .modal-content {
            background-color: var(--bg-card);
            margin: 3% auto;
            padding: 0;
            border-radius: 8px;
            width: 92%;
            max-width: 1000px;
            max-height: 85vh;
            overflow: hidden;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
            animation: slideIn 0.3s;
        }
        @keyframes slideIn {
            from {
                transform: translateY(-30px);
                opacity: 0;
            }
            to {
                transform: translateY(0);
                opacity: 1;
            }
        }
        .modal-header {
            padding: 14px 20px;
            background: linear-gradient(135deg, var(--accent-primary) 0%, var(--accent-secondary) 100%);
            color: white;
            border-radius: 8px 8px 0 0;
        }
        .modal-header h2 {
            margin: 0;
            font-size: 1.2em;
        }
        .modal-header p {
            margin: 3px 0 0 0;
            opacity: 0.9;
            font-size: 0.8em;
        }
        .close {
            color: white;
            float: right;
            font-size: 24px;
            font-weight: bold;
            cursor: pointer;
            line-height: 1;
            opacity: 0.8;
            transition: opacity 0.2s;
        }
        .close:hover {
            opacity: 1;
        }
        .modal-body {
            padding: 16px;
            max-height: calc(85vh - 90px);
            overflow-y: auto;
        }
        .history-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 6px;
            font-size: 0.9em;
        }
        .history-table th {
            background: var(--bg-primary);
            color: var(--text-primary);
            padding: 8px 10px;
            text-align: left;
            font-weight: 600;
            border-bottom: 2px solid var(--border-color);
            position: sticky;
            top: 0;
            font-size: 0.85em;
        }
        .history-table td {
            padding: 8px 10px;
            border-bottom: 1px solid var(--border-color);
            color: var(--text-primary);
            font-size: 0.85em;
        }
        .history-table tr:hover {
            background: var(--search-result-bg);
        }
        .version-badge {
            display: inline-block;
            padding: 3px 8px;
            background: var(--accent-primary);
            color: white;
            border-radius: 4px;
            font-size: 0.85em;
            font-weight: 500;
        }
        .version-arrow {
            color: var(--text-muted);
            margin: 0 6px;
            font-size: 0.9em;
        }
        .history-notes {
            max-width: 400px;
            white-space: pre-wrap;
            font-size: 0.8em;
            color: var(--text-secondary);
            line-height: 1.3;
        }
        .empty-history {
            text-align: center;
            padding: 24px;
            color: var(--text-muted);
            font-style: italic;
            font-size: 0.9em;
        }
        .loading-history {
            text-align: center;
            padding: 24px;
            color: var(--text-secondary);
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <button class="theme-toggle" id="themeToggle" aria-label="Toggle dark mode">🌙</button>
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

    <!-- Version History Modal -->
    <div id="historyModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <span class="close" onclick="closeHistoryModal()">&times;</span>
                <h2 id="modalAppName">Version History</h2>
                <p id="modalAppDetails"></p>
            </div>
            <div class="modal-body" id="historyTableContainer">
                <div class="loading-history">Loading version history...</div>
            </div>
        </div>
    </div>

    <footer style="text-align: center; padding: 12px; color: var(--text-muted); font-size: 0.8em;">
        MAVT v` + version.Version + ` &bull; <a href="https://github.com/thomas/mavt">GitHub</a>
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
                    let releaseNotesToggle = '';
                    let releaseNotesContent = '';
                    let isCritical = false;
                    if (app.release_notes && app.release_notes.trim()) {
                        const notesId = 'app-notes-' + index;
                        // Check if release notes contain CVE
                        isCritical = app.release_notes.toUpperCase().includes('CVE');
                        releaseNotesToggle = '<span class="toggle-notes" onclick="event.stopPropagation(); toggleNotes(\'' + notesId + '\', this)">Latest Release Notes (v' + app.version + ') ▼</span>';
                        releaseNotesContent = '<div class="release-notes" id="' + notesId + '" style="display:none;">' +
                            app.release_notes +
                        '</div>';
                    }

                    const versionClass = isCritical ? 'version critical' : 'version';
                    return '<div class="app-card" onclick="showVersionHistory(\'' + app.bundle_id + '\', \'' + app.track_name.replace(/'/g, "\\'") + '\', \'' + app.artist_name.replace(/'/g, "\\'") + '\')">' +
                        '<div class="app-name">' + app.track_name + '</div>' +
                        '<span class="' + versionClass + '">' + app.version + '</span>' +
                        '<div class="app-details">' +
                            '<div class="detail">' +
                                '<span class="detail-label">Bundle:</span>' +
                                '<span class="detail-value">' + app.bundle_id + '</span>' +
                            '</div>' +
                            '<div class="detail">' +
                                '<span class="detail-label">Dev:</span>' +
                                '<span class="detail-value">' + app.artist_name + '</span>' +
                            '</div>' +
                            '<div class="detail">' +
                                '<span class="detail-label">Checked:</span>' +
                                '<span class="detail-value">' + new Date(app.last_checked).toLocaleString() + '</span>' +
                            '</div>' +
                        '</div>' +
                        releaseNotesToggle +
                        releaseNotesContent +
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
                    let isCritical = false;
                    if (update.release_notes && update.release_notes.trim()) {
                        const notesId = 'notes-' + index;
                        // Check if release notes contain CVE
                        isCritical = update.release_notes.toUpperCase().includes('CVE');
                        notesHtml = '<div class="update-notes" id="' + notesId + '" style="display:none;">' +
                            update.release_notes +
                        '</div>' +
                        '<span class="toggle-notes" onclick="toggleNotes(\'' + notesId + '\', this)">Show release notes ▼</span>';
                    }

                    const cardClass = isCritical ? 'update-card critical' : 'update-card';
                    const headerClass = isCritical ? 'update-header critical' : 'update-header';
                    return '<div class="' + cardClass + '">' +
                        '<div class="' + headerClass + '">' +
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

        // Version History Modal Functions
        async function showVersionHistory(bundleId, appName, developer) {
            const modal = document.getElementById('historyModal');
            const modalAppName = document.getElementById('modalAppName');
            const modalAppDetails = document.getElementById('modalAppDetails');
            const historyContainer = document.getElementById('historyTableContainer');

            // Set modal header info
            modalAppName.textContent = appName;
            modalAppDetails.textContent = developer + ' • ' + bundleId;

            // Show modal
            modal.style.display = 'block';

            // Load version history
            historyContainer.innerHTML = '<div class="loading-history">Loading version history...</div>';

            try {
                const response = await fetch('/api/history?bundle_id=' + encodeURIComponent(bundleId));
                if (!response.ok) {
                    throw new Error('Failed to load version history');
                }

                const history = await response.json();

                if (!history || history.length === 0) {
                    historyContainer.innerHTML = '<div class="empty-history">No version history available yet. Updates will appear here when the app version changes.</div>';
                    return;
                }

                // Build table
                let tableHtml = '<table class="history-table">' +
                    '<thead>' +
                        '<tr>' +
                            '<th>Date</th>' +
                            '<th>Version Change</th>' +
                            '<th>Release Notes</th>' +
                        '</tr>' +
                    '</thead>' +
                    '<tbody>';

                history.forEach(update => {
                    const dateStr = new Date(update.updated_at).toLocaleString();
                    const notesText = update.release_notes && update.release_notes.trim()
                        ? update.release_notes
                        : 'No release notes available';

                    tableHtml += '<tr>' +
                        '<td>' + dateStr + '</td>' +
                        '<td>' +
                            '<span class="version-badge">' + update.old_version + '</span>' +
                            '<span class="version-arrow">→</span>' +
                            '<span class="version-badge">' + update.new_version + '</span>' +
                        '</td>' +
                        '<td><div class="history-notes">' + notesText + '</div></td>' +
                    '</tr>';
                });

                tableHtml += '</tbody></table>';
                historyContainer.innerHTML = tableHtml;

            } catch (error) {
                historyContainer.innerHTML = '<div class="error">Failed to load version history: ' + error.message + '</div>';
            }
        }

        function closeHistoryModal() {
            document.getElementById('historyModal').style.display = 'none';
        }

        // Close modal when clicking outside of it
        window.onclick = function(event) {
            const modal = document.getElementById('historyModal');
            if (event.target === modal) {
                closeHistoryModal();
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

        // Dark mode functionality
        const themeToggle = document.getElementById('themeToggle');
        const htmlElement = document.documentElement;
        const THEME_KEY = 'mavt-theme';

        // Check for saved theme preference or default to light mode
        function getSavedTheme() {
            return localStorage.getItem(THEME_KEY) || 'light';
        }

        // Apply theme to the document
        function applyTheme(theme) {
            if (theme === 'dark') {
                htmlElement.setAttribute('data-theme', 'dark');
                themeToggle.textContent = '☀️';
            } else {
                htmlElement.removeAttribute('data-theme');
                themeToggle.textContent = '🌙';
            }
        }

        // Toggle between light and dark themes
        function toggleTheme() {
            const currentTheme = getSavedTheme();
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            localStorage.setItem(THEME_KEY, newTheme);
            applyTheme(newTheme);
        }

        // Apply saved theme on page load
        applyTheme(getSavedTheme());

        // Add click event listener to toggle button
        themeToggle.addEventListener('click', toggleTheme);
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
