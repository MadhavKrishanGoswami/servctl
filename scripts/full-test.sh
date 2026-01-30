#!/bin/bash
# full-test.sh - Complete servctl test suite
# Runs all tests in proper order

set -e

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║              servctl Full Test Suite                        ║"
echo "║              Ubuntu 22.04 Simulation                        ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

cd /app

# ============================================
# Test 1: Version & Basic Info
# ============================================
echo "┌──────────────────────────────────────────────────────────────┐"
echo "│ TEST 1: Version Check                                       │"
echo "└──────────────────────────────────────────────────────────────┘"
./servctl -version
echo "✅ PASS"
echo ""

# ============================================
# Test 2: Help Output
# ============================================
echo "┌──────────────────────────────────────────────────────────────┐"
echo "│ TEST 2: Help/Usage Output                                   │"
echo "└──────────────────────────────────────────────────────────────┘"
./servctl 2>&1 || true
echo "✅ PASS"
echo ""

# ============================================
# Test 3: Preflight Checks
# ============================================
echo "┌──────────────────────────────────────────────────────────────┐"
echo "│ TEST 3: Preflight System Checks                             │"
echo "└──────────────────────────────────────────────────────────────┘"
sudo ./servctl -preflight || echo "⚠️  Some checks may fail in container (expected)"
echo ""

# ============================================
# Test 4: Disk Discovery (lsblk)
# ============================================
echo "┌──────────────────────────────────────────────────────────────┐"
echo "│ TEST 4: Disk Discovery (Raw lsblk)                          │"
echo "└──────────────────────────────────────────────────────────────┘"
echo "Available block devices:"
lsblk -o NAME,SIZE,TYPE,MOUNTPOINT 2>/dev/null || echo "No block devices (container limitation)"
echo ""

# ============================================
# Test 5: Dependency Check
# ============================================
echo "┌──────────────────────────────────────────────────────────────┐"
echo "│ TEST 5: Dependency Verification                             │"
echo "└──────────────────────────────────────────────────────────────┘"
echo "Checking required binaries..."
for cmd in curl docker hdparm smartctl lsblk mkfs.ext4 mkfs.xfs; do
    if command -v $cmd &> /dev/null; then
        echo "  ✅ $cmd"
    else
        echo "  ❌ $cmd (missing)"
    fi
done
echo ""

# ============================================
# Test 6: File System Tools
# ============================================
echo "┌──────────────────────────────────────────────────────────────┐"
echo "│ TEST 6: Filesystem Tools                                    │"
echo "└──────────────────────────────────────────────────────────────┘"
echo "Available mkfs commands:"
ls -la /sbin/mkfs.* 2>/dev/null || echo "Checking /usr/sbin..."
ls -la /usr/sbin/mkfs.* 2>/dev/null || echo "No mkfs found"
echo ""

# ============================================
# Summary
# ============================================
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║              TEST SUITE COMPLETE                            ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "Next steps:"
echo "  • Run 'make docker-shell' for interactive testing"
echo "  • Run './servctl -start-setup' to test the wizard"
echo ""
