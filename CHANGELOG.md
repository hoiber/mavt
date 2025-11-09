# Changelog

All notable changes to MAVT (Mobile App Version Tracker) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2025-11-10

### Added
- **Version Management**:
  - `-version` flag to display version information
  - Version displayed in daemon startup logs
  - Version included in `/api/health` endpoint response
  - Version shown in web UI footer with GitHub link
  - `VERSION` file and `internal/version` package

- **Web Interface**: Beautiful web dashboard accessible at configured server port
  - Search & Add Apps section with live App Store search
  - Tracked Apps display with version info and last checked timestamps
  - Recent Updates view showing version changes from last 7 days
  - Auto-refresh every 30 seconds
  - Modern, responsive UI with gradient header and card-based layout
  - Footer with version information

- **REST API Endpoints**:
  - `GET /` - Web UI dashboard
  - `GET /api/search?q=<term>&limit=<number>` - Search App Store by name
  - `POST /api/track` - Add apps to tracking via API
  - `GET /api/apps` - List all tracked apps (JSON)
  - `GET /api/updates?since=<duration>` - Get recent version updates
  - `GET /api/health` - Health check endpoint

- **App Store Search**:
  - Search apps by name via iTunes Search API
  - Support for limit parameter (1-50 results)
  - Returns full app metadata including bundle ID, version, developer

- **HTTP Server Package** (`internal/server/`):
  - Web interface with embedded HTML/CSS/JavaScript
  - RESTful API handlers
  - Runs in daemon mode alongside version checking
  - Configurable host and port via environment variables

### Changed
- Updated `internal/appstore/client.go` to include `SearchApps()` method
- Modified daemon mode to start HTTP server in goroutine
- Enhanced documentation in README.md and CLAUDE.md with web UI and API usage

### Technical Details
- Web UI uses vanilla JavaScript with Fetch API
- Search debouncing (500ms) to reduce API calls
- Button state management for track operations
- Error handling for all API operations
- CORS-friendly JSON endpoints

## [1.0.0] - 2025-11-10

### Added
- Initial release
- CLI tool for tracking iOS app versions
- Daemon mode with configurable check intervals
- JSON file-based storage
- iTunes Search API integration
- Docker and docker-compose support
- Version history tracking
- Environment-based configuration
- Health monitoring and logging
