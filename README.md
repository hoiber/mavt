# MAVT - Mobile App Version Tracker

A Go application that tracks Apple App Store app version updates. Run it as a CLI tool or as a containerized daemon to monitor your favorite iOS apps for version changes.

## Features

- 📱 Track multiple iOS apps by bundle ID
- 🔍 **Web UI with App Store search** - Find and add apps instantly
- 🔄 Automatic version change detection
- 📝 Store complete version history with release notes
- ⏰ Daemon mode for continuous monitoring
- 🌐 Web dashboard with real-time updates
- 🔔 **Apprise notifications** - Get notified via Discord, Slack, email, Telegram, and 80+ services
- 🔌 REST API for programmatic access
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

# Access the web interface
# Open http://localhost:7738 in your browser
```

The web interface lets you:
- 🔍 Search the App Store by name
- ➕ Add apps to tracking with one click
- 📊 View all tracked apps and their versions
- 📈 See recent version updates

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
# Show version information
./mavt -version

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

### Web Interface

When running in daemon mode, access the web dashboard at `http://localhost:<port>` (default 8080, or 7738 in Docker).

**Features:**
- **Search Apps**: Type any app name to search the App Store (e.g., "Instagram", "WhatsApp")
- **One-Click Tracking**: Click "Track" button to instantly add apps to monitoring
- **Dashboard**: View all tracked apps with version info, last checked time, and developer
- **Update History**: See version changes from the last 7 days
- **Auto-Refresh**: Page updates every 30 seconds

### REST API

The HTTP server provides a REST API for programmatic access:

```bash
# Search for apps
curl "http://localhost:8080/api/search?q=instagram&limit=5"

# Get all tracked apps
curl http://localhost:8080/api/apps

# Add an app to tracking
curl -X POST -H "Content-Type: application/json" \
  -d '{"bundle_id":"com.burbn.instagram"}' \
  http://localhost:8080/api/track

# Get recent updates (last 24 hours)
curl "http://localhost:8080/api/updates?since=24h"

# Health check
curl http://localhost:8080/api/health
```

### Finding Bundle IDs

**Easiest way**: Use the web interface search! Just type the app name.

Or manually:
1. Search on the App Store website
2. Look at the URL: `https://apps.apple.com/us/app/app-name/id<TRACK_ID>`
3. Use the API: `curl "http://localhost:8080/api/search?q=app-name"`

Example bundle IDs:
- Instagram: `com.burbn.instagram`
- WhatsApp: `net.whatsapp.WhatsApp`
- TikTok: `com.zhiliaoapp.musically`
- Safari: `com.apple.mobilesafari`
- Apple Music: `com.apple.Music`

## Configuration

Configure via environment variables (see [.env.example](.env.example)):

| Variable | Description | Default |
|----------|-------------|---------|
| `MAVT_APPS` | Comma-separated list of bundle IDs to track | - |
| `MAVT_CHECK_INTERVAL` | How often to check for updates | `1h` |
| `MAVT_DATA_DIR` | Directory for storing data | `./data` |
| `MAVT_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `MAVT_SERVER_PORT` | HTTP server port | `8080` |
| `MAVT_SERVER_HOST` | HTTP server host | `0.0.0.0` |
| `MAVT_APPRISE_URL` | Apprise notification URL (optional) | - |

## Notifications

MAVT supports sending notifications via [Apprise](https://github.com/caronc/apprise) when app updates are detected. You can send notifications to Discord, Slack, email, Telegram, and 80+ other services.

### Setup with Apprise API

1. Run Apprise API service:
```bash
docker run -d -p 8000:8000 --name apprise caronc/apprise
```

2. Configure MAVT with the Apprise URL:
```bash
# In docker-compose.yml or .env
MAVT_APPRISE_URL=http://apprise:8000/notify
```

3. Configure your notification services in Apprise

### Direct Service URLs

You can also use direct service URLs without running Apprise API:

```bash
# Discord webhook
MAVT_APPRISE_URL=discord://webhook_id/webhook_token

# Slack webhook
MAVT_APPRISE_URL=slack://TokenA/TokenB/TokenC

# Telegram
MAVT_APPRISE_URL=tgram://bot_token/chat_id

# Email (SMTP)
MAVT_APPRISE_URL=mailto://user:pass@smtp.gmail.com
```

For more service URLs, see the [Apprise URL documentation](https://github.com/caronc/apprise/wiki).

### Notification Format

When updates are detected, MAVT sends notifications with:
- **Single update**: App name, version change, and release notes (truncated if long)
- **Multiple updates**: Summary of all updates (up to 10 shown, then "... and X more")

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
