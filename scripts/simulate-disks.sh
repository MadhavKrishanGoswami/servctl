#!/bin/bash
# simulate-disks.sh - Creates virtual disks using loop devices
# Run with: sudo ./simulate-disks.sh

set -e

echo "=== Creating Simulated Disks ==="
echo ""

# Create disk image files
mkdir -p /tmp/disks

# Simulate SSD (500GB shown as small file)
echo "[1/4] Creating simulated SSD (ssd1.img - 100MB)..."
dd if=/dev/zero of=/tmp/disks/ssd1.img bs=1M count=100 2>/dev/null
LOOP_SSD1=$(losetup -f --show /tmp/disks/ssd1.img)
echo "  Created: $LOOP_SSD1 (simulating 500GB SSD)"

# Simulate HDD 1 (2TB shown as small file)
echo "[2/4] Creating simulated HDD 1 (hdd1.img - 100MB)..."
dd if=/dev/zero of=/tmp/disks/hdd1.img bs=1M count=100 2>/dev/null
LOOP_HDD1=$(losetup -f --show /tmp/disks/hdd1.img)
echo "  Created: $LOOP_HDD1 (simulating 2TB HDD)"

# Simulate HDD 2 (2TB backup)
echo "[3/4] Creating simulated HDD 2 (hdd2.img - 100MB)..."
dd if=/dev/zero of=/tmp/disks/hdd2.img bs=1M count=100 2>/dev/null
LOOP_HDD2=$(losetup -f --show /tmp/disks/hdd2.img)
echo "  Created: $LOOP_HDD2 (simulating 2TB HDD backup)"

# Simulate NVMe (fast cache)
echo "[4/4] Creating simulated NVMe (nvme1.img - 50MB)..."
dd if=/dev/zero of=/tmp/disks/nvme1.img bs=1M count=50 2>/dev/null
LOOP_NVME=$(losetup -f --show /tmp/disks/nvme1.img)
echo "  Created: $LOOP_NVME (simulating 256GB NVMe)"

echo ""
echo "=== Simulated Disk Summary ==="
echo ""
losetup -l
echo ""
echo "=== lsblk output ==="
lsblk
echo ""
echo "Disks ready for testing!"
echo ""
echo "To cleanup: sudo ./simulate-disks.sh cleanup"

# Store loop devices for cleanup
echo "$LOOP_SSD1" > /tmp/disks/loops.txt
echo "$LOOP_HDD1" >> /tmp/disks/loops.txt
echo "$LOOP_HDD2" >> /tmp/disks/loops.txt
echo "$LOOP_NVME" >> /tmp/disks/loops.txt
