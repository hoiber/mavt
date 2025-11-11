# Troubleshooting Guide

This guide covers common issues when running MAVT, especially on Linux/Ubuntu systems.

## Container Health Issues

### Container shows as "unhealthy"

**Symptoms:**
- Container starts but shows as "unhealthy" in `docker ps`
- Running `docker ps` shows status like "Up X minutes (unhealthy)"

**Common Causes & Solutions:**

#### 1. Volume Permission Issues (Most Common on Ubuntu/Linux)

The container runs as user `1000:1000`, and the volume needs proper permissions.

**Solution:**

```bash
# Stop the container
docker-compose down

# Create volume if it doesn't exist
docker volume create mavt-data

# Fix permissions on the volume
docker run --rm -v mavt-data:/data alpine sh -c "chown -R 1000:1000 /data && chmod -R 755 /data"

# Verify permissions
docker run --rm -v mavt-data:/data alpine ls -la /data

# Restart container
docker-compose up -d
```

#### 2. No Apps Tracked Yet

If you're using an older healthcheck that looks for `.json` files, it will fail until you add at least one app.

**Solution:**
- The latest docker-compose files have been updated to check if the data directory is writable instead
- Update your docker-compose file to use the latest healthcheck configuration
- Or add apps using the web UI at `http://localhost:7738`

#### 3. Data Directory Not Mounted

**Check if volume is mounted:**

```bash
docker inspect mavt | grep -A 10 Mounts
```

**Solution:**

```bash
# Recreate the container with proper volume
docker-compose down -v
docker volume create mavt-data
docker-compose up -d
```

### Check Container Logs

Always check logs for detailed error messages:

```bash
# View logs
docker-compose logs -f mavt

# Or with docker
docker logs mavt -f
```

## Common Startup Errors

### "Permission denied" errors

**Symptoms:**
```
Error: failed to create data directory: mkdir /app/data: permission denied
```

**Solution:**
Ensure volume has correct permissions (see Volume Permission Issues above).

### "Address already in use"

**Symptoms:**
```
Error: listen tcp :8080: bind: address already in use
```

**Solution:**
Change the host port in docker-compose.yml:

```yaml
ports:
  - "8080:8080"  # Change first number to different port
```

### Cannot connect to App Store API

**Symptoms:**
- Logs show connection timeouts to `itunes.apple.com`
- No apps can be added or updated

**Solution:**

1. Check internet connectivity:
   ```bash
   docker exec mavt ping -c 3 itunes.apple.com
   ```

2. Check DNS resolution:
   ```bash
   docker exec mavt nslookup itunes.apple.com
   ```

3. If behind proxy, you may need to configure Docker daemon with proxy settings.

## Ubuntu-Specific Issues

### SELinux/AppArmor Blocking Volume Access

**Symptoms:**
- Permission denied errors despite correct ownership
- Container unhealthy with no clear error

**Solution for SELinux:**

```bash
# Temporarily disable SELinux (not recommended for production)
sudo setenforce 0

# Or add volume label
docker run --rm -v mavt-data:/data:z alpine chown -R 1000:1000 /data
```

**Solution for AppArmor:**

```bash
# Check AppArmor status
sudo aa-status

# If Docker profile is blocking, you may need to reload
sudo systemctl reload apparmor
```

### Snap Docker Issues

If you installed Docker via Snap on Ubuntu, you may encounter permission issues.

**Solution:**
Install Docker using the official apt repository instead:

```bash
# Remove snap docker
sudo snap remove docker

# Install from official repository
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

Log out and back in for group changes to take effect.

## Web UI Issues

### Cannot Access Web UI

**Check container is running:**

```bash
docker ps | grep mavt
```

**Check port binding:**

```bash
docker port mavt
# Should show: 8080/tcp -> 0.0.0.0:7738
```

**Test from inside container:**

```bash
docker exec mavt wget -O- http://localhost:8080/api/health
```

If this works but you can't access from host:
- Check firewall: `sudo ufw status`
- Allow port: `sudo ufw allow 7738`

### Web UI Loads but Can't Search Apps

**Symptoms:**
- Search returns no results or errors
- API calls fail

**Check logs:**

```bash
docker-compose logs -f | grep -i error
```

**Common causes:**
- App Store API rate limiting (wait a few minutes)
- Network connectivity issues
- Incorrect `MAVT_COUNTRY` code

## Data Issues

### Version Updates Not Being Detected

**Check configuration:**

```bash
docker exec mavt env | grep MAVT
```

Ensure:
- `MAVT_CHECK_INTERVAL` is reasonable (e.g., `4h`, not too short)
- `MAVT_APPS` contains valid bundle IDs

**Force a check:**

```bash
docker exec mavt /app/mavt -check
```

### Lost Data After Restart

**Symptoms:**
- All tracked apps disappear after container restart
- No version history

**Cause:** Not using persistent volume

**Solution:**
Ensure using named volume in docker-compose.yml:

```yaml
volumes:
  mavt-data:/app/data  # External named volume

volumes:
  mavt-data:
    external: true
```

Create volume before starting:
```bash
docker volume create mavt-data
```

## Debugging Commands

### Check Health Status

```bash
# View health check logs
docker inspect mavt | grep -A 20 Health

# Manually run health check command
docker exec mavt sh -c "test -d /app/data && test -w /app/data && echo OK || echo FAIL"
```

### Check File Permissions

```bash
# List data directory
docker exec mavt ls -la /app/data

# Check ownership
docker exec mavt stat -c '%U:%G' /app/data

# Should show: mavt:mavt or 1000:1000
```

### Check Process Running

```bash
# Check if process is running
docker exec mavt ps aux

# Should see: /app/mavt -daemon
```

### Test API Endpoints

```bash
# Health check
curl http://localhost:7738/api/health

# List apps
curl http://localhost:7738/api/apps

# Search
curl "http://localhost:7738/api/search?q=instagram&limit=1"
```

## Performance Issues

### High CPU Usage

**Cause:** Check interval too short

**Solution:** Increase `MAVT_CHECK_INTERVAL`:

```yaml
environment:
  - MAVT_CHECK_INTERVAL=12h  # or 24h
```

### Container Taking Long to Start

**Normal behavior:**
- First start: Creates directories, initializes storage
- With many apps: Initial check of all apps can take time

**Check progress:**
```bash
docker-compose logs -f
```

## Getting Help

If you've tried the above solutions and still have issues:

1. Collect diagnostic information:
   ```bash
   # Container status
   docker ps -a | grep mavt

   # Logs
   docker logs mavt --tail 100 > mavt-logs.txt

   # Configuration
   docker inspect mavt > mavt-inspect.json

   # Volume info
   docker volume inspect mavt-data
   ```

2. Check [GitHub Issues](https://github.com/thomas/mavt/issues)

3. Create a new issue with:
   - Your OS and version
   - Docker version: `docker --version`
   - Docker Compose version: `docker-compose --version`
   - Diagnostic information from step 1
   - Description of the problem
