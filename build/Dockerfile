# Full test container with simulated drives
# Simulates Ubuntu 24.04 with multiple disk types and Go toolchain

FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

# Install all dependencies including Go
RUN apt-get update && apt-get install -y \
    # Go toolchain
    golang-go \
    git \
    curl \
    ca-certificates \
    # Storage utilities
    util-linux \
    hdparm \
    smartmontools \
    e2fsprogs \
    xfsprogs \
    btrfs-progs \
    dosfstools \
    parted \
    gdisk \
    lvm2 \
    # RAID/pooling
    mdadm \
    mergerfs \
    # System utilities
    cron \
    ufw \
    sudo \
    iputils-ping \
    dnsutils \
    kmod \
    && rm -rf /var/lib/apt/lists/*

# Install ZFS if available
RUN apt-get update && apt-get install -y zfsutils-linux || true \
    && rm -rf /var/lib/apt/lists/*

# Install Docker CLI + Compose
RUN curl -fsSL https://get.docker.com | sh || true && \
    apt-get update && apt-get install -y docker-compose-plugin || true \
    && rm -rf /var/lib/apt/lists/*

# Create test user with sudo
RUN useradd -m -s /bin/bash testuser && \
    echo "testuser ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Create test disk images directory
RUN mkdir -p /test-disks

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download || true

# Copy source code
COPY . .

# Build the binary
RUN go build -o /app/bin/servctl ./cmd/servctl || true

# Copy scripts
RUN chmod +x /app/scripts/*.sh 2>/dev/null || true

# Default: run all tests including integration
CMD ["go", "test", "./...", "-v", "-tags=integration", "-count=1"]
