#!/bin/sh
set -e

# Ensure data directory structure exists with correct permissions
if [ -w "/app/data" ]; then
    # Directory is writable, create subdirectories if they don't exist
    mkdir -p /app/data/apps /app/data/updates
    echo "Data directory structure initialized"
else
    echo "ERROR: /app/data is not writable by current user ($(id -u):$(id -g))"
    echo "Please ensure the volume mounted at /app/data has correct permissions"
    exit 1
fi

# Execute the main application
exec "$@"
