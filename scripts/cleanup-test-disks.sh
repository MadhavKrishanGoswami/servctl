#!/bin/bash
# Cleanup test disk images and loop devices
# Run with: sudo ./scripts/cleanup-test-disks.sh

set -e

DISK_DIR="/test-disks"

echo "=== Cleaning up test disks ==="

# Remove loop devices
echo "Removing loop devices..."
for loop in $(losetup -l | grep test-disks | awk '{print $1}'); do
    echo "  Detaching $loop"
    losetup -d "$loop" || true
done

# Remove disk images
if [ -d "$DISK_DIR" ]; then
    echo "Removing disk images..."
    rm -rf "$DISK_DIR"
fi

echo "=== Cleanup complete ==="
