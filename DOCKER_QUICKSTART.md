# Docker Quick Start Guide

This guide will help you get MAVT up and running quickly using Docker and the pre-built GitHub Container Registry image.

## Prerequisites

- Docker installed and running
- Docker Compose installed (comes with Docker Desktop)

## Quick Start

### 1. Download the docker-compose file

```bash
# Create a directory for MAVT
mkdir mavt && cd mavt

# Download the docker-compose file
curl -O https://raw.githubusercontent.com/thomas/mavt/main/docker-compose.ghcr.yml
```

### 2. Create the data volume

```bash
# Create a named volume for persistent storage
docker volume create mavt-data

# Set proper permissions (run as user 1000:1000)
docker run --rm -v mavt-data:/data alpine sh -c "chown -R 1000:1000 /data"
```

### 3. Configure your apps (optional)

Edit `docker-compose.ghcr.yml` to customize:
- `MAVT_APPS`: Add your app bundle IDs (comma-separated)
- `MAVT_CHECK_INTERVAL`: How often to check (default: 4h)
- `MAVT_COUNTRY`: Your App Store region (default: AU)
- `MAVT_APPRISE_URL`: Enable notifications (optional)

### 4. Start MAVT

```bash
# Start in daemon mode
docker-compose -f docker-compose.ghcr.yml up -d

# View logs
docker-compose -f docker-compose.ghcr.yml logs -f
```

### 5. Access the Web UI

Open your browser to: **http://localhost:7738**

You can:
- 🔍 Search for apps by name
- ➕ Add apps to tracking with one click
- 📊 View all tracked apps and their versions
- 📈 See recent version updates

## Finding Apps to Track

### Using the Web UI (Easiest)
1. Go to http://localhost:7738
2. Use the search bar to find apps by name (e.g., "Instagram", "WhatsApp")
3. Click "Track" to add them

### Common Bundle IDs

Here are some popular apps and their bundle IDs:

| App | Bundle ID |
|-----|-----------|
| Instagram | `com.burbn.instagram` |
| WhatsApp | `net.whatsapp.WhatsApp` |
| TikTok | `com.zhiliaoapp.musically` |
| Spotify | `com.spotify.client` |
| YouTube | `com.google.ios.youtube` |
| Twitter/X | `com.atebits.Tweetie2` |
| Facebook | `com.facebook.Facebook` |
| Reddit | `com.reddit.Reddit` |
| Discord | `com.hammerandchisel.discord` |
| Telegram | `ph.telegra.Telegraph` |
| Safari | `com.apple.mobilesafari` |
| Apple Music | `com.apple.Music` |
| Gmail | `com.google.Gmail` |
| Chrome | `com.google.chrome.ios` |

## Managing MAVT

### View logs
```bash
docker-compose -f docker-compose.ghcr.yml logs -f
```

### Stop MAVT
```bash
docker-compose -f docker-compose.ghcr.yml down
```

### Restart MAVT
```bash
docker-compose -f docker-compose.ghcr.yml restart
```

### Update to latest version
```bash
# Pull the latest image
docker-compose -f docker-compose.ghcr.yml pull

# Restart with new image
docker-compose -f docker-compose.ghcr.yml up -d
```

### Backup your data
```bash
# Export volume data
docker run --rm -v mavt-data:/data -v $(pwd):/backup alpine \
  tar czf /backup/mavt-backup.tar.gz -C /data .

# Restore from backup
docker run --rm -v mavt-data:/data -v $(pwd):/backup alpine \
  tar xzf /backup/mavt-backup.tar.gz -C /data
```

## Enabling Notifications

MAVT supports notifications via [Apprise](https://github.com/caronc/apprise), which can send to 80+ services.

### Option 1: Direct Service URLs (Simplest)

Edit `docker-compose.ghcr.yml` and set `MAVT_APPRISE_URL`:

```yaml
# Discord webhook
- MAVT_APPRISE_URL=discord://webhook_id/webhook_token

# Slack webhook
- MAVT_APPRISE_URL=slack://TokenA/TokenB/TokenC

# Telegram
- MAVT_APPRISE_URL=tgram://bot_token/chat_id

# Email (SMTP)
- MAVT_APPRISE_URL=mailto://user:pass@smtp.gmail.com
```

### Option 2: Using Apprise API

Uncomment the Apprise service in `docker-compose.ghcr.yml` and set:

```yaml
- MAVT_APPRISE_URL=http://apprise:8000/notify
```

See the [Apprise documentation](https://github.com/caronc/apprise/wiki) for all supported services.

## API Access

MAVT provides a REST API for programmatic access:

### Get all tracked apps
```bash
curl http://localhost:7738/api/apps
```

### Search for apps
```bash
curl "http://localhost:7738/api/search?q=instagram&limit=5"
```

### Add an app to tracking
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"bundle_id":"com.burbn.instagram"}' \
  http://localhost:7738/api/track
```

### Get recent updates
```bash
curl "http://localhost:7738/api/updates?since=24h"
```

### Get version history for a specific app
```bash
curl "http://localhost:7738/api/history?bundle_id=com.burbn.instagram"
```

### Health check
```bash
curl http://localhost:7738/api/health
```

## Troubleshooting

### Container shows as "unhealthy"

This is a common issue on Ubuntu/Linux systems due to volume permissions.

**Quick Fix:**
```bash
# Stop container
docker-compose -f docker-compose.ghcr.yml down

# Fix permissions
docker run --rm -v mavt-data:/data alpine sh -c "chown -R 1000:1000 /data && chmod -R 755 /data"

# Restart
docker-compose -f docker-compose.ghcr.yml up -d
```

### Container won't start
```bash
# Check logs
docker-compose -f docker-compose.ghcr.yml logs

# Check if volume exists
docker volume ls | grep mavt-data

# Check permissions
docker run --rm -v mavt-data:/data alpine ls -la /data
```

### Permission errors
```bash
# Fix volume permissions
docker run --rm -v mavt-data:/data alpine sh -c "chown -R 1000:1000 /data"
```

### Cannot access web UI
- Check if port 7738 is already in use
- Change the port mapping in docker-compose.ghcr.yml: `"8080:8080"` instead of `"7738:8080"`
- Check firewall settings (Ubuntu): `sudo ufw allow 7738`

### App not found errors
- Verify the bundle ID is correct
- Try a different `MAVT_COUNTRY` code (apps may not be available in all regions)
- Use the web UI search to find the correct bundle ID

### For More Help

See the comprehensive [TROUBLESHOOTING.md](TROUBLESHOOTING.md) guide for:
- Detailed debugging steps
- Ubuntu/Linux-specific issues
- SELinux/AppArmor problems
- Performance tuning
- And more...

## Advanced Configuration

### Custom Port
Change the port mapping in docker-compose.ghcr.yml:
```yaml
ports:
  - "8080:8080"  # Use port 8080 instead of 7738
```

### Multiple Notification Services
You can only set one `MAVT_APPRISE_URL`, but you can:
1. Use Apprise API with multiple configured services
2. Use a service like [ntfy.sh](https://ntfy.sh) that supports multiple subscriptions

### Different Check Intervals
Adjust based on your needs:
- More frequent: `MAVT_CHECK_INTERVAL=30m` (but respect API rate limits)
- Less frequent: `MAVT_CHECK_INTERVAL=12h` or `24h`

## Support

- Report issues: [GitHub Issues](https://github.com/thomas/mavt/issues)
- View source code: [GitHub Repository](https://github.com/thomas/mavt)
- Docker images: [GitHub Container Registry](https://github.com/thomas/mavt/pkgs/container/mavt)
