#!/bin/bash
# start-with-docker.sh - Start Docker daemon and prepare environment
# This is the entrypoint for the DinD container

set -e

echo "🐳 Starting Docker daemon..."

# Start Docker daemon in background
dockerd --storage-driver=vfs &

# Wait for Docker to be ready (up to 60 seconds)
echo "⏳ Waiting for Docker daemon to start..."
WAIT_COUNT=0
MAX_WAIT=60
while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if docker info >/dev/null 2>&1; then
        echo "✅ Docker is ready!"
        break
    fi
    WAIT_COUNT=$((WAIT_COUNT + 1))
    if [ $((WAIT_COUNT % 10)) -eq 0 ]; then
        echo "   Still waiting... ($WAIT_COUNT/${MAX_WAIT}s)"
    fi
    sleep 1
done

if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker failed to start within ${MAX_WAIT} seconds"
    exit 1
fi

# Create virtual disks for testing
echo ""
echo "💾 Setting up virtual disks..."
if [ -f /app/simulate-disks.sh ]; then
    /app/simulate-disks.sh 2 256 || echo "⚠️ Could not create virtual disks (may need --privileged)"
fi

echo ""
echo "════════════════════════════════════════════════════"
echo "  🚀 servctl Test Environment Ready!"
echo "════════════════════════════════════════════════════"
echo ""
echo "  Available commands:"
echo "    ./servctl -version"
echo "    ./servctl -preflight"
echo "    ./servctl -dry-run -start-setup"
echo "    ./servctl -start-setup"
echo ""
echo "  Virtual disks: /dev/loop1, /dev/loop2"
echo ""

# Keep container running or execute command
if [ $# -gt 0 ]; then
    exec "$@"
else
    exec /bin/bash
fi
