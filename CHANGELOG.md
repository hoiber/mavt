# Changelog

All notable changes to MAVT (Mobile App Version Tracker) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.4] - 2025-11-10

### Added
- **GitHub Link**: Added GitHub repository link button in header for easy access to source code

### Changed
- **Modern UI Design**: Complete redesign with clean, modern aesthetic
  - Updated to modern blue color scheme (`#0066ff`) inspired by GitHub's design
  - Light mode: Clean light gray backgrounds with subtle borders
  - Dark mode: Deep dark backgrounds (`#0d1117`) with bright blue accents
- **Streamlined Header**: Redesigned header to be thinner and more compact
  - Reduced padding by 40% (from 20px to 12px)
  - Changed to horizontal layout with left/center/right sections
  - Removed subtitle for cleaner appearance
  - Title size reduced from 1.8em to 1.3em
- **Improved Visual Elements**:
  - Enhanced button hover effects with lift animation and subtle glow
  - Better card shadows and borders for depth
  - Refined typography and spacing throughout
  - Version badges now use modern blue styling

## [1.0.3] - 2025-11-10

### Added
- **Regional Support**: Added `MAVT_COUNTRY` environment variable to specify App Store region
  - Defaults to `AU` (Australia)
  - Supports any ISO 3166-1 alpha-2 country code (e.g., US, GB, CA, DE, FR, JP)
  - Fetches region-specific app metadata, release notes, and pricing
  - Configured via `MAVT_COUNTRY` in docker-compose.yml and environment variables

### Changed
- **App Store Client**: Updated to use country parameter in all API calls
  - `LookupByBundleID()` now uses configured country
  - `LookupByTrackID()` now uses configured country
  - `SearchApps()` now uses configured country
- **Tracker**: Updated to accept config parameter and pass country to App Store client
- **Documentation**: Added `MAVT_COUNTRY` to README.md configuration table

## [1.0.2] - 2025-11-10

### Fixed
- **GitHub Actions**: Added missing `id-token` and `attestations` permissions for build attestation
  - Resolves "Unable to get ACTIONS_ID_TOKEN_REQUEST_URL env variable" error

## [1.0.1] - 2025-11-10

### Changed
- **Security**: Pinned GitHub Actions to full commit hashes for improved security
  - `actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683` (v4.2.2)
  - `docker/setup-buildx-action@c47758b77c9736f4b2ef4073d4d51994fabfe349` (v3.7.1)
  - `docker/login-action@9780b0c442fbb1117ed29e0efdff1e18412f7567` (v3.3.0)
  - `docker/metadata-action@8e5442c4ef9f78752691e2d8f8d19755c6f78e81` (v5.5.1)
  - `docker/build-push-action@4f58ea79222b3b9dc2c8bbdd6debcef730109a75` (v6.9.0)
  - `actions/attest-build-provenance@1c608d11d69870c2092266b3f9a6f3abbf17002c` (v1.4.3)

### Fixed
- Added missing `id` to build-and-push step for attestation reference

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
