#!/bin/bash
# Setup virtual disk images for integration testing
# Run with: sudo ./scripts/setup-test-disks.sh

set -e

DISK_DIR="/test-disks"

echo "=== Setting up test disk images ==="

# Create directory
mkdir -p "$DISK_DIR"

# Create disk images (simulating different sizes)
echo "Creating disk images..."
dd if=/dev/zero of="$DISK_DIR/disk1.img" bs=1M count=100 status=progress
dd if=/dev/zero of="$DISK_DIR/disk2.img" bs=1M count=100 status=progress
dd if=/dev/zero of="$DISK_DIR/disk3.img" bs=1M count=50 status=progress
dd if=/dev/zero of="$DISK_DIR/disk4.img" bs=1M count=200 status=progress

# Setup loop devices
echo "Setting up loop devices..."
losetup -f "$DISK_DIR/disk1.img"
losetup -f "$DISK_DIR/disk2.img"
losetup -f "$DISK_DIR/disk3.img"
losetup -f "$DISK_DIR/disk4.img"

# Show created devices
echo ""
echo "=== Test devices created ==="
losetup -l | grep test-disks
echo ""
echo "Disk layout:"
lsblk -o NAME,SIZE,TYPE,FSTYPE,MOUNTPOINT

echo ""
echo "=== Ready for testing ==="
echo "Run tests with: go test ./... -v -tags=integration"
