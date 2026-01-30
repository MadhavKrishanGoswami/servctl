#!/bin/bash
# start-with-docker.sh - Starts Docker daemon then runs tests
# Used in Docker-in-Docker container

set -e

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║         servctl Docker-in-Docker Test Environment           ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

echo "Starting Docker daemon..."
dockerd --storage-driver=vfs &>/var/log/dockerd.log &

# Wait for Docker to be ready
echo "Waiting for Docker daemon..."
for i in {1..30}; do
    if docker info &>/dev/null; then
        echo "✅ Docker daemon is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "⚠️  Docker daemon failed to start (this is OK for basic tests)"
        cat /var/log/dockerd.log | tail -10
    fi
    sleep 1
done

echo ""
echo "Docker status:"
docker --version || echo "Docker CLI available"
docker compose version 2>/dev/null || echo "Docker Compose: checking..."

echo ""
echo "────────────────────────────────────────────────────────────"
echo ""

# Run the full test suite
/app/full-test.sh

# Keep container running if interactive
if [ -t 0 ]; then
    echo ""
    echo "════════════════════════════════════════════════════════════"
    echo "Interactive shell ready. Try:"
    echo "  ./servctl -preflight"
    echo "  ./servctl -start-setup"
    echo "  docker ps"
    echo "════════════════════════════════════════════════════════════"
    echo ""
    cd /app
    exec /bin/bash
fi
