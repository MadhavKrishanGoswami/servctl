# Sample Docker Compose File for the infra
```yaml
name: family-cloud

services:
  # --- IMMICH CORE ---
  immich-server:
    container_name: immich_server
    image: ghcr.io/immich-app/immich-server:${IMMICH_VERSION:-release}
    volumes:
      - ${UPLOAD_LOCATION}:/usr/src/app/upload
      - /etc/localtime:/etc/localtime:ro
    env_file: .env
    ports:
      - '2283:2283'
    depends_on: [redis, database]
    restart: always

  immich-machine-learning:
    container_name: immich_machine_learning
    image: ghcr.io/immich-app/immich-machine-learning:${IMMICH_VERSION:-release}
    volumes:
      - model-cache:/cache
    env_file: .env
    restart: always

  # --- NEXTCLOUD ---
  nextcloud:
    image: nextcloud:${NEXTCLOUD_TAG:-apache}
    container_name: nextcloud
    ports:
      - '8080:80'
    depends_on:
      - database
    environment:
      # Database Connection
      - POSTGRES_HOST=${POSTGRES_HOST}
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      
      # Admin Auto-Setup
      - NEXTCLOUD_ADMIN_USER=${NEXTCLOUD_ADMIN_USER}
      - NEXTCLOUD_ADMIN_PASSWORD=${NEXTCLOUD_ADMIN_PASSWORD}
      - NEXTCLOUD_DATA_DIR=/var/www/html/data
      
      # NETWORK FIXES (Crucial for "Black Screen")
      - NEXTCLOUD_TRUSTED_DOMAINS=192.168.1.24
      - OVERWRITEHOST=192.168.1.24:8080
      - OVERWRITEPROTOCOL=http
    volumes:
      - ${NEXT_DATA}:/var/www/html/data
      - nextcloud_config:/var/www/html
    restart: always
  # --- INFRASTRUCTURE (On SSD) ---
  redis:
    container_name: immich_redis
    image: docker.io/valkey/valkey:9
    restart: always

  database:
    container_name: immich_postgres
    image: ghcr.io/immich-app/postgres:14-vectorchord0.4.3-pgvectors0.2.0
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_USER: ${DB_USERNAME}
      POSTGRES_DB: ${DB_DATABASE_NAME}
      POSTGRES_INITDB_ARGS: '--data-checksums'
    volumes:
      - ${DB_DATA_LOCATION}:/var/lib/postgresql/data
    shm_size: 128mb
    restart: always
  # --- MONITORING ---
  glances:
    container_name: glances
    image: nicolargo/glances:latest-full  # "Full" includes tools to read HDD temps
    restart: always
    pid: host              # Allows seeing "Real" Host CPU usage
    network_mode: host     # Uses port 61208 directly
    cap_add:
      - SYS_ADMIN          # Required to read HDD SMART data
      - SYS_RAWIO          # Required to read HDD SMART data
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro  # Monitor container stats
      - /:/rootfs:ro                                  # Monitor Host Disk usage
      - /run/udev:/run/udev:ro                        # Hardware device mapping
    environment:
      - TZ=Asia/Kolkata         # Set your timezone
      - GLANCES_OPT=-w          # Run in Web Server mode
  # ---- UPDATE NOTIFIER (NEW) ---
  diun:
    image: crazymax/diun:latest
    container_name: diun
    command: serve
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock"
      - "diun_data:/data"
    environment:
      - "TZ=Asia/Kolkata"
      - "LOG_LEVEL=info"
      - "LOG_JSON=false"
      # Check every 6 hours
      - "DIUN_WATCH_SCHEDULE=0 8 1,15 * *"
      # Watch all containers by default? (True = Easy Mode)
      - "DIUN_PROVIDERS_DOCKER_WATCHBYDEFAULT=true"
      # Discord Webhook (PASTE YOUR URL HERE or in .env)
      - "DIUN_NOTIF_DISCORD_WEBHOOKURL=https://discord.com/api/webhooks/1464614181431935311/kJh0A510vDNaTNMNmpydzKQ217gVnShzayCka6kPxLHOpsoEp9jnZHlZXiCkBHHn55ac"
      # Render fancy cards
      - "DIUN_NOTIF_DISCORD_RENDERFIELDS=true"
    restart: always
volumes:
  model-cache:
  nextcloud_config:
  scrutiny_config:
  scrutiny_influxdb:
  diun_data:
``` 

# Sample .env File
```env
# ==========================================
# 1. SHARED DATABASE (Postgres)
# ==========================================
# This connects Immich & Nextcloud to the SAME server
DB_USERNAME=postgres
DB_PASSWORD=madhav8224
DB_DATABASE_NAME=immich
DB_DATA_LOCATION=/var/lib/docker/volumes/pgdata

# ==========================================
# 2. IMMICH CONFIG
# ==========================================
IMMICH_VERSION=release
UPLOAD_LOCATION=/mnt/data/gallery

# ==========================================
# 3. NEXTCLOUD CONFIG
# ==========================================
NEXTCLOUD_TAG=apache
NEXT_DATA=/mnt/data/cloud/data

# --- AUTO CONFIGURATION ---
# These tell Nextcloud how to talk to the DB we just created in Step 1
POSTGRES_HOST=database
POSTGRES_DB=nextcloud
POSTGRES_USER=postgres
POSTGRES_PASSWORD=madhav8224

# --- ADMIN USER ---
NEXTCLOUD_ADMIN_USER=madhav
NEXTCLOUD_ADMIN_PASSWORD=madhav8224
```
