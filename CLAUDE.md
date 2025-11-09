# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MAVT (Mobile App Version Tracker) is a Go application that monitors Apple App Store applications for version updates. It provides both CLI and daemon modes for tracking app versions over time.

**Key Features:**
- Track multiple iOS apps by bundle ID
- Detect and record version changes
- Store version history with release notes
- Run as a daemon for continuous monitoring
- Docker containerization support

## Development Commands

### Go Module Management
```bash
# Initialize Go module (if not already done)
go mod init github.com/username/mavt

# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Building
```bash
# Build the project
go build ./...

# Build specific package
go build ./path/to/package

# Build with output binary
go build -o mavt ./cmd/mavt
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Run specific test
go test -run TestName ./path/to/package

# Run tests matching a pattern
go test -run TestName.* ./...
```

### Code Quality
```bash
# Format code
go fmt ./...

# Run go vet for static analysis
go vet ./...

# Run golangci-lint (if installed)
golangci-lint run
```

### Running
```bash
# Add an app to track
go run ./cmd/mavt -add com.apple.mobilesafari

# List all tracked apps
go run ./cmd/mavt -list

# Check for updates immediately
go run ./cmd/mavt -check

# Run as daemon (continuous monitoring)
go run ./cmd/mavt -daemon

# Show version history for an app
go run ./cmd/mavt -updates com.apple.mobilesafari

# Show recent updates (e.g., last 24 hours)
go run ./cmd/mavt -recent 24h

# Build and run the binary
go build -o mavt ./cmd/mavt
./mavt -add com.apple.Music
```

### Docker Commands
```bash
# Build Docker image
docker build -t mavt:latest .

# Run with docker-compose (daemon mode)
docker-compose up -d

# View logs
docker-compose logs -f

# Stop container
docker-compose down

# Run one-time check
docker run --rm -v $(pwd)/data:/app/data \
  -e MAVT_APPS=com.apple.mobilesafari \
  mavt:latest -check

# Interactive shell in container
docker-compose exec mavt sh
```

## Project Structure

```
mavt/
├── cmd/mavt/              # Main application entry point
│   └── main.go            # CLI commands and daemon mode
├── internal/              # Private application code
│   ├── appstore/          # iTunes/App Store API client
│   │   └── client.go      # HTTP client for App Store lookups
│   ├── config/            # Configuration management
│   │   └── config.go      # Environment-based config loader
│   ├── storage/           # Data persistence layer
│   │   └── storage.go     # JSON file-based storage
│   └── tracker/           # Core tracking logic
│       └── tracker.go     # Version monitoring and updates
├── pkg/models/            # Public data models
│   └── app.go             # AppInfo and VersionUpdate structs
├── data/                  # Runtime data directory (gitignored)
│   ├── apps/              # Stored app information
│   └── updates/           # Version update history
├── Dockerfile             # Multi-stage Docker build
├── docker-compose.yml     # Docker Compose configuration
└── .env.example           # Example environment variables
```

## Architecture

### Data Flow
1. **App Store Client** ([internal/appstore/client.go](internal/appstore/client.go)) queries the iTunes Search API by bundle ID or track ID
2. **Tracker** ([internal/tracker/tracker.go](internal/tracker/tracker.go)) compares fetched data with stored versions to detect changes
3. **Storage** ([internal/storage/storage.go](internal/storage/storage.go)) persists app info and version updates as JSON files
4. **Main** ([cmd/mavt/main.go](cmd/mavt/main.go)) provides CLI interface and daemon mode orchestration

### Key Components

**App Store API Integration:**
- Uses iTunes Search API (`https://itunes.apple.com/lookup`)
- No authentication required
- Returns app metadata including version, release notes, and release date
- Rate limiting: Be respectful, no official limit documented

**Storage Strategy:**
- File-based JSON storage in `data/` directory
- `data/apps/{bundleID}.json` - Current app information
- `data/updates/{bundleID}.json` - Array of version updates
- Thread-safe with RWMutex for concurrent access

**Configuration:**
- Environment-based using `MAVT_*` prefixed variables
- See [.env.example](.env.example) for all options
- Supports comma-separated app lists for batch tracking

### Adding New Features

When extending functionality:
- App Store interactions go in `internal/appstore/`
- Business logic goes in `internal/tracker/`
- Data models in `pkg/models/` (public) or `internal/*/models.go` (private)
- CLI commands in `cmd/mavt/main.go`

## Environment Variables

See [.env.example](.env.example) for complete list. Key variables:
- `MAVT_APPS` - Comma-separated bundle IDs to track
- `MAVT_CHECK_INTERVAL` - How often to check (e.g., "1h", "30m")
- `MAVT_DATA_DIR` - Where to store tracking data
- `MAVT_LOG_LEVEL` - Logging verbosity (debug, info, warn, error)
