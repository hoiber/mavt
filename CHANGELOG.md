# Changelog

All notable changes to MAVT (Mobile App Version Tracker) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-11-10

### Added
- **Core Features**:
  - CLI tool for tracking iOS app versions
  - Daemon mode with configurable check intervals (default: 4 hours)
  - JSON file-based storage for app data and version history
  - iTunes Search API integration
  - Docker and docker-compose support
  - Environment-based configuration

- **Web Interface**: Beautiful web dashboard accessible at configured server port
  - Search & Add Apps section with live App Store search
  - Tracked Apps display with version info and last checked timestamps
  - Recent Updates view showing version changes from last 7 days
  - Version history modal - Click any tracked app to view complete version update history
  - Interactive version history table showing date, version changes, and release notes
  - Collapsible release notes in both Tracked Apps and Recent Updates sections
  - Dark mode toggle with localStorage persistence
  - Compact, single-line layout for app information
  - Auto-refresh every 30 seconds
  - Modern, responsive UI with gradient header and card-based layout
  - 30% larger text and wider container for better readability

- **Security Features**:
  - CVE detection in release notes with visual highlighting
  - Red pulsing version badges for apps with CVE mentions
  - Critical security update indicators in Recent Updates section
  - Automatic highlighting of security-critical updates

- **Notifications**:
  - Apprise integration for sending notifications when updates are detected
  - Support for 80+ notification services (Discord, Slack, Telegram, email, etc.)
  - Single and batch notification formatting
  - Configurable via `MAVT_APPRISE_URL` environment variable
  - Automatic notification sending on update detection

- **REST API Endpoints**:
  - `GET /` - Web UI dashboard
  - `GET /api/search?q=<term>&limit=<number>` - Search App Store by name
  - `POST /api/track` - Add apps to tracking via API
  - `GET /api/apps` - List all tracked apps (JSON)
  - `GET /api/updates?since=<duration>` - Get recent version updates
  - `GET /api/history?bundle_id=<id>` - Get version history for an app
  - `GET /api/health` - Health check endpoint

- **Version Management**:
  - `-version` flag to display version information
  - Version displayed in daemon startup logs
  - Version included in `/api/health` endpoint response
  - Version shown in web UI footer with GitHub link
  - `VERSION` file and `internal/version` package

### Technical Details
- Web UI uses vanilla JavaScript with Fetch API
- Search debouncing (500ms) to reduce API calls
- Button state management for track operations
- Error handling for all API operations
- CORS-friendly JSON endpoints
- Log injection attack prevention with input sanitization
- Flexbox-based responsive layout
