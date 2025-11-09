# MAVT - Mobile App Version Tracker

A Go application that tracks Apple App Store app version updates. Run it as a CLI tool or as a containerized daemon to monitor your favorite iOS apps for version changes.

## Features

- 📱 Track multiple iOS apps by bundle ID
- 🔄 Automatic version change detection
- 📝 Store complete version history with release notes
- ⏰ Daemon mode for continuous monitoring
- 🐳 Docker support with docker-compose
- 💾 Simple JSON file-based storage
- 🚀 No database required

## Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/thomas/mavt.git
cd mavt

# Create environment file
cp .env.example .env

# Edit .env to add apps you want to track
# MAVT_APPS=com.apple.mobilesafari,com.apple.Music

# Start the tracker
docker-compose up -d

# View logs
docker-compose logs -f
```

### Using Go

```bash
# Install dependencies
go mod download

# Build the application
go build -o mavt ./cmd/mavt

# Add apps to track
./mavt -add com.apple.mobilesafari
./mavt -add com.apple.Music

# Check for updates
./mavt -check

# Run as daemon
./mavt -daemon
```

## Usage

### CLI Commands

```bash
# Add an app to tracking
./mavt -add <bundle-id>

# List all tracked apps
./mavt -list

# Check for updates immediately
./mavt -check

# Run in daemon mode (continuous monitoring)
./mavt -daemon

# Show version history for a specific app
./mavt -updates <bundle-id>

# Show recent updates (e.g., last 24 hours)
./mavt -recent 24h
```

### Finding Bundle IDs

You can find an app's bundle ID by:
1. Searching on the App Store website
2. Looking at the URL: `https://apps.apple.com/us/app/app-name/id<TRACK_ID>`
3. Using the iTunes Search API to search by name

Example bundle IDs:
- Safari: `com.apple.mobilesafari`
- Apple Music: `com.apple.Music`
- Chrome: `com.google.chrome.ios`
- Facebook: `com.facebook.Facebook`

## Configuration

Configure via environment variables (see [.env.example](.env.example)):

| Variable | Description | Default |
|----------|-------------|---------|
| `MAVT_APPS` | Comma-separated list of bundle IDs to track | - |
| `MAVT_CHECK_INTERVAL` | How often to check for updates | `1h` |
| `MAVT_DATA_DIR` | Directory for storing data | `./data` |
| `MAVT_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `MAVT_SERVER_PORT` | HTTP server port (future use) | `8080` |
| `MAVT_SERVER_HOST` | HTTP server host (future use) | `0.0.0.0` |

## Data Storage

MAVT stores data as JSON files in the configured data directory:

```
data/
├── apps/
│   ├── com.apple.mobilesafari.json
│   └── com.apple.Music.json
└── updates/
    ├── com.apple.mobilesafari.json
    └── com.apple.Music.json
```

- `apps/` - Current version information for each tracked app
- `updates/` - Complete version history with timestamps and release notes

## Development

See [CLAUDE.md](CLAUDE.md) for detailed development documentation.

```bash
# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Format code
go fmt ./...

# Static analysis
go vet ./...
```

## Docker

### Building

```bash
docker build -t mavt:latest .
```

### Running

```bash
# Daemon mode with docker-compose
docker-compose up -d

# One-time check
docker run --rm -v $(pwd)/data:/app/data \
  -e MAVT_APPS=com.apple.mobilesafari \
  mavt:latest -check

# Interactive mode
docker run -it --rm -v $(pwd)/data:/app/data mavt:latest -list
```

## License

See [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
