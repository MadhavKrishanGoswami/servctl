#!/bin/bash
# simulate-disks.sh - Create virtual block devices for testing servctl
# Usage: ./simulate-disks.sh [number_of_disks] [size_mb]

set -e

NUM_DISKS=${1:-2}
SIZE_MB=${2:-512}

echo "🔧 Creating $NUM_DISKS virtual disks (${SIZE_MB}MB each)..."

# Ensure loop module is loaded
modprobe loop 2>/dev/null || true

for i in $(seq 1 "$NUM_DISKS"); do
    DISK_FILE="/tmp/vdisk${i}.img"
    LOOP_DEV="/dev/loop${i}"
    
    # Skip if already exists
    if losetup "$LOOP_DEV" 2>/dev/null; then
        echo "  ℹ️  $LOOP_DEV already exists, skipping"
        continue
    fi
    
    # Create disk image file
    echo "  📁 Creating $DISK_FILE (${SIZE_MB}MB)..."
    dd if=/dev/zero of="$DISK_FILE" bs=1M count="$SIZE_MB" status=none
    
    # Create loop device
    echo "  🔗 Attaching to $LOOP_DEV..."
    losetup "$LOOP_DEV" "$DISK_FILE"
    
    # Make it look like a real disk to lsblk
    # Add a partition table
    echo "  📋 Creating partition table..."
    parted -s "$LOOP_DEV" mklabel gpt
    
    echo "  ✅ Virtual disk $i ready: $LOOP_DEV"
done

echo ""
echo "📊 Virtual disks created:"
lsblk -o NAME,SIZE,TYPE,MOUNTPOINT | grep -E "^loop[0-9]" || echo "  (use 'lsblk' to verify)"

echo ""
echo "🧪 Ready for servctl testing!"
echo "   Run: ./servctl -dry-run -start-setup"
